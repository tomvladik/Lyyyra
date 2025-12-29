package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func (a *App) PrepareDatabase() {
	slog.Info(fmt.Sprintf("PrepareDatabase: %s", a.dbFilePath))
	// Open an SQLite database (file-based)
	db, err := sql.Open("sqlite3", a.dbFilePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Error opening database: %s", err))
		return
	}
	defer db.Close()

	// Create tables if they don't exist
	tableScripts := []string{
		`DROP TABLE IF EXISTS verses_fts;`,
		`DROP TABLE IF EXISTS authors_fts;`,
		`DROP TABLE IF EXISTS songs_fts;`,
		`DROP TABLE IF EXISTS authors;`,
		`DROP TABLE IF EXISTS verses;`,
		`DROP TABLE IF EXISTS songs;`,
		`CREATE TABLE IF NOT EXISTS songs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            entry INTEGER,
            title TEXT,
            title_d TEXT,
            verse_order TEXT
        ); DELETE FROM songs;`,

		`CREATE VIRTUAL TABLE IF NOT EXISTS songs_fts USING fts5 (
            id,
            title,
            content=songs,
            content_rowid=id
        ); DELETE FROM songs_fts;`,

		`CREATE TABLE IF NOT EXISTS authors (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            song_id INTEGER,
            author_type TEXT,
            author_value TEXT,
            author_value_d TEXT,
            FOREIGN KEY(song_id) REFERENCES songs(id)
            ON DELETE CASCADE
        ); DELETE FROM authors;`,

		`CREATE VIRTUAL TABLE IF NOT EXISTS authors_fts USING fts5 (
            id,
            author_value,
            content=authors,
            content_rowid=id
        ); DELETE FROM authors_fts;`,

		`CREATE TABLE IF NOT EXISTS verses (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            song_id INTEGER,
            name TEXT,
            lines TEXT,
            lines_d TEXT,
            FOREIGN KEY(song_id) REFERENCES songs(id)
            ON DELETE CASCADE
        ); DELETE FROM verses;`,

		`CREATE VIRTUAL TABLE IF NOT EXISTS verses_fts USING fts5 (
            id,
            lines,
            content=verses,
            content_rowid=id
        ); DELETE FROM verses_fts;`,
	}

	// Execute each table creation script
	for _, script := range tableScripts {
		if _, err := db.Exec(script); err != nil {
			slog.Error(fmt.Sprintf("Error executing script: %s", err))
			return
		}
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
	if a.testRun {
		xmlFiles = xmlFiles[:25]
	}
	// Process each XML file
	for _, xmlFile := range xmlFiles {
		// Construct the full path to the XML file
		// Read XML data from file
		xmlFilePath := fmt.Sprintf("%s/%s", a.songBookDir, xmlFile.Name())

		song, err := parseXmlSong(xmlFilePath)
		if err != nil {
			continue
		}

		title_d := removeDiacritics(song.Title)
		// Insert song data into the songs table
		result, err := db.Exec(`INSERT INTO songs (title, title_d, verse_order, entry) VALUES (?, ?, ?, ?)`,
			song.Title, title_d, song.VerseOrder, song.Songbook.Entry)
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

		_, err = db.Exec(`INSERT INTO songs_fts (id, title) VALUES (?, ?)`,
			songID, title_d)
		if err != nil {
			slog.Error(fmt.Sprintf("Error inserting song index for file %s: %v\n", xmlFile.Name(), err))
			continue
		}

		// Insert author data into the authors table
		for _, author := range song.Authors {
			author_d := removeDiacritics(author.Value)
			result, err := db.Exec(`INSERT INTO authors (song_id, author_type, author_value, author_value_d) VALUES (?, ?, ?, ?)`,
				songID, author.Type, author.Value, author_d)
			if err != nil {
				slog.Error(fmt.Sprintf("Error inserting author data for file %s: %v\n", xmlFile.Name(), err))
				continue
			}
			// Get the ID
			authorID, err := result.LastInsertId()
			if err != nil {
				slog.Error(fmt.Sprintf("Error getting last author insert ID for file %s: %v\n", xmlFile.Name(), err))
				continue
			}
			_, err = db.Exec(`INSERT INTO authors_fts (id, author_value) VALUES (?, ?)`,
				authorID, author_d)
			if err != nil {
				slog.Error(fmt.Sprintf("Error inserting author index for file %s: %v\n", xmlFile.Name(), err))
				continue
			}

		}

		// Insert verse data into the verses table
		for _, verse := range song.Lyrics.Verses {
			lines_d := removeDiacritics(verse.Lines)
			result, err := db.Exec(`INSERT INTO verses (song_id, name, lines, lines_d) VALUES (?, ?, ?, ?)`,
				songID, verse.Name, verse.Lines, lines_d)
			if err != nil {
				slog.Error(fmt.Sprintf("Error inserting verse data for file %s: %v\n", xmlFile.Name(), err))
				continue
			}
			// Get the ID
			verseID, err := result.LastInsertId()
			if err != nil {
				slog.Error(fmt.Sprintf("Error getting last verse insert ID for file %s: %v\n", xmlFile.Name(), err))
				continue
			}
			_, err = db.Exec(`INSERT INTO verses_fts (id, lines) VALUES (?, ?)`,
				verseID, lines_d)
			if err != nil {
				slog.Error(fmt.Sprintf("Error inserting author index for file %s: %v\n", xmlFile.Name(), err))
				continue
			}
		}

		slog.Debug(fmt.Sprintf("Data inserted  %s : %s file %s\n", song.Songbook.Entry, song.Title, xmlFile.Name()))
	}
}

func (a *App) GetSongs(orderBy string, searchPattern string) ([]dtoSong, error) {
	var result []dtoSong
	db, err := sql.Open("sqlite3", a.dbFilePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Error opening database: %s", err))
		a.status.DatabaseReady = false
		return result, err
	}
	defer db.Close()

	query_pre := `
    SELECT s.id,
        entry,
        title,
    REPLACE(GROUP_CONCAT(lines, char(10)), '<br />', '===') AS all_verses,
    COALESCE((SELECT author_value
            FROM authors
            WHERE song_id = s.id AND author_type = 'music'
            ORDER BY id LIMIT 1),'') AS authorMusic,
    COALESCE((SELECT author_value
            FROM authors
            WHERE song_id = s.id AND author_type = 'words'
            ORDER BY id LIMIT 1),'') AS authorLyric
  FROM songs s
  JOIN verses v ON s.id = v.song_id
`

	// query_middle := fmt.Sprintf(`
	// JOIN verses_fts vfts ON v.id = vfts.id
	// JOIN authors_fts afts ON authors.id=afts.id
	// WHERE vfts MATCH '%s'
	// `, searchPattern)

	query_post := `
GROUP BY
       s.id,
     entry,
     title
order by ` + orderBy + `, v.name`
	// Perform a full-text search on the lyrics
	//searchTerm := "your_search_term_here"
	rows, err := db.Query(query_pre + query_post)
	if err != nil {
		slog.Error(fmt.Sprintf("Error querying data: %s", err))
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			title, allVerses, authorMusic, authorLyric string
			id, entry                                  int
		)
		err := rows.Scan(&id, &entry, &title, &allVerses, &authorMusic, &authorLyric)
		if err != nil {
			slog.Error(fmt.Sprintf("Error scanning row: %s", err))
			return result, err
		}

		result = append(result, dtoSong{Id: id, Entry: entry, Title: title, Verses: allVerses, AuthorMusic: authorMusic, AuthorLyric: authorLyric})
	}
	return result, nil
}

