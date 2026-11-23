package app

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateToV3_AddsVarhanyFileColumn(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	db, err := sql.Open("sqlite3", app.dbFilePath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('songs') WHERE name = 'varhany_file'`).Scan(&count)
	if err != nil {
		t.Fatalf("query pragma: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected varhany_file column, got %d", count)
	}
}

func TestGetSongs_IncludesVarhanyFile(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	db, err := sql.Open("sqlite3", app.dbFilePath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`INSERT OR IGNORE INTO songbooks (songbook_acronym, name) VALUES ('EZ','Evangelický zpěvník')`)
	if err != nil {
		t.Fatalf("insert songbook: %v", err)
	}
	_, err = db.Exec(`INSERT INTO songs (id, entry, title, title_d, verse_order, kytara_file, varhany_file, songbook_acronym) VALUES (1, 2, 'Test', 'test', '', 'kytara_002.pdf', 'varhany_002.pdf', 'EZ')`)
	if err != nil {
		t.Fatalf("insert song: %v", err)
	}
	_, err = db.Exec(`INSERT INTO verses (song_id, name, lines, lines_d) VALUES (1, 'v1', 'line', 'line')`)
	if err != nil {
		t.Fatalf("insert verse: %v", err)
	}

	songs, err := app.GetSongs("entry", "", "EZ")
	if err != nil {
		t.Fatalf("GetSongs: %v", err)
	}
	if len(songs) != 1 {
		t.Fatalf("expected 1 song, got %d", len(songs))
	}
	if songs[0].KytaraFile != "kytara_002.pdf" {
		t.Fatalf("expected kytara file, got %q", songs[0].KytaraFile)
	}
	if songs[0].VarhanyFile != "varhany_002.pdf" {
		t.Fatalf("expected varhany file, got %q", songs[0].VarhanyFile)
	}
}

func TestGetSongs2_IncludesVarhanyFile(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	db, err := sql.Open("sqlite3", app.dbFilePath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`INSERT OR IGNORE INTO songbooks (songbook_acronym, name) VALUES ('EZ','Evangelický zpěvník')`)
	if err != nil {
		t.Fatalf("insert songbook: %v", err)
	}
	_, err = db.Exec(`INSERT INTO songs (id, entry, title, title_d, verse_order, kytara_file, varhany_file, songbook_acronym) VALUES (1, 2, 'Test', 'test', '', 'kytara_002.pdf', 'varhany_002.pdf', 'EZ')`)
	if err != nil {
		t.Fatalf("insert song: %v", err)
	}

	songs, err := app.GetSongs2("entry", "", "EZ")
	if err != nil {
		t.Fatalf("GetSongs2: %v", err)
	}
	if len(songs) != 1 {
		t.Fatalf("expected 1 song, got %d", len(songs))
	}
	if songs[0].VarhanyFile != "varhany_002.pdf" {
		t.Fatalf("expected varhany file, got %q", songs[0].VarhanyFile)
	}
}

func TestUpdateSongFilenames_VarhanyColumn(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	db, err := sql.Open("sqlite3", app.dbFilePath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`INSERT OR IGNORE INTO songbooks (songbook_acronym, name) VALUES ('EZ','Evangelický zpěvník')`)
	if err != nil {
		t.Fatalf("insert songbook: %v", err)
	}
	_, err = db.Exec(`INSERT INTO songs (id, entry, title, title_d, verse_order, songbook_acronym) VALUES (1, 2, 'Test', 'test', '', 'EZ')`)
	if err != nil {
		t.Fatalf("insert song: %v", err)
	}

	err = app.updateSongFilenames(map[int]string{2: "varhany_002.pdf"}, "varhany_file")
	if err != nil {
		t.Fatalf("updateSongFilenames: %v", err)
	}

	var got string
	err = db.QueryRow(`SELECT COALESCE(varhany_file,'') FROM songs WHERE entry=2 AND songbook_acronym='EZ'`).Scan(&got)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if got != "varhany_002.pdf" {
		t.Fatalf("expected varhany_002.pdf, got %q", got)
	}
}

func TestDownloadOptionalPdf(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ok.pdf" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("%PDF-1.4 test"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	app := &App{}
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "x.pdf")
	if err := app.downloadOptionalPdf(server.URL+"/ok.pdf", target); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(data) != "%PDF-1.4 test" {
		t.Fatalf("unexpected content: %q", string(data))
	}

	missing := filepath.Join(tmpDir, "missing.pdf")
	if err := app.downloadOptionalPdf(server.URL+"/missing.pdf", missing); err == nil {
		t.Fatal("expected error for missing pdf")
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatalf("missing file should not exist, stat err: %v", err)
	}
}

func TestProcessVarhanyPDFs_DownloadsAndUpdatesDB(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/002.pdf":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("%PDF-1.4 two"))
		case "/004.pdf":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("%PDF-1.4 four"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	app := setupTestDB(t)
	defer teardownTestDB(app)
	app.pdfDir = t.TempDir()
	app.varhanyBaseURL = server.URL

	db, err := sql.Open("sqlite3", app.dbFilePath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`INSERT OR IGNORE INTO songbooks (songbook_acronym, name) VALUES ('EZ','Evangelický zpěvník')`)
	if err != nil {
		t.Fatalf("insert songbook: %v", err)
	}
	for _, entry := range []int{2, 3, 4} {
		_, err = db.Exec(`INSERT INTO songs (id, entry, title, title_d, verse_order, songbook_acronym) VALUES (?, ?, ?, ?, '', 'EZ')`, entry, entry, fmt.Sprintf("Song %d", entry), fmt.Sprintf("song %d", entry))
		if err != nil {
			t.Fatalf("insert song %d: %v", entry, err)
		}
	}

	if err := app.ProcessVarhanyPDFs(); err != nil {
		t.Fatalf("ProcessVarhanyPDFs: %v", err)
	}

	for _, tc := range []struct {
		entry int
		want  string
	}{
		{2, "varhany_002.pdf"},
		{3, ""},
		{4, "varhany_004.pdf"},
	} {
		var got string
		err = db.QueryRow(`SELECT COALESCE(varhany_file,'') FROM songs WHERE entry=? AND songbook_acronym='EZ'`, tc.entry).Scan(&got)
		if err != nil {
			t.Fatalf("query entry %d: %v", tc.entry, err)
		}
		if got != tc.want {
			t.Fatalf("entry %d: expected %q, got %q", tc.entry, tc.want, got)
		}
	}

	if _, err := os.Stat(filepath.Join(app.pdfDir, "varhany_002.pdf")); err != nil {
		t.Fatalf("expected downloaded file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(app.pdfDir, "varhany_003.pdf")); !os.IsNotExist(err) {
		t.Fatalf("did not expect missing file, stat err: %v", err)
	}
}

func TestProcessVarhanyPDFs_RemoteIntegration(t *testing.T) {
	if os.Getenv("RUN_REMOTE_INTEGRATION_TESTS") == "" {
		t.Skip("skipping remote integration test; set RUN_REMOTE_INTEGRATION_TESTS=1 to enable")
	}

	app := setupTestDB(t)
	defer teardownTestDB(app)
	app.pdfDir = t.TempDir()

	db, err := sql.Open("sqlite3", app.dbFilePath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`INSERT OR IGNORE INTO songbooks (songbook_acronym, name) VALUES ('EZ','Evangelický zpěvník')`)
	if err != nil {
		t.Fatalf("insert songbook: %v", err)
	}
	_, err = db.Exec(`INSERT INTO songs (id, entry, title, title_d, verse_order, songbook_acronym) VALUES (1, 2, 'Test 2', 'test 2', '', 'EZ')`)
	if err != nil {
		t.Fatalf("insert song: %v", err)
	}

	if err := app.ProcessVarhanyPDFs(); err != nil {
		t.Fatalf("ProcessVarhanyPDFs: %v", err)
	}

	var got string
	err = db.QueryRow(`SELECT COALESCE(varhany_file,'') FROM songs WHERE entry=2 AND songbook_acronym='EZ'`).Scan(&got)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if got != "varhany_002.pdf" {
		t.Fatalf("expected varhany_002.pdf, got %q", got)
	}
	if _, err := os.Stat(filepath.Join(app.pdfDir, "varhany_002.pdf")); err != nil {
		t.Fatalf("expected downloaded organ pdf: %v", err)
	}
}
