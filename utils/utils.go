package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

const (
	CONST_BASE_DATA_DIR   = "data"    // 下载数据目录
	CONST_BASE_TMP_DIR    = ".tmp"    // 临时数据目录
	CONST_BASE_OS_TMP_DIR = "pulldlr" // 临时数据目录
)

func Logger() *zerolog.Logger {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	// consoleWriter.FormatLevel = func(i interface{}) string {
	// 	return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	// }

	consoleWriter.FormatFieldValue = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("%s", i))
	}
	// multi := zerolog.MultiLevelWriter(consoleWriter, os.Stdout)
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()
	return &logger
}

func LoggerDebug(msg string) {
	Logger().Debug().Msg(msg)
}
func LoggerInfo(msg string) {
	Logger().Info().Msg(msg)
}
func LoggerError(msg string) {
	Logger().Error().Msg(msg)
}

func GetBaseUrl(srcUrl string) string {
	u, _ := url.Parse(srcUrl)
	return u.Scheme + "://" + u.Host
}

// 获取临时文件目录
func GetDownloadTmpDir() string {
	dir, _ := os.Getwd()
	return path.Join(dir, CONST_BASE_DATA_DIR, CONST_BASE_TMP_DIR)
}

// 获取下载文件目录
func GetDownloadDataDir() string {
	dir, _ := os.Getwd()
	return path.Join(dir, CONST_BASE_DATA_DIR)
}

// 创建系统级临时文件，返回一个文件句柄
func CreateTmpFile() (tmpFile *os.File, err error) {
	tmpDir := path.Join(os.TempDir(), CONST_BASE_OS_TMP_DIR)
	err1 := os.MkdirAll(tmpDir, os.ModePerm)
	if err1 != nil {
		return nil, err1
	}
	file, err := ioutil.TempFile(tmpDir, CONST_BASE_OS_TMP_DIR+"_")
	return file, err
}

// 清理系统级临时文件，返回一个文件句柄
func CleanTmpFile() error {
	tmpDir := path.Join(os.TempDir(), CONST_BASE_OS_TMP_DIR)
	return os.RemoveAll(tmpDir)
}

func CopyFile(sourceFile string, distFile string) (err error) {
	LoggerDebug("CopyFile: " + sourceFile + " -> " + distFile)
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(distFile, input, 0644)
	if err != nil {
		return err
	}
	return nil
}

func GetMD5(str string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(str))
	return hex.EncodeToString(md5Ctx.Sum(nil))
}

// 清理分片文件中的无用数据，影响分片合并后的播放
func CleanSliceUselessData(sliceData []byte) (result []byte) {
	// syncByte := uint8(71) // 0x47 的十进制
	syncTag1 := 0x47
	syncTag2 := 0x40

	bLen := len(sliceData)
	for j := 0; j < bLen; j++ {
		if j > 188 { // 188个字节以前没有找到需要清理的数据，则停止清理任务
			// result = sliceData[j:]
			break
		}
		// 清除无用数据
		if sliceData[j] == byte(syncTag1) && sliceData[j+1] == byte(syncTag2) {
			// fmt.Printf("===== %d / %d ======\n", j, bLen)
			result = sliceData[j:]
			break
		}
	}
	return
}

func DrawProgressBar(prefix string, proportion float32, width int, suffix ...string) {
	// Debug 不显示进度条
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		LoggerInfo("******* 视频下载进度：" + strconv.Itoa(int(proportion*100)) + "% ")
		return
	} else {
		pos := int(proportion * float32(width))
		s := fmt.Sprintf("【%s】 [%s%*s] %6.2f%% \t%s",
			prefix, strings.Repeat("■", pos), width-pos, "", proportion*100, strings.Join(suffix, ""))
		fmt.Print("\r" + s)
	}

}
