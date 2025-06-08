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
		songBookDir: filepath.Join(appDir, "songbooks"),
		status:      AppStatus{},
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

func createMinimalZip() []byte {
	// Create a buffer to hold the ZIP file data
	buf := new(bytes.Buffer)

	// Create a new ZIP archive writer
	zipWriter := zip.NewWriter(buf)

	// Add an empty file to the ZIP archive
	_, err := zipWriter.Create("empty.txt")
	if err != nil {
		panic(err) // Handle error properly in real code
	}

	// Close the ZIP writer to finalize the archive
	zipWriter.Close()

	return buf.Bytes()
}

func TestDownloadSongBase(t *testing.T) {
	app := setupTestApp(t)
	defer teardownTestApp(app)

	// Create a test server to serve a test zip file
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write(createMinimalZip())
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
	//homeDir, _ := os.UserHomeDir()

	//app.appDir = filepath.Join(homeDir, "Lyyyra", "www.evangelickyzpevnik.cz")
	app.dbFilePath = filepath.Join(app.appDir, "Songs.db")
	app.songBookDir = filepath.Join(app.appDir, "SongBook")
	app.xmlUrl = XMLUrl
	// defer teardownTestApp(app)

	// Mock the status
	app.status.SongsReady = false
	app.status.DatabaseReady = false

	// Test downloading EZ resources
	err := app.DownloadEz()
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
