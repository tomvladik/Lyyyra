package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// App struct
type App struct {
	ctx         context.Context
	appDir      string
	pdfFiles    SongFilesSources
	xmlUrl      string
	dbFilePath  string
	songBookDir string
	urlDomain   string
}

// NewApp creates a new App application struct
func NewApp() *App {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
		return &App{}
	}

	xmlUrl := "https://www.evangelickyzpevnik.cz/res/archive/001/000243.zip"
	pdfUrl := "https://www.evangelickyzpevnik.cz/zpevnik/kapitoly-a-pisne/"
	// Parse the URL
	parsedURL, err := url.Parse(pdfUrl)
	if err != nil {
		log.Fatal(err)
		return &App{}
	}
	appDir := filepath.Join(homeDir, "Lyyyra", strings.Replace(parsedURL.Host, ":", "_", -1))
	return &App{
		appDir:      appDir,
		pdfFiles:    SongFilesSources{Domain: parsedURL.Host, Url: pdfUrl, UrlScheme: parsedURL.Scheme, Items: []FileItem{}},
		xmlUrl:      xmlUrl,
		dbFilePath:  filepath.Join(appDir, "Songs.db"),
		songBookDir: filepath.Join(appDir, "SongBook"),
		urlDomain:   parsedURL.Host,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}
