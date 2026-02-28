package app

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/text/encoding/charmap"
)

func unzip(zipFile, destination string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, file := range r.File {
		// Decode CP852-encoded names (common in Czech/Central European zips)
		name := file.Name
		if file.NonUTF8 {
			if decoded, err := charmap.CodePage852.NewDecoder().String(name); err == nil {
				name = decoded
			}
		}

		filePath := filepath.Join(destination, name)

		// If the entry is a directory, create it
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				return err
			}
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

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func copyDir(src string, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectories
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy files
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
