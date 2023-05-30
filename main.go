/**
 * @Author: Hardews
 * @Date: 2023/5/25 16:46
 * @Description:
**/

package main

import (
	"feishu-crawler/logic"
)

func main() {
	// 初始化配置
	logic.Init()
	// 获取所有共享文件夹的路径
	go logic.GetAllFolderDir()
	// 获取所有知识库的路径
	logic.GetAllWikiDir()
	for true {
	}
}