func (a *App) GetSongs2(orderBy string, searchPattern string) ([]dtoSongHeader, error) {
	var result []dtoSongHeader
	db, err := sql.Open("sqlite3", a.dbFilePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Error opening database: %s", err))
		a.status.DatabaseReady = false
		return result, err
	}
	defer db.Close()

	query_pre := `
SELECT DISTINCT s.id,
       entry,
       s.title,
       title_d,
       verse_order
  FROM songs s
--  JOIN verses v ON v.song_id = s.id
--  JOIN authors a ON a.song_id = s.id
`

	//  query_middle := fmt.Sprintf(`
	//  JOIN verses vfts ON v.id = vfts.id
	//  JOIN authors_fts afts ON authors.id=afts.id
	//  WHERE vfts MATCH '%s'
	//  `, searchPattern)

	query_post := `
order by ` + orderBy
	// Perform a full-text search on the lyrics
	//searchTerm := "your_search_term_here"
	query := query_pre + query_post
	rows, err := db.Query(query)
	if err != nil {
		slog.Error(fmt.Sprintf("Error querying data: %s for: %s", err, query))
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			title, title_d, verse_order sql.NullString
			id, entry                   int
		)
		err := rows.Scan(&id, &entry, &title, &title_d, &verse_order)
		if err != nil {
			slog.Error(fmt.Sprintf("Error scanning row: %s", err))
			return result, err
		}

		result = append(result, dtoSongHeader{Id: id, Entry: entry, Title: title.String, TitleD: title_d.String})
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
