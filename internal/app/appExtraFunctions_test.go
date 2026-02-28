package app

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// ─── serializeToYaml / deserializeFromYaml ───────────────────────────────────

func TestSerializeDeserializeYaml_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	app := &App{appDir: tmpDir}

	original := AppStatus{
		DatabaseReady: true,
		SongsReady:    true,
		Sorting:       Title,
		SearchPattern: "Alleluia",
		BuildVersion:  "1.2.3",
	}

	app.serializeToYaml("status.yaml", &original)

	var loaded AppStatus
	if err := app.deserializeFromYaml(&loaded, "status.yaml"); err != nil {
		t.Fatalf("deserializeFromYaml: %v", err)
	}

	if loaded.DatabaseReady != original.DatabaseReady {
		t.Errorf("DatabaseReady: got %v, want %v", loaded.DatabaseReady, original.DatabaseReady)
	}
	if loaded.SongsReady != original.SongsReady {
		t.Errorf("SongsReady: got %v, want %v", loaded.SongsReady, original.SongsReady)
	}
	if loaded.Sorting != original.Sorting {
		t.Errorf("Sorting: got %q, want %q", loaded.Sorting, original.Sorting)
	}
	if loaded.SearchPattern != original.SearchPattern {
		t.Errorf("SearchPattern: got %q, want %q", loaded.SearchPattern, original.SearchPattern)
	}
}

func TestDeserializeFromYaml_MissingFile(t *testing.T) {
	app := &App{appDir: t.TempDir()}
	var status AppStatus
	if err := app.deserializeFromYaml(&status, "nonexistent.yaml"); err == nil {
		t.Error("expected error when file does not exist")
	}
}

func TestDeserializeFromYaml_InvalidYaml(t *testing.T) {
	tmpDir := t.TempDir()
	app := &App{appDir: tmpDir}
	invalidYAML := []byte("{{{invalid")
	if err := os.WriteFile(filepath.Join(tmpDir, "bad.yaml"), invalidYAML, 0644); err != nil {
		t.Fatal(err)
	}
	var status AppStatus
	if err := app.deserializeFromYaml(&status, "bad.yaml"); err == nil {
		t.Error("expected error for invalid YAML")
	}
}

// ─── GetSongVerses ───────────────────────────────────────────────────────────

