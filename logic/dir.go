/**
 * @Author: Hardews
 * @Date: 2023/5/26 19:23
 * @Description:
**/

package logic

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

type dirData struct {
	Code int `json:"code"`
	Data struct {
		Entities struct {
			Nodes map[string]docInfo `json:"nodes"`
		} `json:"entities"`
		NodeList []string `json:"node_list"`
	} `json:"data"`
	Msg string `json:"msg"`
}

var (
	DirMap   = sync.Map{}
	DocMap   = sync.Map{}
	SheetMap = sync.Map{}
)

type docInfo struct {
	ObjToken string `json:"obj_token"`
	Url      string `json:"url"`
}

func GetAllFolderDir() {
	if len(urls.DirToken) == 0 {
		return
	}
	// 获取到所有的地址
	for _, token := range urls.DirToken {
		// 获取该地址的子目录
		str := FolderListUrl + token
		resByte := sendRequest(str)
		var resJson dirData
		json.Unmarshal(resByte, &resJson)
		// 处理得到的子目录
		GetDir(resJson)
	}
}

func GetDir(resJ dirData) {
	// 遍历得到的子目录节点
	for _, info := range resJ.Data.Entities.Nodes {
		// 如果子目录是文件夹
		if strings.Contains(info.Url, "/drive/folder") {
			// 先判断是否扫描过这个子文件夹
			if _, ok := DirMap.Load(info.Url); !ok {
				// 没有则扫描一遍
				go func() {
					// 分解 合并 得到 url
					tokenArr := strings.Split(info.Url, "/")
					resByte := sendRequest(FolderListUrl + tokenArr[len(tokenArr)-1])
					var resJson dirData
					json.Unmarshal(resByte, &resJson)

					go func() {
						GetDir(resJson)
					}()
					// 这是为了防止 goroutine 泄露
					time.Sleep(3 * time.Second)
				}()
				DirMap.Store(info.Url, struct{}{})
			}
		}
		// 文档
		if strings.Contains(info.Url, "wiki") || strings.Contains(info.Url, "docx") {
			if _, ok := DocMap.Load(info.Url); !ok {
				DocMap.Store(info.Url, struct{}{})

				tokenArr := strings.Split(info.Url, "/")
				token := tokenArr[len(tokenArr)-1]

				if strings.Contains(token, "doxcn") {
					token = FindDocToken(info.Url)
				}

				go func() {
					docResByte := sendRequest(strings.ReplaceAll(docUrl, "{.id}", token))
					var res docJson
					json.Unmarshal(docResByte, &res)
					DealWithDoc(res)
				}()
			}
		}
		// 表格
		if strings.Contains(info.Url, "sheet") {
			if _, ok := SheetMap.Load(info.Url); !ok {
				SheetMap.Store(info.Url, struct{}{})
			}

			go func() {
				GetSheetIdAndBlockToken(info.Url)
			}()
		}
	}
}

func FindDocToken(docxUrl string) string {
	var token string
	res := sendRequest(docxUrl)

	re := regexp.MustCompile(`"token":"([a-zA-Z0-9]{27})"`)
	match := re.FindStringSubmatch(string(res))
	if len(match) > 1 {
		token = match[1]
	} else {
		fmt.Println("未找到匹配的 token")
		return ""
	}

	return strings.TrimSpace(token)
}
