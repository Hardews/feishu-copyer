/**
 * @Author: Hardews
 * @Date: 2023/5/25 21:53
 * @Description: 初始化参数
**/

package logic

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type Url struct {
	DirToken  []string `json:"dir_token"`
	WikiToken []string `json:"wiki_token"`
	Url       string   `json:"url"`
	Xcrf      string   `json:"xcrf"`
}

var (
	cookie string
	xcrf   string
	urls   Url
)

var (
	docUrl        string
	cellUrl       string
	FolderListUrl string
	WikiListUrl   string
	WikiBaseUrl   string
)

func Init() {
	// 读取 cookie
	cookieFile, err := os.Open("./cookie.txt")
	if err != nil {
		log.Println("open cookie file failed,err is:", err)
		return
	}
	cookieByte, err := io.ReadAll(cookieFile)
	if err != nil {
		log.Println("read cookie file failed,err is:", err)
		return
	}

	cookie = string(cookieByte)

	// 读取配置
	urlFile, err := os.Open("./url.json")
	if err != nil {
		log.Println("open url file failed,err is:", err)
		return
	}

	urlByte, err := io.ReadAll(urlFile)
	if err != nil {
		log.Println("read url file failed,err is:", err)
		return
	}

	// 获取 dir token
	json.Unmarshal(urlByte, &urls)

	// 设置 csrf token
	xcrf = urls.Xcrf

	// 初始化所有 url
	FolderListUrl = "https://" + urls.Url + "/space/api/explorer/v3/children/list/?thumbnail_width=1028&thumbnail_height=1028&thumbnail_policy=4&obj_type=0&obj_type=2&obj_type=22&obj_type=3&obj_type=30&obj_type=8&obj_type=11&obj_type=12&length=100&asc=1&rank=5&token="
	WikiListUrl = "https://" + urls.Url + "/space/api/wiki/v2/tree/get_info/?wiki_token="
	WikiBaseUrl = "https://" + urls.Url + "/wiki/"
	docUrl = "https://" + urls.Url + "/space/api/docx/pages/client_vars?id={.id}&open_type=1"
	cellUrl = "https://" + urls.Url + "/space/api/v2/sheet/sub_block?token={.token}&block_token={.block_token}&schema_version=1"
}
