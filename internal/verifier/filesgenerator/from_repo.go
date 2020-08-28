package filesgenerator

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/baking-bad/bcdhub/internal/verifier/compilers"
)

// FromRepo - download files from github or gitlab repo url and save them to dir
func FromRepo(url, dir string) ([]string, error) {
	data, err := downloadFile(url)
	if err != nil {
		return nil, err
	}

	return unzipFiles(data, dir)
}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("file not found")
	}

	return ioutil.ReadAll(resp.Body)
}

func unzipFiles(data []byte, dest string) ([]string, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	var filenames []string

	// Read all the files from zip archive
	for _, zipFile := range zipReader.File {
		fpath := filepath.Join(dest, zipFile.Name)

		if zipFile.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if !compilers.IsValidExtension(filepath.Ext(zipFile.Name)) {
			continue
		}

		filenames = append(filenames, fpath)

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return nil, err
		}

		f, err := zipFile.Open()
		if err != nil {
			return nil, fmt.Errorf("zipFile.Open() %v", err)
		}
		defer f.Close()

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			return nil, fmt.Errorf("os.OpenFile() %v", err)
		}
		defer outFile.Close()

		if _, err = io.Copy(outFile, f); err != nil {
			return nil, err
		}
	}

	return filenames, nil
}
