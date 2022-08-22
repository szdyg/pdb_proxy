package pdb

import (
	"github.com/gin-gonic/gin"
	"os"
	"path"
	"pdb_proxy/conf"
)

func PdbQuery(c *gin.Context) {
	pdbName := c.Param("pdbname")
	pdbHash := c.Param("pdbhash")
	pdbQuery := pdbName + "/" + pdbHash + "/" + pdbName
	pdbPath := path.Join(conf.PdbDir, pdbQuery)
	_, err := os.Stat(pdbPath)
	if err == nil {
		c.File(pdbPath)
	} else {
		pdbUrl := conf.PdbServer + "/" + pdbQuery
		err := DownLoadFile(pdbUrl, pdbPath)
		if err != nil {
			os.Remove(pdbPath)
			c.String(500, "download pdb err")
		} else {
			c.File(pdbPath)
		}
	}

}
