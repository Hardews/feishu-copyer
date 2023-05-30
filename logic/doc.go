/**
 * @Author: Hardews
 * @Date: 2023/5/25 21:48
 * @Description:文档处理
**/

package logic

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

var columnMap = map[string]string{
	"heading1": "# ",
	"heading2": "## ",
	"heading3": "### ",
	"heading4": "#### ",
	"heading5": "##### ",
	"heading6": "###### ",
	"text":     "",
	"bullet":   "- ",
}

// 文档中每一行的存储形式
type line struct {
	Data struct {
		ColumnsId []string `json:"columns_id"`
		RowsId    []string `json:"rows_id"`

		File struct {
			Name  string `json:"name"`
			Token string `json:"token"`
		} `json:"file"`

		Image struct {
			Token string `json:"token"`
		} `json:"image"`

		Seq      string   `json:"seq"`
		Children []string `json:"children"`
		Done     bool     `json:"done"`
		Text     struct {
			Apool struct {
				NumToAttrib struct {
					Field1 []string `json:"0"`
					Field2 []string `json:"1"`
				} `json:"numToAttrib"`
			} `json:"apool"`
			InitialAttributedTexts struct {
				Attribs struct {
					Field1 string `json:"0"`
				} `json:"attribs"`
				Text struct {
					Field1 string `json:"0"`
				} `json:"text"`
			} `json:"initialAttributedTexts"`
		} `json:"text"`
		Type string `json:"type"`
	} `json:"data"`
	Id string `json:"id"`
}

type meta struct {
	Title string `json:"title"`
}

type docJson struct {
	Code int `json:"code"`
	Data struct {
		BlockMap           map[string]line `json:"block_map"`
		BlockSequence      []string        `json:"block_sequence"`
		ExternalMentionUrl interface{}     `json:"external_mention_url"`
		HasMore            bool            `json:"has_more"`
		Id                 string          `json:"id"`
		MentionPageTitle   struct {
		} `json:"mention_page_title"`
		MetaMap map[string]meta `json:"meta_map"`
		Type    string          `json:"type"`
	} `json:"data"`
	Msg string `json:"msg"`
}

