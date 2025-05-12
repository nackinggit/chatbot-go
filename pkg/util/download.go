package util

import (
	"io"
	"net/http"
	"os"
	"time"
)

var downloadClient = &http.Client{
	Timeout: 60 * time.Second,
}

func OpenUrl(url string) (io.ReadCloser, error) {
	resp, err := downloadClient.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func DownloadFile(url string, filePath string) error {
	// 发起GET请求
	resp, err := downloadClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 创建本地文件
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将从网络读取的内容写入到本地文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
