package main

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	status      AppStatus
	logFile     *os.File
}

// NewApp creates a new App application struct
func NewApp() *App {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("Can't open home dir", err)
		return &App{}
	}

	// Open or create a log file
	logFile, err := os.OpenFile(filepath.Join(homeDir, "Lyyyra", "app.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error(err.Error())
	}
	// Set the output of the log package to the log file
	logOptions := slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewTextHandler(logFile, &logOptions))
	slog.SetDefault(logger)

	slog.Info("=========================================================================================")

	xmlUrl := "https://www.evangelickyzpevnik.cz/res/archive/001/000243.zip"
	pdfUrl := "https://www.evangelickyzpevnik.cz/zpevnik/kapitoly-a-pisne/"
	// Parse the URL
	parsedURL, err := url.Parse(pdfUrl)
	if err != nil {
		slog.Error(err.Error())
		return &App{}
	}
	appDir := filepath.Join(homeDir, "Lyyyra", strings.Replace(parsedURL.Host, ":", "_", -1))

	app := App{
		appDir:      appDir,
		pdfFiles:    SongFilesSources{Domain: parsedURL.Host, Url: pdfUrl, UrlScheme: parsedURL.Scheme, Items: []FileItem{}},
		xmlUrl:      xmlUrl,
		dbFilePath:  filepath.Join(appDir, "Songs.db"),
		songBookDir: filepath.Join(appDir, "SongBook"),
		urlDomain:   parsedURL.Host,
		logFile:     logFile,
	}
	err = app.deserializeFromYaml(&app.status, "status.yaml")
	if err != nil {
		app.status = AppStatus{}
		app.saveStatus()
	}

	return &app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) saveStatus() {
	a.status.LastSave = time.Now()
	a.serializeToYaml("status.yaml", &a.status)
}
