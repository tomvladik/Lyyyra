package app

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func Test_parseXmlSongKK(t *testing.T) {
	song, err := parseXmlSongKK("testdata/kk_sample_0.xml")
	if err != nil {
		t.Fatalf("Failed to parse KK XML: %v", err)
	}

	if song.Title != "065 Litanie k nejsvětějšímu Srdci Ježíšovu" {
		t.Errorf("Expected title '065 Litanie k nejsvětějšímu Srdci Ježíšovu', got '%s'", song.Title)
	}

	if song.HymnNumber != "065" {
		t.Errorf("Expected hymn number '065', got '%s'", song.HymnNumber)
	}

	if song.Lyrics == "" {
		t.Error("Expected lyrics to be non-empty")
	}

	// Check that lyrics contain verse markers
	if !strings.Contains(song.Lyrics, "[V1]") {
		t.Error("Expected lyrics to contain [V1] verse marker")
	}

	t.Logf("Successfully parsed KK song: %s", song.Title)
	t.Logf("Hymn number: %s", song.HymnNumber)
	t.Logf("Lyrics length: %d chars", len(song.Lyrics))
}

func Test_insertSongKK_stripsTitleNumber(t *testing.T) {
	// Create temporary database using production logic
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Use production initialization to get latest schema (V3)
	app := &App{
		dbFilePath: dbPath,
	}
	app.InitializeDatabase()

	// Open connection for test queries
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	tests := []struct {
		name          string
		title         string
		hymnNumber    string
		expectedTitle string
		expectedEntry int
	}{
		{
			name:          "strips numeric prefix with space",
			title:         "413 Žije! Kristus povstal z hrobu",
			hymnNumber:    "413",
			expectedTitle: "Žije! Kristus povstal z hrobu",
			expectedEntry: 413,
		},
		{
			name:          "strips zero-padded prefix",
			title:         "065 Litanie k nejsvětějšímu Srdci Ježíšovu",
			hymnNumber:    "065",
			expectedTitle: "Litanie k nejsvětějšímu Srdci Ježíšovu",
			expectedEntry: 65,
		},
		{
			name:          "handles title without numeric prefix",
			title:         "Chvalte Hospodina",
			hymnNumber:    "100",
			expectedTitle: "Chvalte Hospodina",
			expectedEntry: 100,
		},
		{
			name:          "handles title starting with number in name",
			title:         "3 králové přišli",
			hymnNumber:    "50",
			expectedTitle: "králové přišli",
			expectedEntry: 50,
		},
		{
			name:          "handles hymn number with letter suffix",
			title:         "511A Viděl jsem pramen vody",
			hymnNumber:    "511A",
			expectedTitle: "Viděl jsem pramen vody",
			expectedEntry: 511,
		},
		{
			name:          "handles hymn number with lowercase letter",
			title:         "067b Závěr litanie",
			hymnNumber:    "067b",
			expectedTitle: "Závěr litanie",
			expectedEntry: 67,
		},
		{
			name:          "handles complex hymn number with dot",
			title:         "2.4. Antifony na Zelený čtvrtek",
			hymnNumber:    "2.4",
			expectedTitle: "Antifony na Zelený čtvrtek",
			expectedEntry: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			song := &SongKK{
				Title:      tt.title,
				HymnNumber: tt.hymnNumber,
			}

			songID, err := app.insertSongKK(db, song, "KK")
			if err != nil {
				t.Fatalf("insertSongKK failed: %v", err)
			}

			// Verify the song was inserted correctly
			var title string
			var entry int
			err = db.QueryRow("SELECT title, entry FROM songs WHERE id = ?", songID).Scan(&title, &entry)
			if err != nil {
				t.Fatalf("Failed to query inserted song: %v", err)
			}

			if title != tt.expectedTitle {
				t.Errorf("Expected title '%s', got '%s'", tt.expectedTitle, title)
			}

			if entry != tt.expectedEntry {
				t.Errorf("Expected entry %d, got %d", tt.expectedEntry, entry)
			}
		})
	}
}

