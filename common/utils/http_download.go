package utils

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"regexp"

	"gitee.com/mangenotework/commander/common/logger"
)

// HTTPDownLoad 下载
// TODO 多线程下载
func HTTPDownLoad(url, filePath string) (string, error) {
	filename := ""
	resp, err := http.Head(url)
	if err != nil {
		logger.Error("resp, err := http.Head(strURL)  报错: strURL = ", url, err)
		return filename, err
	}

	logger.Info("%#v\n", resp)
	fileLength := int(resp.ContentLength)
	ContentDisposition := resp.Header.Get("Content-Disposition")
	fileNameList := RegFindAll(`(?is:filename="(.*?)")`, ContentDisposition)
	if len(fileNameList) > 0 {
		filename = filePath + fileNameList[0]
	} else {
		logger.Error("未知文件名")
		return filename, fmt.Errorf("未知文件名")
	}

	logger.Info("filename = ", filename)
	// 文件是否存在
	ok, _ := PathExists(filename)
	if ok {
		// 删除文件
		err = os.RemoveAll(filename)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return filename, err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", 0, fileLength))
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("http.DefaultClient.Do(req)", "error")
		return filename, err
	}
	defer resp.Body.Close()

	// 创建文件
	//filename = path.Base(filename)
	logger.Info("filename = ", filename)
	flags := os.O_CREATE | os.O_WRONLY
	f, err := os.OpenFile(filename, flags, fs.ModePerm)
	if err != nil {
		fmt.Println("创建文件失败: ", err)
	}
	defer f.Close()

	// 写入数据
	buf := make([]byte, 16*1024)
	_, err = io.CopyBuffer(f, resp.Body, buf)
	if err != nil {
		return filename, err
	}
	return filename, nil
}

// RegFindAll 正则匹配所有
func RegFindAll(regStr, rest string) (dataList []string) {
	reg := regexp.MustCompile(regStr)
	resList := reg.FindAllStringSubmatch(rest, -1)
	for _, v := range resList {
		if len(v) < 1 {
			continue
		}
		dataList = append(dataList, v[1])
	}
	return
}
