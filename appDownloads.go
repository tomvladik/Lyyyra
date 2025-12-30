package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

func (a *App) downloadFile(fileUrl, fileName string) (string, error) {
	// Create or truncate the file in the app directory
	fullPath := filepath.Join(a.appDir, fileName)
	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		return "", err
	}
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

func (a *App) downloadSupplementalPDFs() error {
	if len(SupplementalPDFs) == 0 || a.testRun {
		return nil
	}

	storageDir := a.pdfDir
	if storageDir == "" {
		storageDir = filepath.Join(a.appDir, "PdfSources")
	}

	if err := os.MkdirAll(storageDir, os.ModePerm); err != nil {
		return err
	}

	for _, pdf := range SupplementalPDFs {
		fileName := pdf.FileName
		if fileName == "" {
			fileName = path.Base(pdf.URL)
		}

		targetPath := filepath.Join(storageDir, fileName)
		if _, err := os.Stat(targetPath); err == nil {
			slog.Info("Supplemental PDF already present", "file", targetPath)
			continue
		}

		relativePath, err := filepath.Rel(a.appDir, targetPath)
		if err != nil {
			relativePath = fileName
		}

		if _, err := a.downloadFile(pdf.URL, relativePath); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) startSupplementalDownload() <-chan error {
	if a.status.WebResourcesReady || len(SupplementalPDFs) == 0 {
		return nil
	}

	a.supplementalMu.Lock()
	defer a.supplementalMu.Unlock()
	if a.supplementalErrCh != nil {
		return a.supplementalErrCh
	}

	ch := make(chan error, 1)
	a.supplementalErrCh = ch

	go func() {
		err := a.downloadSupplementalPDFs()
		ch <- err
		close(ch)

		a.supplementalMu.Lock()
		a.supplementalErrCh = nil
		a.supplementalMu.Unlock()
	}()

	return ch
}

func (a *App) ensureBackgroundSupplementalDownload() {
	if a.testRun {
		return
	}
	ch := a.startSupplementalDownload()
	if ch == nil {
		return
	}

	go func() {
		if err := <-ch; err != nil {
			slog.Warn("Background supplemental PDF download failed", "error", err)
			return
		}
		a.status.WebResourcesReady = true
		a.saveStatus()
	}()
}

func (a *App) DownloadSongBase() error {
	if err := os.MkdirAll(a.songBookDir, os.ModePerm); err != nil {
		return err
	}
	a.updateProgress("Stahuji XML soubory...", 0)

	fileName, err := a.downloadFile(a.xmlUrl, "Songs.zip")
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	defer os.Remove(fileName)

	a.updateProgress("Rozbaluji soubory...", 50)

	if err := unzip(fileName, a.songBookDir); err != nil {
		slog.Error("Failed to unzip song base", "error", err)
		return err
	}
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
	// Set progress flag at the start
	a.startProgress("Zahajuji přípravu dat...")

	supplementalErrCh := a.startSupplementalDownload()

	if !a.status.SongsReady && !a.testRun {
		err = a.DownloadSongBase()
		if err != nil {
			a.clearProgress()
			return err

		}
		a.status.DatabaseReady = false
		a.status.SongsReady = true
		a.saveStatus()
	}

	if !a.status.DatabaseReady {
		a.updateProgress("Připravuji databázi...", 0)

		a.PrepareDatabase()
		a.FillDatabase()
		a.status.DatabaseReady = true
		a.saveStatus()
	}

	if supplementalErrCh != nil {
		if err := <-supplementalErrCh; err != nil {
			a.clearProgress()
			return err
		}
		a.status.WebResourcesReady = true
		a.saveStatus()

		// Process kytara.pdf after download
		a.updateProgress("Zpracovávám PDF soubory...", 90)
		if err := a.ProcessKytaraPDF(); err != nil {
			slog.Warn("Error processing kytara.pdf", "error", err)
			// Don't fail the whole process if PDF splitting fails
		}
	}

	// Clear progress flag at the end
	a.clearProgress()

	return nil
}
