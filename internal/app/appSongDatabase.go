package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func (a *App) PrepareDatabase() {
	slog.Info(fmt.Sprintf("PrepareDatabase: %s", a.dbFilePath))

	_ = a.withDB(func(db *sql.DB) error {
		// Create tables if they don't exist
		tableScripts := []string{
			`DROP TABLE IF EXISTS authors;`,
			`DROP TABLE IF EXISTS verses;`,
			`DROP TABLE IF EXISTS songs;`,
			`CREATE TABLE IF NOT EXISTS songs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            entry INTEGER,
            title TEXT,
            title_d TEXT,
            verse_order TEXT,
            kytara_file TEXT
        ); DELETE FROM songs;`,
			`CREATE TABLE IF NOT EXISTS authors (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            song_id INTEGER,
            author_type TEXT,
            author_value TEXT,
            author_value_d TEXT,
            FOREIGN KEY(song_id) REFERENCES songs(id)
            ON DELETE CASCADE
        ); DELETE FROM authors;`,
			`CREATE TABLE IF NOT EXISTS verses (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            song_id INTEGER,
            name TEXT,
            lines TEXT,
            lines_d TEXT,
            FOREIGN KEY(song_id) REFERENCES songs(id)
            ON DELETE CASCADE
        ); DELETE FROM verses;`,
		}

		// Execute each table creation script
		for _, script := range tableScripts {
			if _, err := db.Exec(script); err != nil {
				slog.Error(fmt.Sprintf("Error executing script: %s", err))
				return err
			}
		}
		return nil
	})
}

func (a *App) FillDatabase() {
	xmlFiles, err := a.getXmlFiles()
	if err != nil {
		a.status.SongsReady = false
		return
	}

	totalFiles := len(xmlFiles)
	a.updateProgress("Plním databázi...", 0)

	_ = a.withDB(func(db *sql.DB) error {
		for i, xmlFile := range xmlFiles {
			a.updateDatabaseProgress(i, totalFiles)
			if err := a.processSongFile(db, xmlFile); err != nil {
				slog.Error("Failed to process song file", "file", xmlFile.Name(), "error", err)
			}
		}
		return nil
	})
}

// getXmlFiles reads and validates XML files from the songbook directory
func (a *App) getXmlFiles() ([]os.DirEntry, error) {
	xmlFiles, err := os.ReadDir(a.songBookDir)
	if err != nil {
		slog.Error("Error reading XML files directory", "error", err)
		return nil, err
	}
	if len(xmlFiles) == 0 {
		slog.Error("No XML files in directory", "directory", a.songBookDir)
		return nil, fmt.Errorf("no XML files found")
	}

	if a.testRun && len(xmlFiles) > 25 {
		xmlFiles = xmlFiles[:25]
	}

	return xmlFiles, nil
}

// updateDatabaseProgress updates progress every 10 files or at the end
func (a *App) updateDatabaseProgress(current, total int) {
	if current%10 == 0 || current == total-1 {
		percent := int((float64(current+1) / float64(total)) * 100)
		message := fmt.Sprintf("Plním databázi... (%d/%d)", current+1, total)
		a.updateProgress(message, percent)
	}
}

// processSongFile parses an XML file and inserts song data into the database
func (a *App) processSongFile(db *sql.DB, xmlFile os.DirEntry) error {
	xmlFilePath := filepath.Join(a.songBookDir, xmlFile.Name())

	song, err := parseXmlSong(xmlFilePath)
	if err != nil {
		return fmt.Errorf("failed to parse XML: %w", err)
	}

	songID, err := a.insertSong(db, song)
	if err != nil {
		return fmt.Errorf("failed to insert song: %w", err)
	}

	if err := a.insertAuthors(db, songID, song.Authors, xmlFile.Name()); err != nil {
		return fmt.Errorf("failed to insert authors: %w", err)
	}

	if err := a.insertVerses(db, songID, song.Lyrics.Verses, xmlFile.Name()); err != nil {
		return fmt.Errorf("failed to insert verses: %w", err)
	}

	slog.Debug("Data inserted", "entry", song.Songbook.Entry, "title", song.Title, "file", xmlFile.Name())
	return nil
}

