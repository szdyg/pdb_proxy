package pdb

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
)

func DownLoadFile(url string, filepath string) error {

	//log.Printf("Download file from %s to %s", url, filepath)

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.New("file not exist")
	}

	os.Remove(filepath)
	dir := path.Dir(filepath)
	os.MkdirAll(dir, 0644)

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
