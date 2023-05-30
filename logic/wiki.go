/**
 * @Author: Hardews
 * @Date: 2023/5/28 1:17
 * @Description:
**/

package logic

import (
	"encoding/json"
	"strings"
	"sync"
	"time"
)

var (
	WiNode    = sync.Map{}
	WiDocNode = sync.Map{}
)

type WikiNode struct {
	WikiToken string `json:"wiki_token"`
	ObjToken  string `json:"obj_token"`
	Title     string `json:"title"`
}

type WikiJson struct {
	Code int `json:"code"`
	Data struct {
		Tree struct {
			Nodes map[string]WikiNode `json:"nodes"`
		} `json:"tree"`
	} `json:"data"`
}

func GetAllWikiDir() {
	if len(urls.WikiToken) == 0 {
		return
	}
	for _, wikiToken := range urls.WikiToken {
		if _, ok := WiNode.Load(wikiToken); !ok {
			res := sendRequest(WikiListUrl + wikiToken)
			var wikiList WikiJson
			json.Unmarshal(res, &wikiList)

			GetWikiDir(wikiList)
			WiNode.Store(wikiToken, struct{}{})
		}
	}
}

func GetWikiDir(wikiList WikiJson) {
	for _, node := range wikiList.Data.Tree.Nodes {
		// 是否查询过可不可以是文档
		if _, ok := WiDocNode.Load(node.WikiToken); !ok {
			go func() {
				WikiDocUrl := strings.ReplaceAll(docUrl, "{.id}", node.ObjToken)
				docResByte := sendRequest(WikiDocUrl)
				var res docJson
				json.Unmarshal(docResByte, &res)
				if res.Code == 0 {
					DealWithDoc(res)
				} else {
					// 也可能是表格？
					GetSheetIdAndBlockToken(WikiBaseUrl+node.WikiToken, node.ObjToken, node.Title)
				}
			}()
		}
		// 是否查询过子目录
		if _, ok := WiNode.Load(node.WikiToken); !ok {
			subdirectoryRes := sendRequest(WikiListUrl + node.WikiToken)
			var wikiList1 WikiJson
			json.Unmarshal(subdirectoryRes, &wikiList1)

			WiNode.Store(node.WikiToken, struct{}{})
			go func() {
				GetWikiDir(wikiList1)
			}()
		}
		time.Sleep(3 * time.Second)
	}
	return
}
