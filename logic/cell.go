/**
 * @Author: Hardews
 * @Date: 2023/5/25 21:48
 * @Description:
**/

package logic

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type cellJson struct {
	Code int `json:"code"`
	Data struct {
		Blocks []struct {
			BlockToken    string `json:"block_token"`
			GzipDatatable string `json:"gzip_datatable"`
		} `json:"blocks"`
	} `json:"data"`
}

type cellRes struct {
	Rows []struct {
		Columns []struct {
			Value   interface{} `json:"value"`
			StyleId int         `json:"styleId"`
		} `json:"columns"`
	} `json:"rows"`
}

func delWithCell(sheetName string, data cellJson) {
	var cellR cellRes
	base64Data := data.Data.Blocks[0].GzipDatatable
	// 对Base64字符串进行解码
	compressedBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		log.Println("base64 decode error:", err)
		return
	}

	// 创建一个bytes.Buffer并将解码后的数据写入其中
	buf := bytes.NewBuffer(compressedBytes)

	// 创建一个gzip.Reader来读取压缩数据
	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		log.Println("gzip.NewReader error:", err)
		return
	}
	defer gzipReader.Close()

	// 读取解压缩后的数据
	decompressedData, err := io.ReadAll(gzipReader)
	if err != nil {
		log.Println("ioutil.ReadAll error:", err)
		return
	}

	err = json.Unmarshal(decompressedData, &cellR)
	time.Sleep(2 * time.Second)
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}()

	// 创建一个工作表
	index, err := f.NewSheet("sheet1")
	if err != nil {
		log.Println(err)
		return
	}
	// 设置工作簿的默认工作表
	f.SetActiveSheet(index)

	var (
		nowRow int = 1
	)

	for _, row := range cellR.Rows {
		// 遍历这行的值
		var rowArr []interface{}
		for _, column := range row.Columns {
			rowArr = append(rowArr, column.Value)
		}
		// 按行赋值
		err := f.SetSheetRow("sheet1", "A"+strconv.Itoa(nowRow), &rowArr)
		if err != nil {
			log.Println("set sheet row failed,err:", err)
		}
		nowRow++
	}

	// 根据指定路径保存文件
	if err := f.SaveAs("./result/excel/" + sheetName + ".xlsx"); err != nil {
		log.Println(err)
	}

	fmt.Println("表格：", sheetName)
}

type SheetGetMap struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Blocks []struct {
		Name  string `json:"name"`
		Token string `json:"token"`
	} `json:"blocks"`
}

type SheetGetJson struct {
	Sheets map[string]SheetGetMap `json:"sheets"`
}

type GzipSnapshot struct {
	GzipSnapshot string `json:"gzip_snapshot"`
}

func GetSheetIdAndBlockToken(url string, objToken ...string) {
	res := sendRequest(url)

	re := regexp.MustCompile(`"gzip_snapshot":"(.+?)\"`)
	match := re.FindStringSubmatch(string(res))
	var gzipSnapshot string
	if len(match) > 0 {
		gzipSnapshot = match[1]
	} else {
		fmt.Println("不是表格", url)
		return
	}
	if len(gzipSnapshot) == 0 {
		return
	}

	gzipSnapshot, err := strconv.Unquote(`"` + gzipSnapshot + `"`)
	if err != nil {
		log.Println(err)
		log.Println(gzipSnapshot)
		return
	}

	gzipSnapshotBytes, err := base64.StdEncoding.DecodeString(gzipSnapshot)
	if err != nil {
		log.Println("base64 decode error:", err)
		return
	}

	// 创建一个bytes.Buffer并将解码后的数据写入其中
	buf := bytes.NewBuffer(gzipSnapshotBytes)

	// 创建一个gzip.Reader来读取压缩数据
	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		log.Println("gzip.NewReader error:", err)
		return
	}
	defer gzipReader.Close()

	// 读取解压缩后的数据
	decompressedData, err := io.ReadAll(gzipReader)
	if err != nil {
		log.Println("ioutil.ReadAll error:", err)
		return
	}

	time.Sleep(2 * time.Second)

	var cellR SheetGetJson
	err = json.Unmarshal(decompressedData, &cellR)

	var token string
	if len(objToken) != 0 {
		token = objToken[0]
		if token == "" {
			return
		}
	}
	GetUrl := strings.ReplaceAll(cellUrl, "{.token}", token)
	for _, getMap := range cellR.Sheets {
		if len(getMap.Blocks) > 0 {
			for _, block := range getMap.Blocks {
				// 获取 block_token
				urlStr := strings.ReplaceAll(GetUrl, "{.block_token}", block.Token)
				sheetRes := sendRequest(urlStr)
				var cj cellJson
				json.Unmarshal(sheetRes, &cj)
				if cj.Code != 0 {
					return
				}

				var name string
				if len(objToken) > 1 {
					name += objToken[1] + "-"
				}
				name += getMap.Name
				go func() {
					delWithCell(name, cj)
				}()
			}
		}
	}
}
