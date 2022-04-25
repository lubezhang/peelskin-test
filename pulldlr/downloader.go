package pulldlr

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/imdario/mergo"
	"github.com/lubezhang/hls-parse/protocol"
	"github.com/lubezhang/hls-parse/types"
	"github.com/lubezhang/pulldlr/utils"
)

const (
	CONST_BASE_SLICE_FILE_EXT  = ".ts" // 分片文件扩展名
	CONST_BASE_RETRY_MAX_COUNT = 3     // 下载分片文件的最大重试次数
)

func New(url string) (result *Downloader, err error) {
	result = &Downloader{
		m3u8Url: url,
	}
	return result, nil
}

// 下载器参数
type DownloaderOption struct {
	FileName  string // 文件名
	MaxThread int    // 最大下载线程数
}

// 下载器
type Downloader struct {
	m3u8Url    string            // m3u8文件地址
	hlsBase    *protocol.HlsBase // 协议基础对象
	selectVod  *protocol.HlsVod  // 选择下载的视频
	wg         sync.WaitGroup    // 并发线程管理容器
	opts       DownloaderOption  // 下载器参数
	cache      DownloadCacheData // 下载数据管理器
	sliceCount int               // 下载进度，完成文件合并的分片数量
}

// 设置参数
func (dl *Downloader) SetOpts(opts1 DownloaderOption) {
	dl.opts = opts1
}

// 开始下载m3u8文件
func (dl *Downloader) Start() {
	dl._init()
	if reflect.ValueOf(dl.selectVod).IsValid() && len(dl.selectVod.ExtInfs) > 0 {
		go dl.mergeVodFileToMp4()
		dl.startDownload()
		dl.wg.Wait()

		// 正常下载完成，处理异常集合的分片

		dl.cleanTmpFile() // 下载完成，清理临时数据
	} else {
		utils.LoggerInfo("没有选择下载的视频")
		return
	}
	// utils.LoggerInfo("<<<<<<< 下载视频完成:" + dl.opts.FileName)
}

func (dl *Downloader) CheckMaster() (result protocol.HlsMaster, err error) {
	// return path.Join(utils.GetDownloadTmpDir(), utils.GetMD5(dl.opts.FileName), dl.opts.FileName)
	baseUrl := utils.GetBaseUrl(dl.m3u8Url)

	data1, _ := utils.HttpGetFile(dl.m3u8Url)
	strDat1 := string(data1)
	hlsBase, _ := protocol.ParseString(&strDat1, baseUrl)
	if hlsBase.IsMaster() {
		result, _ = hlsBase.GetMaster()
		err = nil
	} else {
		err = errors.New("不是主文件")
	}

	return
}

func (dl *Downloader) _init() {
	dl.sliceCount = 0
	// 设置默认参数
	defaultOpts := DownloaderOption{
		FileName:  time.Now().Format("2006-01-02$15:04:05") + ".mp4", // 生成临时文件名
		MaxThread: 10,
	}
	mergo.MergeWithOverwrite(&defaultOpts, dl.opts) // 合并自定义和默认参数
	// utils.LoggerInfo(">>>>>>> 下载视频:" + defaultOpts.FileName)
	dl.SetOpts(defaultOpts)

	dl.selectMediaVod()
}

