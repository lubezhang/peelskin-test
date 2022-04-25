package utils

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpGetFile(t *testing.T) {
	assetObj := assert.New(t)

	httpUrl := "http://mp.zhisland.com/wmp/user/news/2109180006/consult/viewpoint?page=1&t=1649902826233"
	data, _ := HttpGetFile(httpUrl)
	strData := string(data)
	assetObj.Equal(strings.Contains(strData, "errorCode"), true)
}

func TestDownloadeSliceFile(t *testing.T) {
	url := "https://v3.dious.cc/20220331/NGxXAbhN/2000kb/hls/AaxF3XHP.ts1"
	fileName := "/Users/zhangqinghong/study/go/hls-dl/data/1-1.ts"
	_, err := DownloadeSliceFile(url, fileName, "")
	if err != nil {
		fmt.Println("=================", err)
		// t.Error(err)
	}
}

func TestGetMD5(t *testing.T) {
	assetObj := assert.New(t)
	assetObj.Equal(GetMD5("123456"), "e10adc3949ba59abbe56e057f20f883e")
}

func TestCreateTmpFile(t *testing.T) {
	assetObj := assert.New(t)
	tmpFile, err := CreateTmpFile()
	assetObj.Nil(err)
	tmpFile.WriteString("123456")
	defer tmpFile.Close()

	content, _ := ioutil.ReadFile(tmpFile.Name())
	assetObj.Equal(string(content), "123456")
}

func TestCleanTmpFile(t *testing.T) {
	CleanTmpFile()
}
