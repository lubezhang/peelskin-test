package pulldlr

import (
	"fmt"

	"github.com/lubezhang/hls-parse/protocol"
	"github.com/lubezhang/pulldlr/utils"
)

func ShowProtocolInfoMaster(url string) {
	baseUrl := utils.GetBaseUrl(url)

	data1, _ := utils.HttpGetFile(url)
	strDat1 := string(data1)
	hlsBase, err := protocol.ParseString(&strDat1, baseUrl)

	if err != nil {
		fmt.Println(err)
		return
	}

	if hlsBase.IsMaster() {
		hlsMaster, _ := hlsBase.GetMaster()
		showProtocolMaster(hlsMaster)
	}
}

// 显示协议信息
func ShowProtocolInfo(url string) {
	baseUrl := utils.GetBaseUrl(url)

	data1, _ := utils.HttpGetFile(url)
	strDat1 := string(data1)
	hlsBase, err := protocol.ParseString(&strDat1, baseUrl)

	if err != nil {
		fmt.Println(err)
		return
	}

	if hlsBase.IsMaster() {
		hlsMaster, _ := hlsBase.GetMaster()
		showProtocolMaster(hlsMaster)

		if len(hlsMaster.StreamInfs) == 0 {
			return
		}

		data2, _ := utils.HttpGetFile(hlsMaster.StreamInfs[0].Url)
		strData2 := string(data2)
		hlsBase2, _ := protocol.ParseString(&strData2, baseUrl)
		if hlsBase2.IsVod() {
			hlsVod, _ := hlsBase2.GetVod()
			showProtocolVod(hlsVod)
		}
	} else if hlsBase.IsVod() {
		hlsVod, _ := hlsBase.GetVod()
		showProtocolVod(hlsVod)
	} else {
		fmt.Println("没有协议")
	}
}

func showProtocolMaster(hls protocol.HlsMaster) {
	fmt.Println("")
	fmt.Println("******* master文件 *******")
	fmt.Println("Stream 数量:", len(hls.StreamInfs))
	for idx, stream := range hls.StreamInfs {
		fmt.Printf("Stream-%d 分辨率:%s \n", idx+1, stream.Resolution)
	}
	fmt.Println("******* master文件 *******")
	fmt.Println("")
}

func showProtocolVod(hls protocol.HlsVod) {
	fmt.Println("")
	fmt.Println("******* VOD文件 *******")
	fmt.Println("分片数量：", len(hls.ExtInfs))
	fmt.Println("是否加密：", len(hls.Extkeys) > 0)
	fmt.Println("******* VOD文件 *******")
	fmt.Println("")
}
