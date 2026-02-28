package app

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// ─── getXmlFiles (backward-compat wrapper) ───────────────────────────────────

func TestGetXmlFiles_WithFilesInSongBookDir(t *testing.T) {
	tmpDir := t.TempDir()
	app := &App{songBookDir: tmpDir}

	for _, name := range []string{"a.xml", "b.xml"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("<x/>"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	files, err := app.getXmlFiles()
	if err != nil {
		t.Fatalf("getXmlFiles: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestGetXmlFiles_EmptyDir(t *testing.T) {
	app := &App{songBookDir: t.TempDir()}
	_, err := app.getXmlFiles()
	if err == nil {
		t.Error("expected error for empty dir")
	}
}

// ─── fillKKSongs ─────────────────────────────────────────────────────────────

func TestFillKKSongs_DirectoryMissing(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	// No KK directory – should silently succeed
	err := app.withDB(func(db *sql.DB) error {
		return app.fillKKSongs(db)
	})
	if err != nil {
		t.Fatalf("fillKKSongs with missing dir should return nil, got: %v", err)
	}
}

func TestFillKKSongs_WithFixture(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	// Build KK/Kancional layout in a fresh songBookDir
	tmpSongs := t.TempDir()
	app.songBookDir = tmpSongs
	kkDir := filepath.Join(tmpSongs, "KK", "Kancional")
	if err := os.MkdirAll(kkDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Copy the KK sample fixture into the kancional dir
	src, err := os.ReadFile("testdata/kk_sample_0.xml")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(kkDir, "kk_sample_0.xml"), src, 0644); err != nil {
		t.Fatal(err)
	}

	err = app.withDB(func(db *sql.DB) error {
		return app.fillKKSongs(db)
	})
	if err != nil {
		t.Fatalf("fillKKSongs: %v", err)
	}

	// Verify at least one KK song is in the DB
	songs, err := app.GetSongs("entry", "")
	if err != nil {
		t.Fatalf("GetSongs: %v", err)
	}
	if len(songs) == 0 {
		t.Error("expected at least one KK song after fillKKSongs")
	}
	// Ensure it carries the KK songbook acronym
	found := false
	for _, s := range songs {
		if s.SongbookAcronym == Acronym_KK {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a song with KK songbook acronym")
	}
}

// ─── GetSongProjection ───────────────────────────────────────────────────────

func setupDBWithSong(t *testing.T) (*App, int) {
	t.Helper()
	app := setupTestDB(t)

	xmlDir := t.TempDir()
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
		t.Fatalf("no songs after FillDatabase: %v", err)
	}
	return app, songs[0].Id
}

func TestGetSongProjection_ReturnsValidJSON(t *testing.T) {
	app, songID := setupDBWithSong(t)
	defer teardownTestDB(app)

	result, err := app.GetSongProjection(songID)
	if err != nil {
		t.Fatalf("GetSongProjection: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty JSON")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(result), &payload); err != nil {
		t.Fatalf("invalid JSON from GetSongProjection: %v — got: %s", err, result)
	}
	if _, ok := payload["verse_order"]; !ok {
		t.Error("expected 'verse_order' key in JSON")
	}
	if _, ok := payload["verses"]; !ok {
		t.Error("expected 'verses' key in JSON")
	}
}

func TestGetSongProjection_NonExistentSong(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	result, err := app.GetSongProjection(99999)
	if err != nil {
		t.Fatalf("unexpected error for non-existent song: %v", err)
	}
	// Should return JSON with empty verse_order and empty verses
	if !strings.Contains(result, "verse_order") {
		t.Errorf("expected verse_order in result, got: %q", result)
	}
}

func TestGetSongProjection_VerseOrderPreserved(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	// Insert a song with a known verse_order directly
	var songID int64
	err := app.withDB(func(db *sql.DB) error {
		r, err := db.Exec(
			`INSERT INTO songs (songbook_acronym, title, title_d, verse_order, entry, entry_text) VALUES (?, ?, ?, ?, ?, ?)`,
			"EZ", "Test Song", "test song", "v1 c v2", 999, "999",
		)
		if err != nil {
			return err
		}
		songID, err = r.LastInsertId()
		if err != nil {
			return err
		}
		_, err = db.Exec(`INSERT INTO verses (song_id, name, lines, lines_d) VALUES (?, ?, ?, ?)`,
			songID, "v1", "line one", "line one")
		return err
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	result, err := app.GetSongProjection(int(songID))
	if err != nil {
		t.Fatalf("GetSongProjection: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(result), &payload); err != nil {
		t.Fatalf("bad JSON: %v", err)
	}
	if payload["verse_order"] != "v1 c v2" {
		t.Errorf("expected verse_order 'v1 c v2', got %q", payload["verse_order"])
	}
	verses, ok := payload["verses"].([]interface{})
	if !ok || len(verses) == 0 {
		t.Error("expected non-empty verses array")
	}
}

// ─── applyMigration ──────────────────────────────────────────────────────────

func TestApplyMigration_UnknownVersion(t *testing.T) {
	app := &App{}
	err := app.withDB(func(db *sql.DB) error {
		return app.applyMigration(db, 99)
	})
	// Need a valid DB path to call withDB; let's call directly
	_ = err

	// Call applyMigration directly via openDB
	dbFile, _ := os.CreateTemp("", "testdb_*.sqlite")
	dbPath := dbFile.Name()
	dbFile.Close()
	defer os.Remove(dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	a := &App{}
	if err := a.applyMigration(db, 99); err == nil {
		t.Error("expected error for unknown migration version")
	}
}

// ─── parseXmlSongKK error paths ──────────────────────────────────────────────

func TestParseXmlSongKK_MalformedXML(t *testing.T) {
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "bad.xml")
	if err := os.WriteFile(badFile, []byte("<<< not xml >>>"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parseXmlSongKK(badFile)
	if err == nil {
		t.Error("expected error for malformed XML")
	}
}

func TestParseXmlSongKK_MissingFile(t *testing.T) {
	_, err := parseXmlSongKK("/nonexistent/path.xml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// ─── processEZSongFile error path ────────────────────────────────────────────

func TestProcessEZSongFile_MalformedXML(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	// Create a malformed XML file
	tmpDir := t.TempDir()
	ezDir := filepath.Join(tmpDir, "EZ")
	if err := os.MkdirAll(ezDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ezDir, "bad.xml"), []byte("<<bad"), 0644); err != nil {
		t.Fatal(err)
	}
	app.songBookDir = tmpDir

	entries, _ := os.ReadDir(ezDir)
	err := app.withDB(func(db *sql.DB) error {
		return app.processEZSongFile(db, entries[0], "EZ")
	})
	if err == nil {
		t.Error("expected error for malformed XML")
	}
}

// ─── GetSongVerses error path ─────────────────────────────────────────────

func TestGetSongVerses_InvalidDB(t *testing.T) {
	app := &App{dbFilePath: "/nonexistent/path/db.sqlite"}
	_, err := app.GetSongVerses(1)
	if err == nil {
		t.Error("expected error with invalid db path")
	}
}

// ─── GetSongs error path ─────────────────────────────────────────────────────

func TestGetSongs_InvalidDB(t *testing.T) {
	app := &App{dbFilePath: "/nonexistent/db.sqlite"}
	_, err := app.GetSongs("entry", "")
	if err == nil {
		t.Error("expected error with invalid db path")
	}
	if app.status.DatabaseReady {
		t.Error("expected DatabaseReady=false after error")
	}
}

func TestGetSongs2_InvalidDB(t *testing.T) {
	app := &App{dbFilePath: "/nonexistent/db.sqlite"}
	_, err := app.GetSongs2("entry", "")
	if err == nil {
		t.Error("expected error with invalid db path")
	}
	if app.status.DatabaseReady {
		t.Error("expected DatabaseReady=false after error")
	}
}

// ─── hasDatabaseContent withDB error ─────────────────────────────────────────

func TestHasDatabaseContent_InvalidDB(t *testing.T) {
	app := setupStatusApp(t)
	app.InitializeDatabase()
	// Corrupt the DB path after init so withDB fails
	app.dbFilePath = "/invalid/path/to/db.sqlite"
	// File stat will fail first (no file there), so hasDatabaseContent → false
	if app.hasDatabaseContent() {
		t.Error("expected false for non-existent DB file")
	}
}

// ─── detectSchemaVersion with v2 data ────────────────────────────────────────

func TestDetectSchemaVersion_SchemaVersionTableWithData(t *testing.T) {
	dbFile, err := os.CreateTemp("", "testdb_*.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := dbFile.Name()
	dbFile.Close()
	defer os.Remove(dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Manually create schema_version table and insert version 2
	if _, err := db.Exec(`CREATE TABLE schema_version (version INTEGER PRIMARY KEY, applied_at DATETIME DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO schema_version (version) VALUES (2)`); err != nil {
		t.Fatal(err)
	}

	app := &App{}
	version, err := app.detectSchemaVersion(db)
	if err != nil {
		t.Fatalf("detectSchemaVersion: %v", err)
	}
	if version != 2 {
		t.Errorf("expected version 2, got %d", version)
	}
}

func TestDetectSchemaVersion_EmptySchemaVersionTable(t *testing.T) {
	dbFile, err := os.CreateTemp("", "testdb_*.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := dbFile.Name()
	dbFile.Close()
	defer os.Remove(dbPath)

	db, _ := sql.Open("sqlite3", dbPath)
	defer db.Close()

	// Create schema_version table but leave it empty
	db.Exec(`CREATE TABLE schema_version (version INTEGER PRIMARY KEY, applied_at DATETIME)`) //nolint

	app := &App{}
	version, err := app.detectSchemaVersion(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != 1 {
		t.Errorf("expected version 1 for empty schema_version table, got %d", version)
	}
}