func Test_insertVersesKK_parsesVerseMarkers(t *testing.T) {
	// Create temporary database using production logic
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Use production initialization to get latest schema (V3)
	app := &App{
		dbFilePath: dbPath,
	}
	app.InitializeDatabase()

	// Open connection for test queries
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	app = &App{}
	songID := int64(1)

	lyrics := `[V1]
 Pane, smiluj se.
 Pane, smiluj se.
 Kriste, smiluj se.
[V2]
 Bože, náš nebeský Otče,
     smiluj se nad námi.
 Bože Synu, Vykupiteli světa,
     smiluj se nad námi.
[V3]
 Srdce Ježíšovo, Srdce Syna věčného Otce,
     smiluj se nad námi.`

	err = app.insertVersesKK(db, songID, lyrics, "test.xml")
	if err != nil {
		t.Fatalf("insertVersesKK failed: %v", err)
	}

	// Verify verses were inserted
	rows, err := db.Query("SELECT name, lines FROM verses WHERE song_id = ? ORDER BY id", songID)
	if err != nil {
		t.Fatalf("Failed to query verses: %v", err)
	}
	defer rows.Close()

	var verses []struct {
		name  string
		lines string
	}
	for rows.Next() {
		var name, lines string
		if err := rows.Scan(&name, &lines); err != nil {
			t.Fatalf("Failed to scan verse: %v", err)
		}
		verses = append(verses, struct {
			name  string
			lines string
		}{name, lines})
	}

	if len(verses) != 3 {
		t.Fatalf("Expected 3 verses, got %d", len(verses))
	}

	// Check first verse
	if verses[0].name != "v1" {
		t.Errorf("Expected verse name 'v1', got '%s'", verses[0].name)
	}
	if !strings.Contains(verses[0].lines, "Pane, smiluj se.") {
		t.Errorf("Verse 1 doesn't contain expected text, got: %s", verses[0].lines)
	}

	// Check second verse
	if verses[1].name != "v2" {
		t.Errorf("Expected verse name 'v2', got '%s'", verses[1].name)
	}
	if !strings.Contains(verses[1].lines, "Bože, náš nebeský Otče") {
		t.Errorf("Verse 2 doesn't contain expected text, got: %s", verses[1].lines)
	}

	// Check third verse
	if verses[2].name != "v3" {
		t.Errorf("Expected verse name 'v3', got '%s'", verses[2].name)
	}
	if !strings.Contains(verses[2].lines, "Srdce Ježíšovo") {
		t.Errorf("Verse 3 doesn't contain expected text, got: %s", verses[2].lines)
	}

	t.Logf("Successfully parsed and inserted %d verses", len(verses))
}

func Test_insertVersesKK_handlesEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Use production initialization to get latest schema (V3)
	app := &App{
		dbFilePath: dbPath,
	}
	app.InitializeDatabase()

	// Open connection for test queries
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	app = &App{}
	songID := int64(1)

	// Lyrics with extra blank lines
	lyrics := `[V1]

 First line

 Second line

[V2]

 Third line`

	err = app.insertVersesKK(db, songID, lyrics, "test.xml")
	if err != nil {
		t.Fatalf("insertVersesKK failed: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM verses WHERE song_id = ?", songID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count verses: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 verses, got %d", count)
	}
}

func Test_parseXmlSongKK_fromActualFile(t *testing.T) {
	// Only run if sample file exists
	samplePath := "../../build/tmp/kk_sample_0.xml"
	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skip("Sample KK file not found, skipping test")
	}

	song, err := parseXmlSongKK(samplePath)
	if err != nil {
		t.Fatalf("Failed to parse KK XML: %v", err)
	}

	// Basic validations
	if song.Title == "" {
		t.Error("Title should not be empty")
	}

	if song.HymnNumber == "" {
		t.Error("HymnNumber should not be empty")
	}

	if song.Lyrics == "" {
		t.Error("Lyrics should not be empty")
	}

	// Should have verse markers
	if !strings.Contains(song.Lyrics, "[V") {
		t.Error("Lyrics should contain verse markers like [V1]")
	}

	t.Logf("Parsed KK song: %s (hymn %s)", song.Title, song.HymnNumber)
}

func Test_parseHymnNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "pure number",
			input:    "413",
			expected: 413,
		},
		{
			name:     "zero-padded number",
			input:    "065",
			expected: 65,
		},
		{
			name:     "number with uppercase letter",
			input:    "511A",
			expected: 511,
		},
		{
			name:     "number with lowercase letter",
			input:    "067b",
			expected: 67,
		},
		{
			name:     "number with dot",
			input:    "2.4",
			expected: 2,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "only letters",
			input:    "abc",
			expected: 0,
		},
		{
			name:     "whitespace padded",
			input:    "  413  ",
			expected: 413,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseHymnNumber(tt.input)
			if result != tt.expected {
				t.Errorf("parseHymnNumber(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