// DealWithDoc 处理飞书文档的函数
func DealWithDoc(res docJson) {
	var (
		isTableNow  bool
		isNextOrder bool
		isNextCell  bool
		row         int
		col         int
		nowRow      int
		nowCol      int

		orderNum int = -1
	)
	title := res.Data.MetaMap[res.Data.Id].Title

	var resultFile, err = os.Create("./result/docx/" + title + ".md")
	if err != nil {
		log.Printf("create result file failed,err:%s,filename:%s", err, title)
		return
	}

	fmt.Println("文档文件名:", title)

	defer resultFile.Close()

	resultFile.Write([]byte("# " + title + "\n"))
	for _, s := range res.Data.BlockSequence {
		if res.Data.BlockMap[s].Data.Type == "view" || res.Data.BlockMap[s].Data.Type == "page" || res.Data.BlockMap[s].Data.Type == "sheet" {
			continue
		}
		if res.Data.BlockMap[s].Data.Type == "table" {
			// 表格有几行
			row = len(res.Data.BlockMap[s].Data.RowsId)
			// 表格有几列
			col = len(res.Data.BlockMap[s].Data.ColumnsId)
			isTableNow = true

			// 现在在表格的第几行第几列
			nowRow = 1
			nowCol = 1
			continue
		}
		if res.Data.BlockMap[s].Data.Type == "table_cell" {
			// 下一个是单元格的值
			isNextCell = true
			continue
		}
		if isNextCell && isTableNow {
			// 表头分割
			if nowRow == 2 && nowCol == 1 {
				for i := 0; i < col; i++ {
					resultFile.Write([]byte("|-----"))
				}
				resultFile.Write([]byte("|\n"))
			}
			// 如果现在是单元格的值
			bm := res.Data.BlockMap[s]
			resultFile.Write([]byte("|" + bm.Data.Text.InitialAttributedTexts.Text.Field1))
			if nowCol == col && nowRow == row {
				// 表格结束
				isTableNow = false
				resultFile.Write([]byte("|\n"))
			} else if nowCol == col {
				// 这一列结束
				resultFile.Write([]byte("|\n"))
				nowCol = 0
				nowRow++
			}

			nowCol++
			continue
		}

		bm := res.Data.BlockMap[s]
		if bm.Data.Type == "code" {
			resultFile.Write([]byte("```\n"))
			resultFile.Write([]byte(bm.Data.Text.InitialAttributedTexts.Text.Field1))
			resultFile.Write([]byte("\n```\n"))
			continue
		}

		if bm.Data.Type == "todo" {
			var str = "- ["
			if bm.Data.Done {
				str += "x"
			} else {
				str += " "
			}
			str += "] "
			str += bm.Data.Text.InitialAttributedTexts.Text.Field1 + "\n"
			resultFile.Write([]byte(str))
			continue
		}

		if bm.Data.Type == "image" {
			fileUrl := "https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/preview/"
			token := bm.Data.Image.Token
			fileUrl += token
			// 统一存为 png
			fileUrl += "?mount_point=image_png&preview_type=16"

			resultFile.Write([]byte(fmt.Sprintf("![Image](%s)\n", fileUrl)))
			continue
		}

		if bm.Data.Type == "file" {
			// 以附件的形式存储
			var attachmentFile, err = os.Create("./result/attachment/" + title + "的附件：" + bm.Data.File.Name)
			if err != nil {
				log.Printf("create attachment file failed,err:%s,filename:%s", err, title)
				return
			}
			fmt.Println("附件文件名:", title)

			fileUrl := "https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/preview/"
			token := bm.Data.File.Token
			fileUrl += token
			fileUrl += "?mount_point=docx_file&preview_type=16"

			attachmentByte := sendRequest(fileUrl)
			// 文件越大，获取需要的时间越久
			time.Sleep(5 * time.Second)
			attachmentFile.Write(attachmentByte)
			attachmentFile.Close()
			continue
		}

		if len(bm.Data.Text.Apool.NumToAttrib.Field1) != 0 && bm.Data.Text.Apool.NumToAttrib.Field1[0] == "link" {
			resultFile.Write([]byte("[" + bm.Data.Text.InitialAttributedTexts.Text.Field1 + "](" + bm.Data.Text.Apool.NumToAttrib.Field1[1] + ")"))
			continue
		}

		if bm.Data.Type == "ordered" {
			if bm.Data.Seq == "1" {
				resultFile.Write([]byte("\n"))
				orderNum = 1
				resultFile.Write([]byte(strconv.Itoa(orderNum) + ". " + bm.Data.Text.InitialAttributedTexts.Text.Field1 + "\n"))
				orderNum++
				isNextOrder = true
				continue
			}
			if isNextOrder && bm.Data.Seq == "auto" {
				resultFile.Write([]byte(strconv.Itoa(orderNum) + ". " + bm.Data.Text.InitialAttributedTexts.Text.Field1 + "\n"))
				orderNum++
				isNextOrder = true
			}
			continue
		}

		if bm.Data.Type == "quote_container" {
			if len(bm.Data.Children[0]) != 0 {
				str := "> "
				str += res.Data.BlockMap[bm.Data.Children[0]].Data.Text.InitialAttributedTexts.Text.Field1
				line1 := res.Data.BlockMap[bm.Data.Children[0]]
				line1.Data.Text.InitialAttributedTexts.Text.Field1 = str
				res.Data.BlockMap[bm.Data.Children[0]] = line1
			}
			continue
		}

		if str, ok := columnMap[bm.Data.Type]; ok {
			resultFile.Write([]byte(str + bm.Data.Text.InitialAttributedTexts.Text.Field1 + "\n"))
		}
	}
}
