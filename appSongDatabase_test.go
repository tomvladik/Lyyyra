package main

import (
	"database/sql"
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

func TestDatabaseVersion(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var version string
	err = db.QueryRow("SELECT sqlite_version();").Scan(&version)
	if err != nil {
		t.Fatal(err)
	}

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
		t.Errorf("Sample song was not inserted: %v", err)
	}

	// Check if the song was indexed in FTS
	var found_entry string
	err = db.QueryRow(`
		SELECT entry FROM songs
		WHERE title_d LIKE '%abccdde%'`).Scan(&found_entry)
	if err != nil {
		t.Errorf("Sample song was not inserted: %v", err)
	}

	// Check if the sample author is in the result
	if found_entry != "288" {
		t.Errorf("Expected to get other song, not %v", found_entry)
	}

	err = db.QueryRow(`
    SELECT entry FROM songs
	JOIN authors ON songs.id=authors.song_id
	WHERE authors.author_value_d LIKE '%2018%'`).Scan(&found_entry)
	if err != nil {
		t.Errorf("Sample song was not inserted: %v", err)
	}

	// Check if the sample author is in the result
	if found_entry != "3" {
		t.Errorf("Expected to get other song, not %v", found_entry)
	}

	err = db.QueryRow(`
    SELECT entry FROM songs
	JOIN verses ON songs.id=verses.song_id
	WHERE verses.lines_d LIKE '%tulen%'`).Scan(&found_entry)
	if err != nil {
		t.Errorf("Sample song was not inserted: %v", err)
	}

	// Check if the sample author is in the result
	if found_entry != "4" {
		t.Errorf("Expected to get other song, not %v", found_entry)
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
        INSERT INTO authors (song_id, author_type, author_value) VALUES (1, 'wrods', 'Bedřich Antonn Leoš');
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
