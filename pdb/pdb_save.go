package pdb

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"path"
)

func DownLoadFile(url string, filepath string) error {

	os.Remove(filepath)
	dir := path.Dir(filepath)
	os.MkdirAll(dir, 0644)

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := bufio.NewWriter(file)
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return err
	}
	err = buf.Flush()
	if err != nil {
		return err
	}

	return nil
}
