package pulldlr

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/lubezhang/hls-parse/protocol"
	"github.com/lubezhang/pulldlr/utils"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog"
)

func Command() {
	urlFlag := flag.String("u", "", "m3u8下载地址(http(s)://url/xx/xx/index.m3u8)")
	// nFlag := flag.Int("n", 16, "下载线程数(max goroutines num)")
	oFlag := flag.String("o", "output", "自定义文件名(默认为output)")

	flag.Parse()

	m3u8Url := *urlFlag

	if m3u8Url == "" {
		utils.LoggerError("请输入m3u8地址")
		return
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	ShowProtocolInfo(m3u8Url)

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	dl, _ := New(m3u8Url)
	dl.SetOpts(DownloaderOption{
		FileName: *oFlag,
	})
	dl.Start()
}

func CommandUI() {
	m3u8Url := commandInputUrl()
	if m3u8Url == "" {
		fmt.Println("请输入m3u8地址")
		return
	}

	fileName := commandInputFileName()
	if m3u8Url == "" {
		fmt.Println("请输入下载文件名")
		return
	}

	theadNum := commandSelectThead()
	if theadNum == 0 {
		fmt.Println("请选择并发下载线程数:", theadNum)
	}

	vodUrl, err1 := commandSelectStream(m3u8Url)
	if err1 != nil {
		fmt.Println(err1)
		return
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	ShowProtocolInfo(vodUrl)

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	dl, _ := New(vodUrl)
	dl.SetOpts(DownloaderOption{
		FileName:  fileName,
		MaxThread: theadNum,
	})
	dl.Start()
}

func commandInputUrl() string {
	prompt := promptui.Prompt{
		Label: "请输入下载地址",
	}

	result, err := prompt.Run()

	if err != nil {
		// fmt.Printf("Prompt failed %v\n", err)
		return ""
	}

	return result
}

func commandInputFileName() string {
	prompt := promptui.Prompt{
		Label: "请输入下载文件名",
	}

	result, err := prompt.Run()

	if err != nil {
		// fmt.Printf("Prompt failed %v\n", err)
		return ""
	}

	return result
}

func commandSelectThead() (result int) {
	prompt := promptui.Select{
		Label: "请选择并发下载线程数",
		Items: []int{1, 5, 10, 15, 20},
	}

	_, res, err := prompt.Run()
	if err != nil {
		// fmt.Printf("Prompt failed %v\n", err)
		return 0
	}
	result, _ = strconv.Atoi(res)
	return
}

func commandSelectStream(m3u8Url string) (result string, err error) {
	err = nil
	baseUrl := utils.GetBaseUrl(m3u8Url)
	data1, _ := utils.HttpGetFile(m3u8Url)
	strDat1 := string(data1)
	hlsBase, _ := protocol.ParseString(&strDat1, baseUrl)

	if hlsBase.IsVod() {
		return m3u8Url, nil
	}

	if hlsBase.IsMaster() {
		templates := &promptui.SelectTemplates{
			Label:    "{{ .Resolution }}",
			Active:   "{{ .Resolution | cyan }}",
			Inactive: "{{ .Resolution}} ",
			Selected: "{{ .Resolution | cyan }}",
		}
		master, _ := hlsBase.GetMaster()

		prompt := promptui.Select{
			Label:     "请选择下载的视频流",
			Items:     master.StreamInfs,
			Templates: templates,
		}

		idx, _, err := prompt.Run()
		if err != nil {
			// fmt.Printf("Prompt failed %v\n", err)
			return "", err
		}
		result = master.StreamInfs[idx].Url
	}

	return
}
