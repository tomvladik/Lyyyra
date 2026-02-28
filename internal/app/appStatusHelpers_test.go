package app

import (
	"os"
	"path/filepath"
	"testing"
)

// setupStatusApp creates an App with a temp dir, populated DB, and testRun=true.
func setupStatusApp(t *testing.T) *App {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	app := &App{
		appDir:      tmpDir,
		dbFilePath:  dbPath,
		songBookDir: filepath.Join(tmpDir, "SongBook"),
		pdfDir:      filepath.Join(tmpDir, "PdfSources"),
		testRun:     true,
	}
	return app
}

// makeXMLFile writes a dummy .xml file into the given dir.
func makeXMLFile(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("makeXMLFile MkdirAll: %v", err)
	}
	content := []byte(`<?xml version="1.0"?><song><title>test</title></song>`)
	if err := os.WriteFile(filepath.Join(dir, name), content, 0644); err != nil {
		t.Fatalf("makeXMLFile WriteFile: %v", err)
	}
}

// ─── hasDownloadedSongs ───────────────────────────────────────────────────────

func TestHasDownloadedSongs_EmptyDir(t *testing.T) {
	app := setupStatusApp(t)
	if err := os.MkdirAll(app.songBookDir, 0755); err != nil {
		t.Fatal(err)
	}
	if app.hasDownloadedSongs() {
		t.Error("expected false when dir is empty")
	}
}

func TestHasDownloadedSongs_MissingDir(t *testing.T) {
	app := setupStatusApp(t)
	// Don't create songBookDir
	if app.hasDownloadedSongs() {
		t.Error("expected false when dir is missing")
	}
}

func TestHasDownloadedSongs_WithXMLFile(t *testing.T) {
	app := setupStatusApp(t)
	makeXMLFile(t, app.songBookDir, "song.xml")
	if !app.hasDownloadedSongs() {
		t.Error("expected true when at least one .xml file exists (testRun mode)")
	}
}

