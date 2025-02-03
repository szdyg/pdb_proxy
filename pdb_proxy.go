package main

import (
	"github.com/gin-gonic/gin"
	"pdb_proxy/conf"
	"pdb_proxy/pdb"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.GET("/download/symbols/:pdbname/:pdbhash/:pdbname", pdb.PdbQuery)
	r.Run(conf.ServerPort)
}
