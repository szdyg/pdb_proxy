package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"pdb_proxy/conf"
	"pdb_proxy/pdb"
)

const symbolRoute = "/download/symbols/:pdbname/:pdbhash/:pdbname"

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.GET(symbolRoute, pdb.PdbQuery)
	// 部分符号工具会先用 HEAD 探测符号是否存在，再发 GET 下载。
	r.HEAD(symbolRoute, pdb.PdbQuery)

	srv := &http.Server{
		Addr:    conf.ServerPort,
		Handler: r,
		// 不设 WriteTimeout：符号文件可能有几百 MB，写超时会掐断正常的大文件下载。
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// ReleaseMode 下 gin 不会打印 Run 的错误，直接 Fatal，避免监听失败时静默退出。
	log.Fatal(srv.ListenAndServe())
}
