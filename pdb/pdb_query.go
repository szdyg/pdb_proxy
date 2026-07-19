package pdb

import (
	"os"
	"path"
	"pdb_proxy/conf"
	"time"

	"github.com/gofiber/fiber/v2"
)

func PdbQuery(c *fiber.Ctx) error {
	pdbName := c.Params("pdbname")
	pdbHash := c.Params("pdbhash")
	pdbQuery := pdbName + "/" + pdbHash + "/" + pdbName
	pdbPath := path.Join(conf.PdbDir, pdbQuery)
	//log.Printf("Pdb Path: %s", pdbPath)

	fileInfo, err := os.Stat(pdbPath)
	if err == nil && cacheIsFresh(fileInfo, time.Now()) {
		return sendCachedFile(c, pdbPath)
	}

	pdbUrl := conf.PdbServer + "/" + pdbQuery
	err = DownLoadFile(pdbUrl, pdbPath)
	if err != nil {
		if fileInfo != nil {
			return sendCachedFile(c, pdbPath)
		}
		return c.Status(404).SendString("")
	}
	return sendCachedFile(c, pdbPath)
}

func cacheIsFresh(fileInfo os.FileInfo, now time.Time) bool {
	if conf.PdbCacheTTL == 0 {
		return true
	}
	return now.Sub(fileInfo.ModTime()) < conf.PdbCacheTTL
}

func sendCachedFile(c *fiber.Ctx, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}
	c.Type(path.Ext(filePath))
	return c.SendStream(file, int(fileInfo.Size()))
}
