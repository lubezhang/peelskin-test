package pulldlr

import (
	"errors"
	"sync"
)

type DownloadCacheData struct {
	lockDownload sync.Mutex
	lockComplete sync.Mutex

	downloadReady []DownloadData          // 待下载资源队列
	downloading   map[string]DownloadData // 正在下载的资源队列
	downloadError map[string]DownloadData // 下载异常资源队列
}

func (cache *DownloadCacheData) ReadyLen() int {
	return len(cache.downloadReady)
}

func (cache *DownloadCacheData) DownloadingLen() int {
	return len(cache.downloading)
}

func (cache *DownloadCacheData) ErrorLen() int {
	return len(cache.downloadError)
}

// 将需要加载的数据对象，添加到等待下载队列中
func (cache *DownloadCacheData) Push(list []DownloadData) {
	cache.downloadReady = append(cache.downloadReady, list...)
}

// 从待下载队列中取出一个下载对象，放到正在下载队列中
func (cache *DownloadCacheData) Pop() (result DownloadData, err error) {
	if len(cache.downloadReady) == 0 {
		return result, errors.New("ready queue is empty")
	}

	if cache.downloading == nil {
		cache.downloading = make(map[string]DownloadData)
	}
	// 处理并发
	cache.lockDownload.Lock()

	dr := cache.downloadReady[0]
	cache.downloading[dr.Key] = dr
	cache.downloadReady = cache.downloadReady[1:] // 删除第一个元素
	result = dr
	err = nil

	cache.lockDownload.Unlock()
	return
}

// 完成下载，将下载对象从正在下载队列中移除
// 如果有错误，暂时放到异常队列中，等待重试
func (cache *DownloadCacheData) Complete(downloadData DownloadData, err error) {
	cache.lockComplete.Lock()
	if cache.downloadError == nil {
		cache.downloadError = make(map[string]DownloadData)
	}
	if err == nil {
		delete(cache.downloading, downloadData.Key)
	} else {
		cache.downloadError[downloadData.Key] = downloadData
	}
	cache.lockComplete.Unlock()
}
