package app

import (
	"database/sql"
	"encoding/json"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *App {
	// Create a temporary database file
	dbFile, err := os.CreateTemp("", "testdb_*.sqlite")
	if err != nil {
		t.Fatalf("Failed to create temp database file: %v", err)
	}
	dbFilePath := dbFile.Name()
	dbFile.Close()

	// Initialize the App with the test database file path
	app := &App{
		dbFilePath: dbFilePath,
	}

	// Prepare the database
	app.PrepareDatabase()

	return app
}

func teardownTestDB(app *App) {
	// Remove the temporary database file
	os.Remove(app.dbFilePath)
}

func TestFillDatabase(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	// Create a temporary directory for XML files
	xmlDir, err := os.MkdirTemp("", "testxml_")
	if err != nil {
		t.Fatalf("Failed to create temp XML directory: %v", err)
	}
	defer os.RemoveAll(xmlDir)

	// Copy the sample XML file to the temporary directory
	err = copyDir("testdata/", xmlDir)
	if err != nil {
		t.Fatalf("Failed to copy temp XML directory: %v", err)
	}
	// Set the songBookDir to the temporary directory
	app.songBookDir = xmlDir

	// Fill the database with the sample XML file
	app.FillDatabase()

	// Open the database
	db, err := sql.Open("sqlite3", app.dbFilePath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check if the song was inserted
	var title string
	err = db.QueryRow(`SELECT title FROM songs WHERE title='ABCčDďE'`).Scan(&title)
	if err != nil {
		t.Errorf("Sample song 'ABCčDďE' was not inserted: %v", err)
	}
	if title != "ABCčDďE" {
		t.Errorf("Expected title 'ABCčDďE', got %q", title)
	}

	// Check if the song was indexed with diacritics removed
	var found_entry string
	err = db.QueryRow(`
		SELECT entry FROM songs
		WHERE title_d LIKE '%abccdde%'`).Scan(&found_entry)
	if err != nil {
		t.Errorf("Song not searchable by normalized title: %v", err)
	}

	if found_entry != "288" {
		t.Errorf("Expected entry '288' from diacritics search, got %q", found_entry)
	}

	err = db.QueryRow(`
    SELECT entry FROM songs
	JOIN authors ON songs.id=authors.song_id
	WHERE authors.author_value_d LIKE '%2018%'`).Scan(&found_entry)
	if err != nil {
		t.Errorf("Failed to find song by author search: %v", err)
	}

	if found_entry != "3" {
		t.Errorf("Expected entry '3' from author search, got %q", found_entry)
	}

	err = db.QueryRow(`
    SELECT entry FROM songs
	JOIN verses ON songs.id=verses.song_id
	WHERE verses.lines_d LIKE '%tulen%'`).Scan(&found_entry)
	if err != nil {
		t.Errorf("Failed to find song by verse search: %v", err)
	}

	if found_entry != "4" {
		t.Errorf("Expected entry '4' from verse search, got %q", found_entry)
	}

}

func TestGetSongs(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	// Insert sample data into the database
	db, err := sql.Open("sqlite3", app.dbFilePath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
        INSERT INTO songs (title, verse_order, entry) VALUES
            ('First Song', '1', 1),
            ('Second Song', '1', 2),
            ('Last Song', '1', 3);
        INSERT INTO verses (song_id, name, lines) VALUES
            (1, 'Verse 1', 'First song lines'),
            (2, 'Verse 1', 'Second song first lines'),
            (2, 'Verse 2', 'Second song second line'),
            (3, 'Verse 1', 'Third song first line');
        INSERT INTO authors (song_id, author_type, author_value) VALUES
            (1, 'music', 'Mark Author Music'),
            (1, 'words', 'Quido Author Words'),
            (2, 'music', 'Aloys Author Music'),
            (2, 'words', 'Xaver Author Words'),
            (3, 'music', 'Zuzel Author Music'),
            (3, 'words', 'Anatoliy Author Words');
    `)
	if err != nil {
		t.Fatalf("Failed to insert sample data: %v", err)
	}

	testCases := []struct {
		name          string
		orderBy       string
		expectedOrder []string
	}{
		{
			name:          "Order by title",
			orderBy:       "title",
			expectedOrder: []string{"First Song", "Last Song", "Second Song"},
		},
		{
			name:          "Order by entry",
			orderBy:       "entry",
			expectedOrder: []string{"First Song", "Second Song", "Last Song"},
		},
		{
			name:          "Order by music author",
			orderBy:       "authorMusic",
			expectedOrder: []string{"Second Song", "First Song", "Last Song"}, // Aloys, Mark, Zuzel
		},
		{
			name:          "Order by words author",
			orderBy:       "authorLyric",
			expectedOrder: []string{"Last Song", "First Song", "Second Song"}, // Anatoliy, Quido, Xaver
		},
		{
			name:          "Order by empty string defaults to entry",
			orderBy:       "",
			expectedOrder: []string{"First Song", "Second Song", "Last Song"},
		},
		{
			name:          "Order by invalid string defaults to entry",
			orderBy:       "drop table songs",
			expectedOrder: []string{"First Song", "Second Song", "Last Song"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get songs from the database
			songs, err := app.GetSongs(tc.orderBy, "")
			if err != nil {
				t.Errorf("Failed to get songs: %v", err)
				return
			}

			// Check if all songs are in the result
			if len(songs) != len(tc.expectedOrder) {
				t.Errorf("Expected to get %d songs, got %d", len(tc.expectedOrder), len(songs))
				return
			}

			// Check if songs are in correct order
			for i, expected := range tc.expectedOrder {
				if songs[i].Title != expected {
					t.Errorf("Expected song %d to be '%s', got '%s'", i+1, expected, songs[i].Title)
				}
			}
		})
	}
}

func TestGetSongs2(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	// Insert sample data into the database
	db, err := sql.Open("sqlite3", app.dbFilePath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
        INSERT INTO songs (title, verse_order, entry) VALUES ('Sample Song', '1', 1);
    `)
	if err != nil {
		t.Fatalf("Failed to insert sample data: %v", err)
	}

	// Get songs from the database
	songs, err := app.GetSongs2("title", "")
	if err != nil {
		t.Errorf("Failed to get songs: %v", err)
	}

	// Check if the sample song is in the result
	if len(songs) != 1 || songs[0].Title != "Sample Song" {
		t.Errorf("Expected to get 'Sample Song', got %v", songs)
	}

	invalidOrderSongs, err := app.GetSongs2("invalid", "")
	if err != nil {
		t.Errorf("Failed to get songs with invalid order: %v", err)
	}
	if len(invalidOrderSongs) != 1 || invalidOrderSongs[0].Title != "Sample Song" {
		t.Errorf("Expected fallback ordering to return 'Sample Song', got %v", invalidOrderSongs)
	}
}

func TestGetSongAuthors(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	// Insert sample data into the database
	db, err := sql.Open("sqlite3", app.dbFilePath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
        INSERT INTO songs (title, verse_order, entry) VALUES ('Sample Song', '1', 1);
        INSERT INTO authors (song_id, author_type, author_value) VALUES (1, 'music', 'Sample Author');
        INSERT INTO authors (song_id, author_type, author_value) VALUES (1, 'words', 'Bedřich Antonn Leoš');
    `)
	if err != nil {
		t.Fatalf("Failed to insert sample data: %v", err)
	}

	// Get song authors from the database
	authors, err := app.GetSongAuthors(1)
	if err != nil {
		t.Errorf("Failed to get song authors: %v", err)
	}

	// Check if the sample author is in the result
	if len(authors) != 2 || authors[0].Value != "Sample Author" {
		t.Errorf("Expected to get 'Sample Author', got %v", authors)
	}
}

func TestFindSongByAuthor(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	// Insert sample data into the database
	db, err := sql.Open("sqlite3", app.dbFilePath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
        INSERT INTO songs (title, verse_order, entry) VALUES ('Sample Song', '1', 1);
        INSERT INTO authors (song_id, author_type, author_value) VALUES (1, 'music', 'Sample Author');
        INSERT INTO authors (song_id, author_type, author_value) VALUES (1, 'words', 'Bedřich Antonn Leoš');
        INSERT INTO songs (title, verse_order, entry) VALUES ('Sample Song II.', '1', 333);
        INSERT INTO authors (song_id, author_type, author_value) VALUES (2, 'music', 'Experimentální žluťoučký kůň');
        INSERT INTO authors (song_id, author_type, author_value) VALUES (2, 'words', 'šumař na střeše');
    `)
	if err != nil {
		t.Fatalf("Failed to insert sample data: %v", err)
	}

	// Get song authors from the database
	authors, err := app.GetSongAuthors(1)
	if err != nil {
		t.Errorf("Failed to get song authors: %v", err)
	}

	// Check if the sample author is in the result
	if len(authors) != 2 || authors[0].Value != "Sample Author" {
		t.Errorf("Expected to get 'Sample Author', got %v", authors)
	}
}

func TestGetSongProjection(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	// Insert sample data
	db, err := sql.Open("sqlite3", app.dbFilePath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO songs (title, title_d, verse_order, entry) VALUES ('Proj Song', 'Proj Song', 'c v1 c v2', 1);
		INSERT INTO verses (song_id, name, lines) VALUES
			(1, 'v1', 'Line1\nLine2'),
			(1, 'c', 'Chorus line'),
			(1, 'v2', 'LineA\nLineB');
	`)
	if err != nil {
		t.Fatalf("Failed to insert sample data: %v", err)
	}

	raw, err := app.GetSongProjection(1)
	if err != nil {
		t.Fatalf("GetSongProjection returned error: %v", err)
	}

	var payload struct {
		VerseOrder string `json:"verse_order"`
		Verses     []struct {
			Name  string `json:"name"`
			Lines string `json:"lines"`
		} `json:"verses"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v\nraw: %s", err, raw)
	}

	if payload.VerseOrder != "c v1 c v2" {
		t.Errorf("unexpected verse_order: got %q", payload.VerseOrder)
	}
	if len(payload.Verses) != 3 {
		t.Fatalf("expected 3 verses, got %d", len(payload.Verses))
	}
	if payload.Verses[0].Name != "v1" {
		t.Errorf("first verse name expected 'v1', got %q", payload.Verses[0].Name)
	}
	if payload.Verses[1].Name != "c" {
		t.Errorf("second verse name expected 'c', got %q", payload.Verses[1].Name)
	}
	if payload.Verses[2].Name != "v2" {
		t.Errorf("third verse name expected 'v2', got %q", payload.Verses[2].Name)
	}
}

func TestGetSongsWithSearch(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	// Insert sample data into the database
	db, err := sql.Open("sqlite3", app.dbFilePath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
        INSERT INTO songs (title, title_d, verse_order, entry) VALUES
            ('Příliš žluťoučký kůň', 'prilis zlutoucka kun', '1', 101),
            ('Spring Song', 'spring song', '1', 202),
            ('Third Entry', 'third entry', '1', 303);
        INSERT INTO verses (song_id, name, lines, lines_d) VALUES
            (1, 'Verse 1', 'Příliš žluťoučký verses', 'prilis zlutoucka verses'),
            (2, 'Verse 1', 'Spring is coming now', 'spring is coming now'),
            (3, 'Verse 1', 'Another verse text', 'another verse text');
        INSERT INTO authors (song_id, author_type, author_value, author_value_d) VALUES
            (1, 'music', 'Antonín Dvořák', 'antonin dvorak'),
            (1, 'words', 'František Matěj', 'frantisek matej'),
            (2, 'music', 'Wolfgang Mozart', 'wolfgang mozart'),
            (3, 'music', 'Another Author', 'another author');
    `, &err)
	if err != nil {
		t.Fatalf("Failed to insert sample data: %v", err)
	}

	testCases := []struct {
		name            string
		searchPattern   string
		expectedCount   int
		expectedEntries []int
		testDescription string
	}{
		{
			name:            "Search by diacritics-insensitive title",
			searchPattern:   "prilis",
			expectedCount:   1,
			expectedEntries: []int{101},
			testDescription: "Should find 'Příliš' when searching for 'prilis'",
		},
		{
			name:            "Search by diacritics-insensitive author",
			searchPattern:   "dvorak",
			expectedCount:   1,
			expectedEntries: []int{101},
			testDescription: "Should find 'Dvořák' when searching for 'dvorak'",
		},
		{
			name:            "Search by verse lines",
			searchPattern:   "spring",
			expectedCount:   1,
			expectedEntries: []int{202},
			testDescription: "Should find song with 'Spring' in verses",
		},
		{
			name:            "Search by entry number as integer",
			searchPattern:   "202",
			expectedCount:   1,
			expectedEntries: []int{202},
			testDescription: "Should find song by entry number 202",
		},
		{
			name:            "Search by entry number prefix",
			searchPattern:   "303",
			expectedCount:   1,
			expectedEntries: []int{303},
			testDescription: "Should find songs by exact entry number match",
		},
		{
			name:            "Search by single-digit entry number (exact only)",
			searchPattern:   "1",
			expectedCount:   0,
			expectedEntries: []int{},
			testDescription: "Should not match when entry does not exactly equal 1",
		},
		{
			name:            "Search by two-digit entry number (exact only)",
			searchPattern:   "20",
			expectedCount:   0,
			expectedEntries: []int{},
			testDescription: "Should not match when entry does not exactly equal 20",
		},
		{
			name:            "Search with less than 3 characters (non-numeric) returns all",
			searchPattern:   "ab",
			expectedCount:   3,
			expectedEntries: []int{101, 202, 303},
			testDescription: "Should return all songs for non-numeric patterns less than 3 characters",
		},
		{
			name:            "Empty search returns all",
			searchPattern:   "",
			expectedCount:   3,
			expectedEntries: []int{101, 202, 303},
			testDescription: "Should return all songs for empty search pattern",
		},
		{
			name:            "Case-insensitive search",
			searchPattern:   "SPRING",
			expectedCount:   1,
			expectedEntries: []int{202},
			testDescription: "Should find songs regardless of case",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			songs, err := app.GetSongs("entry", tc.searchPattern)
			if err != nil {
				t.Fatalf("Failed to get songs: %v", err)
			}

			if len(songs) != tc.expectedCount {
				t.Errorf("%s: Expected %d songs, got %d", tc.testDescription, tc.expectedCount, len(songs))
				return
			}

			for i, expected := range tc.expectedEntries {
				if i >= len(songs) {
					t.Errorf("%s: Expected entry %d, but song at index %d does not exist", tc.testDescription, expected, i)
					continue
				}
				if songs[i].Entry != expected {
					t.Errorf("%s: Expected entry %d at position %d, got %d", tc.testDescription, expected, i, songs[i].Entry)
				}
			}
		})
	}
}

func TestGetSongs2WithSearch(t *testing.T) {
	app := setupTestDB(t)
	defer teardownTestDB(app)

	// Insert sample data into the database
	db, err := sql.Open("sqlite3", app.dbFilePath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
        INSERT INTO songs (title, title_d, verse_order, entry) VALUES
            ('Božena Němcová', 'Bozena Nemcova', '1', 505),
            ('Modern Song', 'modern song', '1', 606);
        INSERT INTO verses (song_id, name, lines, lines_d) VALUES
            (1, 'Verse 1', 'Božena verses', 'bozena verses'),
            (2, 'Verse 1', 'Modern times', 'modern times');
        INSERT INTO authors (song_id, author_type, author_value, author_value_d) VALUES
            (1, 'music', 'Česká hudba', 'ceska hudba'),
            (2, 'music', 'Contemporary Artist', 'contemporary artist');
    `)
	if err != nil {
		t.Fatalf("Failed to insert sample data: %v", err)
	}

	testCases := []struct {
		name            string
		searchPattern   string
		expectedCount   int
		expectedEntries []int
	}{
		{
			name:            "Search by diacritics title in GetSongs2",
			searchPattern:   "bozena",
			expectedCount:   1,
			expectedEntries: []int{505},
		},
		{
			name:            "Search by entry number in GetSongs2",
			searchPattern:   "606",
			expectedCount:   1,
			expectedEntries: []int{606},
		},
		{
			name:            "Search by author in GetSongs2",
			searchPattern:   "ceska",
			expectedCount:   1,
			expectedEntries: []int{505},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			songs, err := app.GetSongs2("entry", tc.searchPattern)
			if err != nil {
				t.Fatalf("Failed to get songs: %v", err)
			}

			if len(songs) != tc.expectedCount {
				t.Errorf("Expected %d songs, got %d", tc.expectedCount, len(songs))
				return
			}

			for i, expected := range tc.expectedEntries {
				if i >= len(songs) {
					t.Errorf("Expected entry %d, but song at index %d does not exist", expected, i)
					continue
				}
				if songs[i].Entry != expected {
					t.Errorf("Expected entry %d at position %d, got %d", expected, i, songs[i].Entry)
				}
			}
		})
	}
}
