package pdb

import (
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/singleflight"
	"pdb_proxy/conf"
)

// 符号服务器的 hash 段是 GUID+age（PDB）或 timestamp+size（PE），只可能是十六进制。
var pdbHashRe = regexp.MustCompile(`^[0-9A-Fa-f]{8,40}$`)

// 合并同一符号的并发回源：N 个客户端同时请求未缓存符号时只打一次上游。
var downloadGroup singleflight.Group

func PdbQuery(c *gin.Context) {
	pdbName := c.Param("pdbname")
	pdbHash := c.Param("pdbhash")

	// 显式校验，不依赖 http.ServeFile 内部的 containsDotDot 兜底防穿越。
	if !validPdbName(pdbName) || !pdbHashRe.MatchString(pdbHash) {
		respondNotFound(c)
		return
	}

	pdbQuery := pdbName + "/" + pdbHash + "/" + pdbName
	pdbPath := path.Join(conf.PdbDir, pdbQuery)

	if isRegularFile(pdbPath) {
		servePdb(c, pdbPath)
		return
	}

	_, err, _ := downloadGroup.Do(pdbQuery, func() (interface{}, error) {
		// 等锁期间可能已被同组的另一个请求下载完成。
		if isRegularFile(pdbPath) {
			return nil, nil
		}
		return nil, DownLoadFile(conf.PdbServer+"/"+pdbQuery, pdbPath)
	})
	if err != nil {
		respondNotFound(c)
		return
	}

	servePdb(c, pdbPath)
}

func servePdb(c *gin.Context, pdbPath string) {
	// 路径里带 GUID，内容不可变，允许 CDN 和客户端长期缓存。
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.File(pdbPath)
}

// respondNotFound 的响应体必须为空：symsrv 会把非空的错误响应当成有效负载存进本地缓存。
func respondNotFound(c *gin.Context) {
	c.Status(http.StatusNotFound)
}

func validPdbName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	return !strings.ContainsAny(name, `/\`)
}

func isRegularFile(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.Mode().IsRegular()
}