func TestGetSongVerses(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	// Populate with sample songs
	xmlDir, err := os.MkdirTemp("", "testxml_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(xmlDir)

	ezDir := filepath.Join(xmlDir, "EZ")
	if err := os.MkdirAll(ezDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := copyDir("testdata/", ezDir); err != nil {
		t.Fatalf("copyDir: %v", err)
	}
	app.songBookDir = xmlDir
	app.FillDatabase()

	// Get the first song id
	songs, err := app.GetSongs("entry", "")
	if err != nil || len(songs) == 0 {
		t.Fatalf("GetSongs failed or returned empty: %v", err)
	}
	songID := songs[0].Id

	result, err := app.GetSongVerses(songID)
	if err != nil {
		t.Fatalf("GetSongVerses: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty verses")
	}
	// Separator between verses
	if len(songs[0].Verses) > 0 {
		// verses should be joined by "==="
		// we can only check the separator if there are multiple verses
		t.Logf("GetSongVerses returned %d chars for song %d", len(result), songID)
	}
}

func TestGetSongVerses_MultipleVersesSeparator(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	xmlDir, err := os.MkdirTemp("", "testxml_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(xmlDir)

	ezDir := filepath.Join(xmlDir, "EZ")
	if err := os.MkdirAll(ezDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := copyDir("testdata/", ezDir); err != nil {
		t.Fatalf("copyDir: %v", err)
	}
	app.songBookDir = xmlDir
	app.FillDatabase()

	songs, err := app.GetSongs("entry", "")
	if err != nil || len(songs) == 0 {
		t.Fatal("no songs available")
	}

	// Find a song with more than one verse
	for _, s := range songs {
		if strings.Contains(s.Verses, "\n\n") {
			result, err := app.GetSongVerses(s.Id)
			if err != nil {
				t.Fatalf("GetSongVerses: %v", err)
			}
			if !strings.Contains(result, "===") {
				t.Errorf("expected '===' separator in multi-verse result, got: %q", result[:min(len(result), 100)])
			}
			return
		}
	}
}

func TestGetSongVerses_NonExistentSong(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	result, err := app.GetSongVerses(99999)
	if err != nil {
		t.Fatalf("unexpected error for non-existent song: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string for non-existent song ID, got %q", result)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ─── getXmlFilesFromDir ──────────────────────────────────────────────────────

func TestGetXmlFilesFromDir_MissingDir(t *testing.T) {
	app := &App{}
	_, err := app.getXmlFilesFromDir("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

func TestGetXmlFilesFromDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	app := &App{}
	_, err := app.getXmlFilesFromDir(dir)
	if err == nil {
		t.Error("expected error for empty directory")
	}
}

func TestGetXmlFilesFromDir_WithFiles(t *testing.T) {
	dir := t.TempDir()
	// Create a few files
	for _, name := range []string{"a.xml", "b.xml", "c.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("<x/>"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	app := &App{}
	files, err := app.getXmlFilesFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 3 { // .txt is also returned – function returns all non-dir entries
		t.Errorf("expected 3 files, got %d", len(files))
	}
}

func TestGetXmlFilesFromDir_FiltersDirs(t *testing.T) {
	dir := t.TempDir()
	// Create a subdirectory and a file
	if err := os.MkdirAll(filepath.Join(dir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "song.xml"), []byte("<x/>"), 0644); err != nil {
		t.Fatal(err)
	}
	app := &App{}
	files, err := app.getXmlFilesFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range files {
		if f.IsDir() {
			t.Errorf("expected no directories in result, got %s", f.Name())
		}
	}
}

func TestGetXmlFilesFromDir_TestRunLimits25(t *testing.T) {
	dir := t.TempDir()
	// Create 30 files
	for i := 0; i < 30; i++ {
		name := filepath.Join(dir, strings.Repeat("a", i+1)+".xml")
		if err := os.WriteFile(name, []byte("<x/>"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	app := &App{testRun: true}
	files, err := app.getXmlFilesFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) > 25 {
		t.Errorf("expected at most 25 files in testRun mode, got %d", len(files))
	}
}

// ─── getOrCreateSongbook ─────────────────────────────────────────────────────

func TestGetOrCreateSongbook_CreateNew(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	err := app.withDB(func(db *sql.DB) error {
		acronym, err := app.getOrCreateSongbook(db, "TEST", "Test Songbook")
		if err != nil {
			return err
		}
		if acronym != "TEST" {
			t.Errorf("expected acronym %q, got %q", "TEST", acronym)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("withDB: %v", err)
	}
}

func TestGetOrCreateSongbook_ReturnExisting(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	err := app.withDB(func(db *sql.DB) error {
		// Create it once
		if _, err := app.getOrCreateSongbook(db, "EZ", "Evangelický zpěvník 2021"); err != nil {
			return err
		}
		// Call again – should return "EZ" without error
		acronym, err := app.getOrCreateSongbook(db, "EZ", "Evangelický zpěvník 2021")
		if err != nil {
			return err
		}
		if acronym != "EZ" {
			t.Errorf("expected %q, got %q", "EZ", acronym)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("withDB: %v", err)
	}
}

func TestGetOrCreateSongbook_TooLongAcronym(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	err := app.withDB(func(db *sql.DB) error {
		_, err := app.getOrCreateSongbook(db, "TOOLONGACRONYM", "some name")
		if err == nil {
			t.Error("expected error for acronym > 10 chars")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("withDB: %v", err)
	}
}

// ─── processKKSongFile ───────────────────────────────────────────────────────

func TestProcessKKSongFile(t *testing.T) {
	// Prepare a directory mimicking KK/Kancional layout
	tmpDir := t.TempDir()
	kkDir := filepath.Join(tmpDir, "KK", "Kancional")
	if err := os.MkdirAll(kkDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Copy the KK sample fixture
	src, err := os.ReadFile("testdata/kk_sample_0.xml")
	if err != nil {
		t.Fatalf("reading kk_sample_0.xml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(kkDir, "kk_sample_0.xml"), src, 0644); err != nil {
		t.Fatal(err)
	}

	app := &App{
		dbFilePath:  filepath.Join(tmpDir, "test.db"),
		songBookDir: tmpDir,
	}
	app.InitializeDatabase()

	err = app.withDB(func(db *sql.DB) error {
		songbookAcronym, err := app.getOrCreateSongbook(db, Acronym_KK, "Katolický kancionál")
		if err != nil {
			return err
		}
		entries, err := os.ReadDir(kkDir)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			t.Fatal("no entries in KK dir")
		}
		return app.processKKSongFile(db, entries[0], songbookAcronym)
	})
	if err != nil {
		t.Fatalf("processKKSongFile: %v", err)
	}

	// Verify song was inserted
	songs, err := app.GetSongs("entry", "")
	if err != nil {
		t.Fatalf("GetSongs: %v", err)
	}
	if len(songs) == 0 {
		t.Error("expected at least one song after processKKSongFile")
	}
}
