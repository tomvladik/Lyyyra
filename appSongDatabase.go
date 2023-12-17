package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func (a *App) prepareDatabase() {
	// Open an SQLite database (file-based)
	db, err := sql.Open("sqlite3", a.dbFilePath)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	// Create tables if they don't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS songs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entry INTEGER,
			title TEXT,
			verse_order TEXT
		)
	`)
	if err != nil {
		fmt.Println("Error creating songs table:", err)
		return
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS authors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		song_id INTEGER,
		author_type TEXT,
		author_value TEXT,
		FOREIGN KEY(song_id) REFERENCES songs(id)
	)
`)
	if err != nil {
		fmt.Println("Error creating authors table:", err)
		return
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS verses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			song_id INTEGER,
			name TEXT,
			lines TEXT,
			FOREIGN KEY(song_id) REFERENCES songs(id)
		)
	`)
	if err != nil {
		fmt.Println("Error creating verses table:", err)
		return
	}
}

func (a *App) fillDatabase() {
	db, err := sql.Open("sqlite3", a.dbFilePath)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	// Read all XML files in the specified directory
	xmlFiles, err := os.ReadDir(a.songBookDir)
	if err != nil {
		fmt.Println("Error reading XML files directory:", err)
		return
	}
	defer os.RemoveAll(a.songBookDir)
	// Process each XML file
	for _, xmlFile := range xmlFiles {
		// Construct the full path to the XML file
		xmlFilePath := fmt.Sprintf("%s/%s", a.songBookDir, xmlFile.Name())

		// Read XML data from file
		xmlData, err := os.ReadFile(xmlFilePath)
		if err != nil {
			fmt.Printf("Error reading XML file %s: %v\n", xmlFile.Name(), err)
			continue
		}

		var song Song
		err = xml.Unmarshal(xmlData, &song)
		if err != nil {
			fmt.Printf("Error unmarshalling XML in file %s: %v\n", xmlFile.Name(), err)
			continue
		}

		// Insert song data into the songs table
		result, err := db.Exec(`
		INSERT INTO songs (title, verse_order, entry) VALUES (?, ?, ?)
	`, song.Title, song.VerseOrder, song.Songbook.Entry)
		if err != nil {
			fmt.Printf("Error inserting song data for file %s: %v\n", xmlFile.Name(), err)
			continue
		}

		// Get the ID of the inserted song
		songID, err := result.LastInsertId()
		if err != nil {
			fmt.Printf("Error getting last insert ID for file %s: %v\n", xmlFile.Name(), err)
			continue
		}

		// Insert author data into the authors table
		for _, author := range song.Authors {
			_, err := db.Exec(`
			INSERT INTO authors (song_id, author_type, author_value) VALUES (?, ?, ?)
		`, songID, author.Type, author.Value)
			if err != nil {
				fmt.Printf("Error inserting author data for file %s: %v\n", xmlFile.Name(), err)
				continue
			}
		}

		// Insert verse data into the verses table
		for _, verse := range song.Lyrics.Verses {
			_, err := db.Exec(`
			INSERT INTO verses (song_id, name, lines) VALUES (?, ?, ?)
		`, songID, verse.Name, verse.Lines)
			if err != nil {
				fmt.Printf("Error inserting verse data for file %s: %v\n", xmlFile.Name(), err)
				continue
			}
		}

		fmt.Printf("Data inserted for file %s\n", xmlFile.Name())
	}
}
