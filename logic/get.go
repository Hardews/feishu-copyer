/**
 * @Author: Hardews
 * @Date: 2023/5/25 21:53
 * @Description: 发送请求获取响应数据
**/

package logic

import (
	"io"
	"log"
	"net/http"
	"time"
)

// 发送请求的封装函数
func sendRequest(url string) (resByte []byte) {
	if url == "" {
		return
	}
	var (
		err    error
		req    *http.Request
		resp   *http.Response
		client = &http.Client{Timeout: 30 * time.Second}
	)

	req, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("new req failed, url:%s, err:%s\n", url, err)
		return
	}

	req.Header.Set("cookie", cookie)
	req.Header.Set("Referer", url)
	req.Header.Set("x-csrftoken", xcrf)

	resp, err = client.Do(req)
	if err != nil {
		log.Printf("send req failed, url:%s, err:%s\n", url, err)
		return
	}

	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("read resp failed, url:%s, err:%s\n", url, err)
		return
	}
	// 这是为了让它可以得到结果，也是控制它的爬取速度
	time.Sleep(5 * time.Second)
	return res
}
