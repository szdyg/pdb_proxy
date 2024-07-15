package main

import (
	"github.com/gin-gonic/gin"
	"net/http/httputil"
	"net/url"
	"pdb_proxy/conf"
	"pdb_proxy/pdb"
)

func PassBy(c *gin.Context) {
	remote, err := url.Parse("http://msdl.microsoft.com")
	if err != nil {
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	c.Request.URL.Path = "" //请求API
	proxy.ServeHTTP(c.Writer, c.Request)
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.GET("/download/symbols/:pdbname/:pdbhash/:pdbname", pdb.PdbQuery)
	r.Run(conf.ServerPort)
}
