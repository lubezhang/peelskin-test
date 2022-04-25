package pulldlr

// 下载文件结构
type DownloadData struct {
	Index        int    // 下载资源索引
	Key          string // 下载资源的唯一标示
	Title        string // 文件名
	Url          string // 下载链接
	DownloadPath string // 文件保存路径
	EncryptKey   string // 加密密钥
	// err?: DownloadDataErr// 下载异常数据
}
