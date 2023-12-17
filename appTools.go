package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func unzip(zipFile, destination string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, file := range r.File {
		filePath := filepath.Join(destination, file.Name)

		// If the entry is a directory, create it
		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		// Create the file
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		// Open the zip file
		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		// Copy the contents of the file
		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
	}

	return nil
}