// insertSong inserts a song record and returns the song ID
func (a *App) insertSong(db *sql.DB, song *Song) (int64, error) {
	title_d := removeDiacritics(song.Title)
	result, err := db.Exec(`INSERT INTO songs (title, title_d, verse_order, entry) VALUES (?, ?, ?, ?)`,
		song.Title, title_d, song.VerseOrder, song.Songbook.Entry)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// insertAuthors inserts all author records for a song
func (a *App) insertAuthors(db *sql.DB, songID int64, authors []Author, filename string) error {
	for _, author := range authors {
		author_d := removeDiacritics(author.Value)
		_, err := db.Exec(`INSERT INTO authors (song_id, author_type, author_value, author_value_d) VALUES (?, ?, ?, ?)`,
			songID, author.Type, author.Value, author_d)
		if err != nil {
			slog.Error("Error inserting author", "file", filename, "error", err)
			continue
		}
	}
	return nil
}

// insertVerses inserts all verse records for a song
func (a *App) insertVerses(db *sql.DB, songID int64, verses []Verse, filename string) error {
	for _, verse := range verses {
		lines_d := removeDiacritics(verse.Lines)
		_, err := db.Exec(`INSERT INTO verses (song_id, name, lines, lines_d) VALUES (?, ?, ?, ?)`,
			songID, verse.Name, verse.Lines, lines_d)
		if err != nil {
			slog.Error("Error inserting verse", "file", filename, "error", err)
			continue
		}
	}
	return nil
}

func (a *App) GetSongs(orderBy string, searchPattern string) ([]dtoSong, error) {
	var result []dtoSong
	err := a.withDB(func(db *sql.DB) error {

		query_pre := `
    SELECT s.id,
        entry,
        title,
	GROUP_CONCAT(lines, char(10,10)) AS all_verses,
    COALESCE((SELECT author_value
            FROM authors
            WHERE song_id = s.id AND author_type = 'music'
            ORDER BY id LIMIT 1),'') AS authorMusic,
    COALESCE((SELECT author_value
            FROM authors
            WHERE song_id = s.id AND author_type = 'words'
            ORDER BY id LIMIT 1),'') AS authorLyric,
    COALESCE(kytara_file, '') AS kytara_file
  FROM songs s
  JOIN verses v ON s.id = v.song_id
`

		// query_middle := fmt.Sprintf(`
		// JOIN verses_fts vfts ON v.id = vfts.id
		// JOIN authors_fts afts ON authors.id=afts.id
		// WHERE vfts MATCH '%s'
		// `, searchPattern)

		sortOption := normalizeSortingOption(orderBy)
		orderColumn := orderColumnForSongs(sortOption)
		query_post := `
GROUP BY
			 s.id,
		 entry,
		 title
order by ` + orderColumn + `, v.name`
		// Perform a full-text search on the lyrics
		//searchTerm := "your_search_term_here"
		rows, err := db.Query(query_pre + query_post)
		if err != nil {
			slog.Error(fmt.Sprintf("Error querying data: %s", err))
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var (
				title, allVerses, authorMusic, authorLyric, kytaraFile string
				id, entry                                              int
			)
			err := rows.Scan(&id, &entry, &title, &allVerses, &authorMusic, &authorLyric, &kytaraFile)
			if err != nil {
				slog.Error(fmt.Sprintf("Error scanning row: %s", err))
				return err
			}

			result = append(result, dtoSong{Id: id, Entry: entry, Title: title, Verses: allVerses, AuthorMusic: authorMusic, AuthorLyric: authorLyric, KytaraFile: kytaraFile})
		}
		return nil
	})
	if err != nil {
		a.status.DatabaseReady = false
	}
	return result, err
}

func (a *App) GetSongs2(orderBy string, searchPattern string) ([]dtoSongHeader, error) {
	var result []dtoSongHeader
	err := a.withDB(func(db *sql.DB) error {

		query_pre := `
SELECT DISTINCT s.id,
       entry,
       s.title,
       title_d,
       verse_order,
       kytara_file
  FROM songs s
--  JOIN verses v ON v.song_id = s.id
--  JOIN authors a ON a.song_id = s.id
`

		//  query_middle := fmt.Sprintf(`
		//  JOIN verses vfts ON v.id = vfts.id
		//  JOIN authors_fts afts ON authors.id=afts.id
		//  WHERE vfts MATCH '%s'
		//  `, searchPattern)

		sortOption := normalizeSortingOption(orderBy)
		orderColumn := orderColumnForSongs2(sortOption)
		query_post := `
order by ` + orderColumn
		//searchTerm := "your_search_term_here"
		query := query_pre + query_post
		rows, err := db.Query(query)
		if err != nil {
			slog.Error(fmt.Sprintf("Error querying data: %s for: %s", err, query))
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var (
				title, title_d, verse_order, kytaraFile sql.NullString
				id, entry                               int
			)
			err := rows.Scan(&id, &entry, &title, &title_d, &verse_order, &kytaraFile)
			if err != nil {
				slog.Error(fmt.Sprintf("Error scanning row: %s", err))
				return err
			}

			result = append(result, dtoSongHeader{Id: id, Entry: entry, Title: title.String, TitleD: title_d.String, KytaraFile: kytaraFile.String})
		}
		return nil
	})
	if err != nil {
		a.status.DatabaseReady = false
	}
	return result, err
}

func (a *App) GetSongAuthors(songId int) ([]Author, error) {
	var result []Author
	err := a.withDB(func(db *sql.DB) error {

		// Perform a full-text search on the lyrics
		//searchTerm := "your_search_term_here"
		rows, err := db.Query(`
    SELECT DISTINCT author_type, author_value
    FROM authors
    WHERE song_id = ?
    ORDER BY author_type`, songId)
		if err != nil {
			slog.Error(fmt.Sprintf("Error querying data: %s", err))
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var (
				authType, authValue string
			)
			err := rows.Scan(&authType, &authValue)
			if err != nil {
				slog.Error(fmt.Sprintf("Error scanning row: %s", err))
				return err
			}

			result = append(result, Author{Type: authType, Value: authValue})
		}
		return nil
	})
	return result, err
}

// GetSongVerses returns the concatenated verses (lines) for a given song id.
// Verses are concatenated using '===' as separator to match frontend expectations.
func (a *App) GetSongVerses(songId int) (string, error) {
	var verses []string
	err := a.withDB(func(db *sql.DB) error {
		rows, err := db.Query(`SELECT lines FROM verses WHERE song_id = ? ORDER BY id`, songId)
		if err != nil {
			slog.Error(fmt.Sprintf("Error querying verses: %s", err))
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var lines string
			if err := rows.Scan(&lines); err != nil {
				slog.Error(fmt.Sprintf("Error scanning verse row: %s", err))
				return err
			}
			verses = append(verses, lines)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return strings.Join(verses, "==="), nil
}

// GetSongProjection returns JSON containing verse_order and verses (name + lines)
// Example: { "verse_order": "c v1 c v2", "verses": [{"name":"v1","lines":"..."}, ...] }
func (a *App) GetSongProjection(songId int) (string, error) {
	var verseOrder string
	// read verse_order from songs table
	err := a.withDB(func(db *sql.DB) error {
		row := db.QueryRow(`SELECT verse_order FROM songs WHERE id = ?`, songId)
		if err := row.Scan(&verseOrder); err != nil {
			if err == sql.ErrNoRows {
				verseOrder = ""
				return nil
			}
			slog.Error(fmt.Sprintf("Error reading verse_order: %s", err))
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	verses := []map[string]string{}
	err = a.withDB(func(db *sql.DB) error {
		rows, err := db.Query(`SELECT name, lines FROM verses WHERE song_id = ? ORDER BY id`, songId)
		if err != nil {
			slog.Error(fmt.Sprintf("Error querying verses: %s", err))
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var name, lines string
			if err := rows.Scan(&name, &lines); err != nil {
				slog.Error(fmt.Sprintf("Error scanning verse row: %s", err))
				return err
			}
			verses = append(verses, map[string]string{"name": name, "lines": lines})
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	payload := map[string]interface{}{
		"verse_order": verseOrder,
		"verses":      verses,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
