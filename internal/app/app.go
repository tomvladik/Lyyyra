package app

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

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
		slog.Error("Failed to get user home directory", "error", err)
		return &App{}
	}

	path := filepath.Join(homeDir, "Lyyyra")
	// Create directories along the path
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		slog.Error("Failed to create app directory", "path", path, "error", err)
		return &App{}
	}

	// Open or create a log file
	logFile, err := os.OpenFile(filepath.Join(homeDir, "Lyyyra", "app.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("Failed to open log file", "error", err)
		// Continue without file logging, use stderr
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
		slog.Error("Failed to parse PDF URL", "url", pdfUrl, "error", err)
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

// Shutdown flushes status to disk and closes the log file.
func (a *App) Shutdown() {
	if a == nil {
		return
	}
	a.saveStatus()
	if a.logFile != nil {
		_ = a.logFile.Close()
	}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize database schema and apply migrations on every start
	a.InitializeDatabase()

	// On Windows, ensure MuPDF DLLs are available for PDF cropping
	if goruntime.GOOS == "windows" {
		a.initializeWindowsDLLs()
	}
}

// initializeWindowsDLLs attempts to set up MuPDF DLLs on Windows
func (a *App) initializeWindowsDLLs() {
	// Try to ensure DLLs are available
	// This will work if they're embedded, in PATH, or in current directory
	slog.Info("Initializing MuPDF libraries for Windows...")
	// Note: Actual DLL extraction happens in embedded package when pdf-crop is used
	// Here we just log the status
	slog.Info("PDF cropping will use system MuPDF if available, or gracefully disable if not")
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

// ResetData deletes all stored app data and re-initializes songs and database.
func (a *App) ResetData() error {
	// Show progress in UI
	a.startProgress("Mažu uložená data...")
	defer a.clearProgress()

	// Force close any database connections by opening and immediately closing
	// This ensures SQLite releases any file locks on Windows
	_ = a.withDB(func(db *sql.DB) error {
		return nil
	})

	// Remove application data directory
	if err := os.RemoveAll(a.appDir); err != nil {
		slog.Error("Failed to remove app directory", "path", a.appDir, "error", err)
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

	// Reset status to force re-download
	a.status = AppStatus{
		Sorting:           Title,
		BuildVersion:      buildVersion,
		WebResourcesReady: false,
		SongsReady:        false,
		DatabaseReady:     false,
	}
	a.saveStatus()

	// Reinitialize database schema after deletion
	a.InitializeDatabase()

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
