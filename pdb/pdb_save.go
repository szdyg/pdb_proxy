package pdb

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	// 单次下载的总时长上限。符号文件可能有几百 MB，给足余量，
	// 真正用来兜住"上游卡死"的是 ResponseHeaderTimeout。
	downloadTimeout = 30 * time.Minute
	// 缓存文件权限。os.CreateTemp 建出来是 0600，rename 前显式放宽，
	// 便于 nginx 等其他进程直接读缓存目录。
	cacheFileMode = 0644

	tempPrefix = ".dl-"
	// 清理临时文件的年龄阈值，必须大于 downloadTimeout，
	// 这样永远不会误删一个仍在进行中的下载。
	tempFileTTL = time.Hour
)

var httpClient = &http.Client{
	Timeout: downloadTimeout,
	Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConnsPerHost:   32,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	},
}

// DownLoadFile 从 url 下载内容存到 filepath。
//
// 先写同目录下的临时文件、校验长度、再 rename，保证 filepath 在任何时刻
// 要么不存在、要么内容完整：并发请求同一符号不会互相截断，进程中途被杀
// 也不会留下会被后续请求当成有效缓存的残缺文件。
func DownLoadFile(url string, filepath string) error {
	res, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.New("file not exist")
	}

	dir := path.Dir(filepath)
	// 0755：目录缺少执行位时，非 root 进程无法在其中创建文件。
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	cleanStaleTemps(dir)

	// 临时文件必须和目标同目录，否则 rename 可能跨文件系统而失去原子性。
	tmp, err := os.CreateTemp(dir, tempPrefix+"*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// rename 成功后这次 Remove 会失败，无副作用；失败路径上则负责清理。
	defer os.Remove(tmpName)

	n, err := io.Copy(tmp, res.Body)
	if err != nil {
		tmp.Close()
		return err
	}
	// 上游给了 Content-Length 却没发够，说明是截断的响应，不能入缓存。
	if res.ContentLength >= 0 && n != res.ContentLength {
		tmp.Close()
		return errors.New("short read from upstream")
	}
	if err := tmp.Chmod(cacheFileMode); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpName, filepath)
}

// cleanStaleTemps 清掉进程被强杀时遗留在该目录里的下载临时文件。
// 只扫描当前符号自己的目录（通常只有一两个条目），代价可以忽略。
func cleanStaleTemps(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), tempPrefix) {
			continue
		}
		info, err := e.Info()
		if err != nil || time.Since(info.ModTime()) < tempFileTTL {
			continue
		}
		os.Remove(path.Join(dir, e.Name()))
	}
}
