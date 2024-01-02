package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func (a *App) PrepareDatabase() {
	// Open an SQLite database (file-based)
	db, err := sql.Open("sqlite3", a.dbFilePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Error opening database: %s", err))
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
		);
		DELETE FROM songs;
	`)
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating songs table: %s", err))
		return
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS authors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		song_id INTEGER,
		author_type TEXT,
		author_value TEXT,
		FOREIGN KEY(song_id) REFERENCES songs(id)
		ON DELETE CASCADE
	);
	DELETE FROM authors;
`)
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating authors table: %s", err))
		return
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS verses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			song_id INTEGER,
			name TEXT,
			lines TEXT,
			FOREIGN KEY(song_id) REFERENCES songs(id)
			ON DELETE CASCADE
		);
		DELETE FROM verses;
	`)
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating verses table: %s", err))
		return
	}
}

func (a *App) FillDatabase() {
	db, err := sql.Open("sqlite3", a.dbFilePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Error opening database: %s", err))
		return
	}
	defer db.Close()

	// Read all XML files in the specified directory
	xmlFiles, err := os.ReadDir(a.songBookDir)
	if err != nil {
		slog.Error(fmt.Sprintf("Error reading XML files directory: %s", err))
		a.status.SongsReady = false
		return
	}
	if len(xmlFiles) == 0 {
		slog.Error(fmt.Sprintf("No XML files in directory: %s", a.songBookDir))
		a.status.SongsReady = false
		return
	}
	//defer os.RemoveAll(a.songBookDir)

	// Process each XML file
	for _, xmlFile := range xmlFiles {
		// Construct the full path to the XML file
		// Read XML data from file
		xmlFilePath := fmt.Sprintf("%s/%s", a.songBookDir, xmlFile.Name())

		song, err := parseXmlSong(xmlFilePath)
		if err != nil {
			continue
		}

		// Insert song data into the songs table
		result, err := db.Exec(`
		INSERT INTO songs (title, verse_order, entry) VALUES (?, ?, ?)
	`, song.Title, song.VerseOrder, song.Songbook.Entry)
		if err != nil {
			slog.Error(fmt.Sprintf("Error inserting song data for file %s: %v\n", xmlFile.Name(), err))
			continue
		}

		// Get the ID of the inserted song
		songID, err := result.LastInsertId()
		if err != nil {
			slog.Error(fmt.Sprintf("Error getting last insert ID for file %s: %v\n", xmlFile.Name(), err))
			continue
		}

		// Insert author data into the authors table
		for _, author := range song.Authors {
			_, err := db.Exec(`
			INSERT INTO authors (song_id, author_type, author_value) VALUES (?, ?, ?)
		`, songID, author.Type, author.Value)
			if err != nil {
				slog.Error(fmt.Sprintf("Error inserting author data for file %s: %v\n", xmlFile.Name(), err))
				continue
			}
		}

		// Insert verse data into the verses table
		for _, verse := range song.Lyrics.Verses {
			_, err := db.Exec(`
			INSERT INTO verses (song_id, name, lines) VALUES (?, ?, ?)
		`, songID, verse.Name, verse.Lines)
			if err != nil {
				slog.Error(fmt.Sprintf("Error inserting verse data for file %s: %v\n", xmlFile.Name(), err))
				continue
			}
		}

		slog.Debug(fmt.Sprintf("Data inserted  %s : %s file %s\n", song.Songbook.Entry, song.Title, xmlFile.Name()))
	}
}

func (a *App) GetSongs() ([]dtoSong, error) {
	var result []dtoSong
	db, err := sql.Open("sqlite3", a.dbFilePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Error opening database: %s", err))
		a.status.DatabaseReady = false
		return result, err
	}
	defer db.Close()

	// Perform a full-text search on the lyrics
	//searchTerm := "your_search_term_here"
	rows, err := db.Query(`
SELECT s.id,
       entry,
       title,
	   REPLACE(GROUP_CONCAT(lines, char(10)), '<br />', '===') AS all_verses
  FROM songs s, verses v
  WHERE s.id=v.song_id
  GROUP BY
  	   s.id,
       entry,
       title
  order by entry, v.name`)
	if err != nil {
		slog.Error(fmt.Sprintf("Error querying data: %s", err))
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			title, allVerses string
			id, entry        int
		)
		err := rows.Scan(&id, &entry, &title, &allVerses)
		if err != nil {
			slog.Error(fmt.Sprintf("Error scanning row: %s", err))
			return result, err
		}

		result = append(result, dtoSong{Id: id, Entry: entry, Title: title, Verses: allVerses})
	}
	return result, nil
}

func (a *App) GetSongAuthors(songId int) ([]Author, error) {
	var result []Author
	db, err := sql.Open("sqlite3", a.dbFilePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Error opening database: %s", err))
		return result, err
	}
	defer db.Close()

	// Perform a full-text search on the lyrics
	//searchTerm := "your_search_term_here"
	rows, err := db.Query(`
	SELECT DISTINCT author_type, author_value 
	FROM authors
	WHERE song_id = ?
	ORDER BY author_type`, songId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error querying data: %s", err))
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			authType, authValue string
		)
		err := rows.Scan(&authType, &authValue)
		if err != nil {
			slog.Error(fmt.Sprintf("Error scanning row: %s", err))
			return result, err
		}

		result = append(result, Author{Type: authType, Value: authValue})
	}
	return result, nil
}