func (dl *Downloader) mergeVodFileToMp4() {
	sliceTotal := len(dl.selectVod.ExtInfs)
	if dl.sliceCount >= sliceTotal { // 所有分片文件已经完成合并
		dl.wg.Done()
		return
	}

	for {
		curProgress := sliceTotal - dl.cache.ReadyLen()
		utils.DrawProgressBar(dl.opts.FileName, float32(curProgress)/float32(sliceTotal), 80)
		// 检查片文件是否存在
		sliceFilePath := dl.getTmpFilePath(strconv.Itoa(dl.sliceCount))
		_, err1 := os.Stat(sliceFilePath)
		if err1 != nil {
			break
		}
		utils.LoggerDebug("读取一个分片文件:" + sliceFilePath)

		// 读取一个分片文件
		tsFile, err2 := os.OpenFile(sliceFilePath, os.O_RDONLY, os.ModePerm)
		if err2 != nil {
			break
		}

		buf, _ := ioutil.ReadAll(tsFile)
		buf = utils.CleanSliceUselessData(buf)
		vodFile, _ := os.OpenFile(dl.getVodFilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
		vodFile.Write(buf)

		tsFile.Close()
		vodFile.Close()
		dl.sliceCount = dl.sliceCount + 1
		time.Sleep(80 * time.Millisecond)
	}
	time.Sleep(1 * time.Second) // 等待一会，在继续执行合并操作
	dl.mergeVodFileToMp4()
}

func (dl *Downloader) startDownload() {
	dl.setDecryptKey()
	dl.setDwnloadCache()
	dl.startVodDownloadThead()
}

// 启动Vod文件下载线程，使用多线程并发下载视频分片
func (dl *Downloader) startVodDownloadThead() {
	for i := 0; i < dl.opts.MaxThread; i++ {
		go dl.downloadVodFile()
	}
}

// 将视频片放到数据下载管理器中
func (dl *Downloader) setDwnloadCache() {
	hlsVod := dl.selectVod
	if len(hlsVod.ExtInfs) == 0 {
		return
	}

	var list []DownloadData
	for idx, extinf := range hlsVod.ExtInfs {
		var decryptKey = ""
		if extinf.EncryptIndex >= 0 {
			decryptKey = hlsVod.Extkeys[extinf.EncryptIndex].Key
		}

		list = append(list, DownloadData{
			Index:        idx,
			Key:          utils.GetMD5(extinf.Url),
			Title:        extinf.Title,
			Url:          extinf.Url,
			DownloadPath: dl.getTmpFilePath(strconv.Itoa(idx)),
			EncryptKey:   decryptKey,
		})
	}

	dl.wg.Add(len(list) + 1) // 初始化并发线程计数器
	dl.cache.Push(list)
}

func (dl *Downloader) downloadVodFile() {
	for {
		data, err := dl.cache.Pop()
		if err != nil {
			return
		}
		_, err1 := utils.DownloadeSliceFile(data.Url, data.DownloadPath, data.EncryptKey)
		// 如果下载失败，重试5次
		if err1 != nil {
			utils.LoggerDebug("下载失败重试:" + data.Url)
			for i := 1; i <= CONST_BASE_RETRY_MAX_COUNT; i++ {
				_, err2 := utils.DownloadeSliceFile(data.Url, data.DownloadPath, data.EncryptKey)
				if i >= CONST_BASE_RETRY_MAX_COUNT { // 重试后，还是错误，放入到异常集合中
					dl.cache.Complete(data, err2)

					// 视频分片下载失败，写入一个空文件，保证文件合并线程正常执行
					blankFile, _ := os.OpenFile(data.DownloadPath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
					defer blankFile.Close()
					blankFile.Write([]byte(""))

					break
				}

				if err2 == nil {
					dl.cache.Complete(data, nil)
					break
				}
			}
		} else {
			utils.LoggerDebug("分片下载完成:" + data.DownloadPath)
			dl.cache.Complete(data, nil)
		}
		time.Sleep(20 * time.Millisecond)
		dl.wg.Done()
	}
}

// 获取Vod协议，如果是主文件取视频流的第一个
func (dl *Downloader) selectMediaVod() (err error) {
	utils.LoggerInfo("获取Vod协议文件对象")
	baseUrl := utils.GetBaseUrl(dl.m3u8Url)

	data1, _ := utils.HttpGetFile(dl.m3u8Url)
	strDat1 := string(data1)
	hlsBase, _ := protocol.ParseString(&strDat1, baseUrl)

	if hlsBase.IsMaster() {
		dl.hlsBase = &hlsBase
		hlsMaster, _ := hlsBase.GetMaster()
		if len(hlsMaster.StreamInfs) == 0 {
			return errors.New("master中没有视频流")
		}
		data2, _ := utils.HttpGetFile(hlsMaster.StreamInfs[0].Url)
		strData2 := string(data2)
		hlsBas2, _ := protocol.ParseString(&strData2, baseUrl)
		if hlsBas2.IsVod() {
			selectVod, _ := hlsBas2.GetVod()
			dl.selectVod = &selectVod
		}
	} else if hlsBase.IsVod() {
		selectVod, _ := hlsBase.GetVod()
		dl.selectVod = &selectVod
	} else {
		return errors.New("没有视频回放文件")
	}
	err = nil
	return
}

// 通过链接获取加密密钥，并将密钥填充到加密数据结构中
func (dl *Downloader) setDecryptKey() {
	hls := dl.selectVod
	if len(hls.Extkeys) == 0 {
		return
	}

	var keys []types.TagExtKey
	for _, extkey := range hls.Extkeys {
		if extkey.Method == "AES-128" {
			tmp := extkey
			data, _ := utils.HttpGetFile(extkey.Uri)
			tmp.Key = string(data)
			keys = append(keys, tmp)
		}
	}

	hls.Extkeys = keys
}

func (dl *Downloader) getTmpFilePath(fileName string) string {
	return path.Join(utils.GetDownloadTmpDir(), utils.GetMD5(dl.opts.FileName), fileName+CONST_BASE_SLICE_FILE_EXT)
}
func (dl *Downloader) getVodFilePath() string {
	return path.Join(utils.GetDownloadDataDir(), dl.opts.FileName)
}

// 清理临时文件
func (dl *Downloader) cleanTmpFile() error {
	tmpDir := path.Join(utils.GetDownloadTmpDir(), utils.GetMD5(dl.opts.FileName))
	utils.LoggerInfo("清理临时文件:" + tmpDir)
	err := os.RemoveAll(tmpDir)
	if err != nil {
		utils.LoggerError("清理临时文件失败:" + err.Error())
		return err
	}
	return nil
}
