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
	err = db.QueryRow(`SELECT title FROM songs WHERE title='ABCD'`).Scan(&title)
	if err != nil {
		t.Errorf("Sample song was not inserted: %v", err)
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
        INSERT INTO songs (title, verse_order, entry) VALUES ('Sample Song', '1', 1);
        INSERT INTO verses (song_id, name, lines) VALUES (1, 'Verse 1', 'Sample lines');
    `)
	if err != nil {
		t.Fatalf("Failed to insert sample data: %v", err)
	}

	// Get songs from the database
	songs, err := app.GetSongs("title")
	if err != nil {
		t.Errorf("Failed to get songs: %v", err)
	}

	// Check if the sample song is in the result
	if len(songs) != 1 || songs[0].Title != "Sample Song" {
		t.Errorf("Expected to get 'Sample Song', got %v", songs)
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
	if len(authors) != 1 || authors[0].Value != "Sample Author" {
		t.Errorf("Expected to get 'Sample Author', got %v", authors)
	}
}
