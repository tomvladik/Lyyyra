package app

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

// createZipWithDir builds a zip archive containing a directory entry and a file.
func createZipWithDir(t *testing.T, zipPath string) {
	t.Helper()
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// Add a directory entry
	dirHeader := &zip.FileHeader{Name: "subdir/", Method: zip.Store}
	dirHeader.SetMode(0755 | os.ModeDir)
	if _, err := w.CreateHeader(dirHeader); err != nil {
		t.Fatalf("createZipWithDir dir: %v", err)
	}

	// Add a file inside the directory
	f, err := w.Create("subdir/hello.txt")
	if err != nil {
		t.Fatalf("createZipWithDir file: %v", err)
	}
	if _, err := f.Write([]byte("hello from zip")); err != nil {
		t.Fatalf("createZipWithDir write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("createZipWithDir close: %v", err)
	}
	if err := os.WriteFile(zipPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("createZipWithDir writefile: %v", err)
	}
}

func TestUnzip_WithDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")
	createZipWithDir(t, zipPath)

	dest := filepath.Join(tmpDir, "out")
	if err := unzip(zipPath, dest); err != nil {
		t.Fatalf("unzip with directory: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dest, "subdir")); os.IsNotExist(err) {
		t.Error("expected subdir to be created")
	}
	content, err := os.ReadFile(filepath.Join(dest, "subdir", "hello.txt"))
	if err != nil {
		t.Fatalf("expected hello.txt to exist: %v", err)
	}
	if string(content) != "hello from zip" {
		t.Errorf("unexpected content: %q", string(content))
	}
}

func TestUnzip_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for i := 0; i < 3; i++ {
		f, _ := w.Create(fmt.Sprintf("song-%d.xml", i))
		f.Write([]byte(fmt.Sprintf("<song>%d</song>", i))) //nolint
	}
	w.Close()
	zipPath := filepath.Join(tmpDir, "songs.zip")
	if err := os.WriteFile(zipPath, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(tmpDir, "extracted")
	if err := unzip(zipPath, dest); err != nil {
		t.Fatalf("unzip multiple files: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, err := os.Stat(filepath.Join(dest, fmt.Sprintf("song-%d.xml", i))); os.IsNotExist(err) {
			t.Errorf("song-%d.xml not extracted", i)
		}
	}
}

// ─── downloadFile ─────────────────────────────────────────────────────────────

func TestDownloadFile_FileScheme(t *testing.T) {
	tmpDir := t.TempDir()
	app := &App{appDir: tmpDir}

	// Create a source file
	srcContent := []byte("content from file://")
	srcPath := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcPath, srcContent, 0644); err != nil {
		t.Fatal(err)
	}

	fileURL := fmt.Sprintf("file://%s", srcPath)
	dest, err := app.downloadFile(fileURL, "dest.txt")
	if err != nil {
		t.Fatalf("downloadFile file:// failed: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(srcContent) {
		t.Errorf("content mismatch: got %q, want %q", got, srcContent)
	}
}

func TestDownloadFile_HTTPNon200(t *testing.T) {
	tmpDir := t.TempDir()
	app := &App{appDir: tmpDir}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	_, err := app.downloadFile(ts.URL, "file.txt")
	if err == nil {
		t.Error("expected error for HTTP 404")
	}
}

func TestDownloadFile_InvalidURL(t *testing.T) {
	tmpDir := t.TempDir()
	app := &App{appDir: tmpDir}

	_, err := app.downloadFile("://bad-url", "file.txt")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestDownloadFile_ConnectionRefused(t *testing.T) {
	tmpDir := t.TempDir()
	app := &App{appDir: tmpDir}

	// Use a port where nothing is listening
	_, err := app.downloadFile("http://127.0.0.1:19999/x", "file.txt")
	if err == nil {
		t.Error("expected error for refused connection")
	}
}

// ─── downloadParts ────────────────────────────────────────────────────────────

func TestDownloadParts_WithItems(t *testing.T) {
	tmpDir := t.TempDir()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pdf-content")) //nolint
	}))
	defer ts.Close()

	parsedURL, _ := url.Parse(ts.URL)
	app := &App{
		appDir: tmpDir,
		pdfFiles: SongFilesSources{
			UrlScheme: parsedURL.Scheme,
			Domain:    parsedURL.Host,
			Items: []FileItem{
				{Href: "/song1.pdf", Title: "Song1"},
			},
		},
	}

	// Should not panic; items might fail to download if URL is wrong but no crash
	app.downloadParts()
}

// ─── DownloadInternal ─────────────────────────────────────────────────────────

func TestDownloadInternal_TestRunMode(t *testing.T) {
	tmpDir := t.TempDir()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve a minimal but parseable HTML page (XPath won't match, returns empty list)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><p>no songs here</p></body></html>`)) //nolint
	}))
	defer ts.Close()

	parsedURL, _ := url.Parse(ts.URL)
	app := &App{
		appDir:  tmpDir,
		testRun: true,
		pdfFiles: SongFilesSources{
			Url:       ts.URL,
			UrlScheme: parsedURL.Scheme,
			Domain:    parsedURL.Host,
			Items:     []FileItem{},
		},
	}

	err := app.DownloadInternal()
	if err != nil {
		t.Fatalf("DownloadInternal testRun: %v", err)
	}
}

// ─── startSupplementalDownload / ensureBackgroundSupplementalDownload ─────────

func TestStartSupplementalDownload_WebResourcesReady(t *testing.T) {
	app := &App{status: AppStatus{WebResourcesReady: true}}
	ch := app.startSupplementalDownload()
	if ch != nil {
		t.Error("expected nil when WebResourcesReady=true")
	}
}

func TestStartSupplementalDownload_EmptySupplementalPDFs(t *testing.T) {
	original := SupplementalPDFs
	SupplementalPDFs = []SupplementalPDF{}
	t.Cleanup(func() { SupplementalPDFs = original })

	app := &App{}
	ch := app.startSupplementalDownload()
	if ch != nil {
		t.Error("expected nil when SupplementalPDFs is empty")
	}
}

func TestStartSupplementalDownload_AlreadyRunning(t *testing.T) {
	existing := make(chan error, 1)
	app := &App{
		supplementalErrCh: existing,
	}

	original := SupplementalPDFs
	SupplementalPDFs = []SupplementalPDF{{URL: "http://x", FileName: "x.pdf"}}
	t.Cleanup(func() { SupplementalPDFs = original })

	ch := app.startSupplementalDownload()
	if ch != existing {
		t.Error("expected the existing channel to be returned")
	}
}

func TestEnsureBackgroundSupplementalDownload_TestRun(t *testing.T) {
	app := &App{testRun: true}
	// Should return immediately without starting goroutine
	app.ensureBackgroundSupplementalDownload()
}

func TestEnsureBackgroundSupplementalDownload_NilChannel(t *testing.T) {
	// WebResourcesReady=true means startSupplementalDownload returns nil
	app := &App{status: AppStatus{WebResourcesReady: true}}
	app.ensureBackgroundSupplementalDownload() // should not panic
}
