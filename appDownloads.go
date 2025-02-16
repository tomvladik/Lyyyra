package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func (a *App) downloadFile(fileUrl, fileName string) (string, error) {
	// Create or truncate the file in the app directory
	fullPath := filepath.Join(a.appDir, fileName)
	destFile, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	parsedURL, err := url.Parse(fileUrl)
	if err != nil {
		return "", err
	}

	if parsedURL.Scheme == "file" {
		sourceFile, err := os.Open(fmt.Sprintf("%s%s", parsedURL.Host, parsedURL.Path))
		if err != nil {
			return "", err
		}
		defer sourceFile.Close()
		_, err = io.Copy(destFile, sourceFile)
		if err != nil {
			return "", err
		}
	} else {
		// Make a GET request to the URL
		response, err := http.Get(fileUrl)
		if err != nil {
			return "", err
		}
		defer response.Body.Close()

		// Check if the request was successful (status code 200)
		if response.StatusCode != http.StatusOK {
			return "", fmt.Errorf("HTTP request failed with status %d", response.StatusCode)
		}

		// Copy the content from the response body to the file
		_, err = io.Copy(destFile, response.Body)
		if err != nil {
			return "", err
		}
	}
	slog.Info("Downloaded file", "fileUrl", fileUrl, "fullPath", fullPath)
	return fullPath, nil
}

func (a *App) downloadParts() {
	// Iterate over the nodes and fill the slice
	for i, node := range a.pdfFiles.Items {

		downloadUrl := fmt.Sprintf("%s://%s/%s", a.pdfFiles.UrlScheme, a.pdfFiles.Domain, node.Href)
		localFileName := fmt.Sprintf("%s.PDF", node.Title)
		fullFilePtah, err := a.downloadFile(downloadUrl, localFileName)
		if err != nil {
			slog.Error(err.Error())
		}
		slog.Info("Downloaded to " + fullFilePtah)
		a.pdfFiles.Items[i].LocalFileName = localFileName
		//parsePdf(fullFilePtah)
	}

}

func (a *App) DownloadSongBase() error {
	fileName, err := a.downloadFile(a.xmlUrl, "Songs.zip")
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	defer os.Remove(fileName)
	unzip(fileName, a.songBookDir)
	return nil
}

func (a *App) DownloadInternal() error {

	slog.Info(fmt.Sprintf("Processing from %s://%s", a.pdfFiles.UrlScheme, a.pdfFiles.Domain))

	if err := os.MkdirAll(a.appDir, os.ModePerm); err != nil {
		return err
	}

	fileName, err := a.downloadFile(a.pdfFiles.Url, "INDEX")
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	a.pdfFiles.Items = a.parseHtml(fileName)
	if !a.testRun {
		a.downloadParts()
	}

	a.serializeToYaml("data.yaml", a.pdfFiles)
	return nil
}

func (a *App) DownloadEz() error {

	var err error
	// if !a.status.WebResourcesReady {
	// 	err = a.DownloadInternal()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	a.status.WebResourcesReady = true
	// 	a.saveStatus()
	// }

	if !a.status.SongsReady && !a.testRun {
		err = a.DownloadSongBase()
		if err != nil {
			return err

		}
		a.status.DatabaseReady = false
		a.status.SongsReady = true
		a.saveStatus()
	}

	if !a.status.DatabaseReady {
		a.PrepareDatabase()
		a.FillDatabase()
		a.status.DatabaseReady = true
		a.saveStatus()
	}

	return nil
}
