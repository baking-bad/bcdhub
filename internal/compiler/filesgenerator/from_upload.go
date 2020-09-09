package filesgenerator

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/baking-bad/bcdhub/internal/compiler/compilers"
)

// FromUpload - get files from multipart form and save them to dir
func FromUpload(form *multipart.Form, dir string) ([]string, error) {
	var filenames []string

	for _, fileArray := range form.File {
		for _, file := range fileArray {
			if !compilers.IsValidExtension(filepath.Ext(file.Filename)) {
				return nil, fmt.Errorf("invalid file extension %s in %s", filepath.Ext(file.Filename), file.Filename)
			}

			filename := filepath.Join(dir, filepath.Base(file.Filename))

			if err := saveUploadedFile(file, filename); err != nil {
				return nil, err
			}

			filenames = append(filenames, filename)
		}
	}

	return filenames, nil
}

func saveUploadedFile(file *multipart.FileHeader, filename string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}