func TestHasDownloadedSongs_IgnoresDirectories(t *testing.T) {
	app := setupStatusApp(t)
	if err := os.MkdirAll(app.songBookDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create a subdirectory only (no XML files)
	if err := os.MkdirAll(filepath.Join(app.songBookDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if app.hasDownloadedSongs() {
		t.Error("expected false when dir contains only subdirectories")
	}
}

func TestHasDownloadedSongs_IgnoresNonXMLFiles(t *testing.T) {
	app := setupStatusApp(t)
	if err := os.MkdirAll(app.songBookDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create a non-XML file
	if err := os.WriteFile(filepath.Join(app.songBookDir, "readme.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if app.hasDownloadedSongs() {
		t.Error("expected false when dir contains only non-XML files")
	}
}

// ─── hasDatabaseContent ──────────────────────────────────────────────────────

func TestHasDatabaseContent_NoFile(t *testing.T) {
	app := setupStatusApp(t)
	if app.hasDatabaseContent() {
		t.Error("expected false when database file does not exist")
	}
}

func TestHasDatabaseContent_EmptyDatabase(t *testing.T) {
	app := setupStatusApp(t)
	app.InitializeDatabase()
	// DB exists but has no songs
	if app.hasDatabaseContent() {
		t.Error("expected false when database has no songs")
	}
}

func TestHasDatabaseContent_WithSongs(t *testing.T) {
	app := setupStatusApp(t)
	app.InitializeDatabase()

	// Populate the db via filling EZ songs
	ezDir := filepath.Join(app.songBookDir, "EZ")
	if err := os.MkdirAll(ezDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := copyDir("testdata/", ezDir); err != nil {
		t.Fatalf("copyDir: %v", err)
	}
	app.FillDatabase()

	if !app.hasDatabaseContent() {
		t.Error("expected true after filling the database")
	}
}

// ─── hasPdfSources ───────────────────────────────────────────────────────────

func TestHasPdfSources_NoPdfDir(t *testing.T) {
	app := setupStatusApp(t)
	// pdfDir is set but doesn't have the files
	if app.hasPdfSources() {
		t.Error("expected false when supplemental PDFs are missing")
	}
}

func TestHasPdfSources_AllPDFsPresent(t *testing.T) {
	app := setupTestApp(t)
	if err := os.MkdirAll(app.pdfDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create dummy PDFs matching every entry in SupplementalPDFs
	for _, pdf := range SupplementalPDFs {
		name := pdf.FileName
		if name == "" {
			name = filepath.Base(pdf.URL)
		}
		if err := os.WriteFile(filepath.Join(app.pdfDir, name), []byte("%PDF-1.4"), 0644); err != nil {
			t.Fatalf("failed to write dummy PDF %s: %v", name, err)
		}
	}
	if !app.hasPdfSources() {
		t.Error("expected true when all supplemental PDFs are present")
	}
}

func TestHasPdfSources_MissingOnePDF(t *testing.T) {
	if len(SupplementalPDFs) < 2 {
		t.Skip("need at least 2 supplemental PDFs for this test")
	}
	app := setupStatusApp(t)
	if err := os.MkdirAll(app.pdfDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Only write the first PDF, not all of them
	pdf := SupplementalPDFs[0]
	name := pdf.FileName
	if name == "" {
		name = filepath.Base(pdf.URL)
	}
	if err := os.WriteFile(filepath.Join(app.pdfDir, name), []byte("%PDF-1.4"), 0644); err != nil {
		t.Fatal(err)
	}
	if app.hasPdfSources() {
		t.Error("expected false when at least one supplemental PDF is missing")
	}
}

func TestHasPdfSources_EmptyPdfDir(t *testing.T) {
	app := setupTestApp(t) // pdfDir not created
	app.pdfDir = ""       // explicitly empty
	if app.hasPdfSources() {
		t.Error("expected false when pdfDir is empty string")
	}
}

// ─── reconcileStoredStatus ───────────────────────────────────────────────────

func TestReconcileStoredStatus_UpdatesFlags(t *testing.T) {
	app := setupStatusApp(t)
	app.InitializeDatabase()

	// status.SongsReady starts false, no XML files exist
	app.status = AppStatus{SongsReady: false, DatabaseReady: false}

	app.reconcileStoredStatus()

	// Both should still be false (no files, no songs)
	if app.status.SongsReady {
		t.Error("expected SongsReady false with no XML files")
	}
	if app.status.DatabaseReady {
		t.Error("expected DatabaseReady false with empty DB")
	}
}

func TestReconcileStoredStatus_ResetsInvalidSorting(t *testing.T) {
	app := setupStatusApp(t)
	app.InitializeDatabase()
	app.status = AppStatus{Sorting: "invalid-option"}

	app.reconcileStoredStatus()

	if app.status.Sorting != Entry {
		t.Errorf("expected sorting reset to %q, got %q", Entry, app.status.Sorting)
	}
}

func TestReconcileStoredStatus_PreservesValidSorting(t *testing.T) {
	app := setupStatusApp(t)
	app.InitializeDatabase()
	app.status = AppStatus{Sorting: Title}

	app.reconcileStoredStatus()

	if app.status.Sorting != Title {
		t.Errorf("expected sorting preserved as %q, got %q", Title, app.status.Sorting)
	}
}

func TestReconcileStoredStatus_SongsReadyAfterFill(t *testing.T) {
	app := setupStatusApp(t)
	app.InitializeDatabase()

	ezDir := filepath.Join(app.songBookDir, "EZ")
	if err := os.MkdirAll(ezDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := copyDir("testdata/", ezDir); err != nil {
		t.Fatalf("copyDir: %v", err)
	}
	app.FillDatabase()

	// Seed the songBookDir with a dummy XML file to simulate downloaded songs
	makeXMLFile(t, app.songBookDir, "dummy.xml")

	app.status = AppStatus{}
	app.reconcileStoredStatus()

	if !app.status.SongsReady {
		t.Error("expected SongsReady true after placing XML files")
	}
	if !app.status.DatabaseReady {
		t.Error("expected DatabaseReady true after FillDatabase")
	}
}
