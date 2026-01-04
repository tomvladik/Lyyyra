package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/oliverpool/unipdf/v3/creator"
	"github.com/oliverpool/unipdf/v3/model"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx         context.Context
	appDir      string
	pdfFiles    SongFilesSources
	pdfDir      string
	xmlUrl      string
	dbFilePath  string
	songBookDir string
	urlDomain   string
	status      AppStatus
	logFile     *os.File
	testRun     bool
	// supplemental download coordination
	supplementalMu    sync.Mutex
	supplementalErrCh chan error
}

// NewApp creates a new App application struct
func NewApp() *App {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error(err.Error())
		return &App{}
	}

	path := filepath.Join(homeDir, "Lyyyra")
	// Create directories along the path
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return &App{}
	}

	// Open or create a log file
	logFile, err := os.OpenFile(filepath.Join(homeDir, "Lyyyra", "app.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error(err.Error())
	}
	// Set the output of the log package to the log file
	logOptions := slog.HandlerOptions{Level: slog.LevelInfo}
	logger := slog.New(slog.NewTextHandler(logFile, &logOptions))
	slog.SetDefault(logger)

	slog.Info("=========================================================================================")

	xmlUrl := XMLUrl
	pdfUrl := PDFUrl
	// Parse the URL
	parsedURL, err := url.Parse(pdfUrl)
	if err != nil {
		slog.Error(err.Error())
		return &App{}
	}
	appDir := filepath.Join(homeDir, "Lyyyra", strings.Replace(parsedURL.Host, ":", "_", -1))

	app := App{
		appDir:      appDir,
		pdfDir:      filepath.Join(appDir, "PdfSources"),
		pdfFiles:    SongFilesSources{Domain: parsedURL.Host, Url: pdfUrl, UrlScheme: parsedURL.Scheme, Items: []FileItem{}},
		xmlUrl:      xmlUrl,
		dbFilePath:  filepath.Join(appDir, "Songs.db"),
		songBookDir: filepath.Join(appDir, "SongBook"),
		urlDomain:   parsedURL.Host,
		logFile:     logFile,
	}

	if err := os.MkdirAll(app.songBookDir, os.ModePerm); err != nil {
		slog.Error(fmt.Sprintf("Failed to create directories %s: %v", app.songBookDir, err))
		return &App{}
	}
	if err := os.MkdirAll(app.pdfDir, os.ModePerm); err != nil {
		slog.Error(fmt.Sprintf("Failed to create directories %s: %v", app.pdfDir, err))
		return &App{}
	}
	err = app.deserializeFromYaml(&app.status, "status.yaml")
	if err != nil {
		app.status = AppStatus{Sorting: Title}
	}
	app.status.BuildVersion = buildVersion
	app.reconcileStoredStatus()
	slog.Info(fmt.Sprintf("Status DatabaseReady: %+v", app.status))
	//app.testRun = true
	return &app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetPdfFile returns the PDF contents encoded as a data URL
// so the frontend can display it without violating file:// restrictions.
func (a *App) GetPdfFile(filename string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("no filename provided")
	}

	filePath := filepath.Join(a.pdfDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", filename)
	}

	// Return absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return "data:application/pdf;base64," + encoded, nil
}

// GetCombinedPdf merges the provided PDF files (in order) into a single PDF
// and returns it as a base64 data URL
func (a *App) GetCombinedPdf(filenames []string) (string, error) {
	if len(filenames) == 0 {
		return "", fmt.Errorf("no filenames provided")
	}

	c := creator.New()
	pagesAdded := 0

	for _, name := range filenames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		filePath := filepath.Join(a.pdfDir, name)
		file, err := os.Open(filePath)
		if err != nil {
			return "", fmt.Errorf("unable to open %s: %w", name, err)
		}

		reader, err := model.NewPdfReader(file)
		if err != nil {
			file.Close()
			return "", fmt.Errorf("unable to read %s: %w", name, err)
		}

		numPages, err := reader.GetNumPages()
		if err != nil {
			file.Close()
			return "", fmt.Errorf("unable to get page count for %s: %w", name, err)
		}

		for pageNum := 1; pageNum <= numPages; pageNum++ {
			page, err := reader.GetPage(pageNum)
			if err != nil {
				file.Close()
				return "", fmt.Errorf("unable to read page %d from %s: %w", pageNum, name, err)
			}
			if err := c.AddPage(page); err != nil {
				file.Close()
				return "", fmt.Errorf("unable to add page %d from %s: %w", pageNum, name, err)
			}
			pagesAdded++
		}

		file.Close()
	}

	if pagesAdded == 0 {
		return "", fmt.Errorf("no pages added to compilation")
	}

	buf := &bytes.Buffer{}
	if err := c.Write(buf); err != nil {
		return "", fmt.Errorf("unable to write compiled PDF: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:application/pdf;base64," + encoded, nil
}

// ResetData deletes all stored app data and re-initializes songs and database.
func (a *App) ResetData() error {
	// Show progress in UI
	a.startProgress("Mažu uložená data...")
	defer a.clearProgress()

	// Remove application data directory
	if err := os.RemoveAll(a.appDir); err != nil {
		return err
	}

	// Recreate required directories
	if err := os.MkdirAll(a.appDir, os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(a.songBookDir, os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(a.pdfDir, os.ModePerm); err != nil {
		return err
	}

	// Reset status
	a.status = AppStatus{Sorting: Title, BuildVersion: buildVersion}
	a.saveStatus()

	// Re-run full download and DB preparation
	a.updateProgress("Znovu stahuji a připravuji data...", 0)
	return a.DownloadEz()
}

func (a *App) saveStatus() {
	a.status.LastSave = time.Now().UTC().Format(time.RFC3339)
	a.serializeToYaml("status.yaml", &a.status)
}

func (a *App) SaveSorting(sorting SortingOption) {
	normalized := normalizeSortingOption(string(sorting))
	if normalized != sorting {
		slog.Warn("Received unsupported sorting option, defaulting to entry", "requested", sorting, "normalized", normalized)
	}
	if normalized != a.status.Sorting {
		slog.Info(fmt.Sprintf("Sorting changed from %s to %s", a.status.Sorting, normalized))
		a.status.Sorting = normalized
		a.saveStatus()
	}
}

func (a *App) GetStatus() AppStatus {
	a.status.BuildVersion = buildVersion
	return a.status
}

// Projection control methods that emit events
func (a *App) ProjectionNextVerse() {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "projection:nextVerse")
	}
}

func (a *App) ProjectionPrevVerse() {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "projection:prevVerse")
	}
}

func (a *App) ProjectionNextSong() {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "projection:nextSong")
	}
}

func (a *App) ProjectionPrevSong() {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "projection:prevSong")
	}
}
