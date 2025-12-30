package main

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupTestApp(t *testing.T) *App {
	// Create a temporary directory for the app
	appDir, err := ioutil.TempDir("", "testapp_")
	if err != nil {
		t.Fatalf("Failed to create temp app directory: %v", err)
	}

	// Initialize the App with the test directory
	app := &App{
		appDir:      appDir,
		pdfDir:      filepath.Join(appDir, "PdfSources"),
		songBookDir: filepath.Join(appDir, "songbooks"),
		status:      AppStatus{WebResourcesReady: true},
	}

	return app
}

func teardownTestApp(app *App) {
	// Remove the temporary app directory
	os.RemoveAll(app.appDir)
}

func TestDownloadFile(t *testing.T) {
	app := setupTestApp(t)
	defer teardownTestApp(app)

	// Create a test server to serve a test file
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test content"))
	}))
	defer ts.Close()

	fileUrl := ts.URL
	fileName := "testfile.txt"

	// Test downloading the file
	downloadedFilePath, err := app.downloadFile(fileUrl, fileName)
	if err != nil {
		t.Fatalf("Failed to download file: %v", err)
	}

	// Check if the file was downloaded correctly
	content, err := ioutil.ReadFile(downloadedFilePath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	expectedContent := "test content"
	if string(content) != expectedContent {
		t.Errorf("Expected file content %q, got %q", expectedContent, string(content))
	}
}

func createZipFromFiles(files map[string][]byte) []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	for name, content := range files {
		f, err := zipWriter.Create(name)
		if err != nil {
			panic(err)
		}
		if _, err := f.Write(content); err != nil {
			panic(err)
		}
	}

	zipWriter.Close()
	return buf.Bytes()
}

func TestDownloadSongBase(t *testing.T) {
	app := setupTestApp(t)
	defer teardownTestApp(app)

	// Create a test server to serve a test zip file
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write(createZipFromFiles(map[string][]byte{"empty.txt": []byte("")}))
	}))
	defer ts.Close()

	app.xmlUrl = ts.URL

	// Test downloading the song base
	err := app.DownloadSongBase()
	if err != nil {
		t.Fatalf("Failed to download song base: %v", err)
	}

	// Check if the zip file was downloaded and unzipped correctly
	_, err = os.Stat(filepath.Join(app.songBookDir, "empty.txt"))
	if os.IsNotExist(err) {
		t.Errorf("Expected zip file to be unzipped, but it was not found")
	}
}

func TestDownloadEz(t *testing.T) {
	app := setupTestApp(t)
	app.dbFilePath = filepath.Join(app.appDir, "Songs.db")
	app.songBookDir = filepath.Join(app.appDir, "SongBook")
	app.status.WebResourcesReady = true

	content, err := os.ReadFile(filepath.Join("testdata", "song-1.xml"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}
	zipBytes := createZipFromFiles(map[string][]byte{"song-1.xml": content})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write(zipBytes)
	}))
	defer ts.Close()
	app.xmlUrl = ts.URL

	// Mock the status
	app.status.SongsReady = false
	app.status.DatabaseReady = false

	// Test downloading EZ resources
	err = app.DownloadEz()
	if err != nil {
		t.Fatalf("Failed to download EZ resources: %v", err)
	}

	// Check if the song base was downloaded and the database was prepared
	if !app.status.SongsReady {
		t.Errorf("Expected SongsReady to be true, got false")
	}
	if !app.status.DatabaseReady {
		t.Errorf("Expected DatabaseReady to be true, got false")
	}
}

func TestDownloadEzRemote(t *testing.T) {
	if os.Getenv("RUN_REMOTE_INTEGRATION_TESTS") == "" {
		t.Skip("skipping remote integration test; set RUN_REMOTE_INTEGRATION_TESTS=1 to enable")
	}

	app := setupTestApp(t)
	defer teardownTestApp(app)
	app.dbFilePath = filepath.Join(app.appDir, "Songs.db")
	app.songBookDir = filepath.Join(app.appDir, "SongBook")
	app.status.WebResourcesReady = true
	app.xmlUrl = XMLUrl
	app.status.SongsReady = false
	app.status.DatabaseReady = false

	if err := app.DownloadEz(); err != nil {
		t.Fatalf("Failed to download EZ resources remotely: %v", err)
	}

	if !app.status.SongsReady {
		t.Errorf("Expected SongsReady to be true, got false")
	}
	if !app.status.DatabaseReady {
		t.Errorf("Expected DatabaseReady to be true, got false")
	}
	if _, err := os.Stat(app.dbFilePath); err != nil {
		t.Errorf("Expected database file to exist: %v", err)
	}
	if _, err := os.Stat(app.songBookDir); err != nil {
		t.Errorf("Expected SongBook directory to exist: %v", err)
	}
}

func TestDownloadSupplementalPDFs(t *testing.T) {
	app := setupTestApp(t)
	defer teardownTestApp(app)
	app.status.WebResourcesReady = false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pdf"))
	}))
	defer server.Close()

	original := SupplementalPDFs
	SupplementalPDFs = []SupplementalPDF{{
		URL:      server.URL + "/test.pdf",
		FileName: "choralnik.pdf",
	}}
	t.Cleanup(func() {
		SupplementalPDFs = original
	})

	if err := app.downloadSupplementalPDFs(); err != nil {
		t.Fatalf("Failed to download supplemental pdfs: %v", err)
	}

	if _, err := os.Stat(filepath.Join(app.pdfDir, "choralnik.pdf")); err != nil {
		t.Fatalf("Expected pdf to be downloaded: %v", err)
	}
}

func TestDownloadEzTriggersSupplementalDownloads(t *testing.T) {
	app := setupTestApp(t)
	defer teardownTestApp(app)
	app.status.WebResourcesReady = false
	app.status.SongsReady = true
	app.status.DatabaseReady = true

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pdf"))
	}))
	defer server.Close()

	original := SupplementalPDFs
	SupplementalPDFs = []SupplementalPDF{{
		URL:      server.URL + "/async.pdf",
		FileName: "kytara.pdf",
	}}
	t.Cleanup(func() {
		SupplementalPDFs = original
	})

	if err := app.DownloadEz(); err != nil {
		t.Fatalf("DownloadEz should finish even when supplemental PDFs download concurrently: %v", err)
	}

	if _, err := os.Stat(filepath.Join(app.pdfDir, "kytara.pdf")); err != nil {
		t.Fatalf("Expected supplemental pdf to exist after DownloadEz: %v", err)
	}
	if !app.status.WebResourcesReady {
		t.Fatalf("Expected WebResourcesReady to be true")
	}
}
