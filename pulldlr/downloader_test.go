package pulldlr

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestDownloader(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	// dl, _ := New("https://qq.sd-play.com/20220405/4Si6DIev/hls/index.m3u8")
	dl, _ := New("https://vod.bunediy.com/20200501/P5MEoEqd/index.m3u8")
	dl.SetOpts(DownloaderOption{
		FileName: "11.mp4",
	})
	dl.Start()
}

func TestShowProtocolInfo(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	// ShowProtocolInfo("https://qq.sd-play.com/20220405/4Si6DIev/hls/index.m3u8")
	ShowProtocolInfo("https://vod.bunediy.com/20200411/BrguJtZ4/index.m3u8")
}
