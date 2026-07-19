package pdb

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownLoadFile(url string, filePath string) error {

	//log.Printf("Download file from %s to %s", url, filepath)

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.New("file not exist")
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.CreateTemp(dir, ".pdb-download-*")
	if err != nil {
		return err
	}
	tempPath := file.Name()
	defer os.Remove(tempPath)

	if _, err = io.Copy(file, res.Body); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return replaceCachedFile(tempPath, filePath)
}
