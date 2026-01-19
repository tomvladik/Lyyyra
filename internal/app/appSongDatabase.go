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

// Current database schema version
const CurrentDBVersion = 2

func (a *App) PrepareDatabase() {
	slog.Info(fmt.Sprintf("PrepareDatabase: %s", a.dbFilePath))

	_ = a.withDB(func(db *sql.DB) error {
		// Detect current database version
		currentVersion, err := a.detectSchemaVersion(db)
		if err != nil {
			slog.Error(fmt.Sprintf("Error detecting schema version: %s", err))
			return err
		}

		slog.Info(fmt.Sprintf("Current schema version: %d, expected: %d", currentVersion, CurrentDBVersion))

		// Apply migrations if needed
		if currentVersion == 0 {
			// Fresh database - create v1 schema
			slog.Info("Creating new database with schema v1")
			if err := a.createSchemaV1(db); err != nil {
				slog.Error(fmt.Sprintf("Error creating schema v1: %s", err))
				return err
			}
			currentVersion = 1
		}

		if currentVersion < CurrentDBVersion {
			for v := currentVersion + 1; v <= CurrentDBVersion; v++ {
				slog.Info(fmt.Sprintf("Applying migration to version %d", v))
				if err := a.applyMigration(db, v); err != nil {
					slog.Error(fmt.Sprintf("Error applying migration to version %d: %s", v, err))
					return err
				}
			}
		}
		return nil
	})
}

// detectSchemaVersion detects the current schema version
// Returns 0 if no tables exist (fresh database - v0)
// Returns 1 if songs table exists but no schema_version table (old release - v1)
// Returns the version from schema_version table if it exists (v2+)
func (a *App) detectSchemaVersion(db *sql.DB) (int, error) {
	// Check if schema_version table exists
	var schemaVersionExists int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND name='schema_version'
	`).Scan(&schemaVersionExists)
	if err != nil {
		return 0, err
	}

	if schemaVersionExists > 0 {
		// schema_version table exists - read the version (v2+)
		var version int
		row := db.QueryRow(`SELECT MAX(version) FROM schema_version`)
		err = row.Scan(&version)
		if err == sql.ErrNoRows {
			return 1, nil
		}
		return version, err
	}

	// schema_version doesn't exist - check if songs table exists (V1 indicator)
	var songsExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND name='songs'
	`).Scan(&songsExists)
	if err != nil {
		return 0, err
	}

	if songsExists > 0 {
		// songs table exists but no schema_version - this is v1 (old release)
		return 1, nil
	}

	// No tables at all - fresh database (v0)
	return 0, nil
}

// applyMigration applies a specific migration version
func (a *App) applyMigration(db *sql.DB, version int) error {
	switch version {
	case 2:
		return a.migrateToV2(db)
	default:
		return fmt.Errorf("unknown migration version: %d", version)
	}
}

// ============ SCHEMA V1 (Original Release) ============
// createSchemaV1 creates the initial database schema (v1)
// This is the original schema from the first release - simple without songbook support
func (a *App) createSchemaV1(db *sql.DB) error {
	tableScripts := []string{
		`CREATE TABLE IF NOT EXISTS songs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            entry INTEGER,
            title TEXT,
            title_d TEXT,
            verse_order TEXT,
            kytara_file TEXT
        );`,
		`CREATE TABLE IF NOT EXISTS authors (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            song_id INTEGER,
            author_type TEXT,
            author_value TEXT,
            author_value_d TEXT,
            FOREIGN KEY(song_id) REFERENCES songs(id)
            ON DELETE CASCADE
        );`,
		`CREATE TABLE IF NOT EXISTS verses (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            song_id INTEGER,
            name TEXT,
            lines TEXT,
            lines_d TEXT,
            FOREIGN KEY(song_id) REFERENCES songs(id)
            ON DELETE CASCADE
        );`,
		`CREATE INDEX IF NOT EXISTS idx_songs_entry ON songs(entry);`,
		`CREATE INDEX IF NOT EXISTS idx_songs_title_d ON songs(title_d);`,
		`CREATE INDEX IF NOT EXISTS idx_authors_song_id ON authors(song_id);`,
		`CREATE INDEX IF NOT EXISTS idx_authors_value_d ON authors(author_value_d);`,
		`CREATE INDEX IF NOT EXISTS idx_verses_song_id ON verses(song_id);`,
		`CREATE INDEX IF NOT EXISTS idx_verses_lines_d ON verses(lines_d);`,
	}

	for _, script := range tableScripts {
		if _, err := db.Exec(script); err != nil {
			return fmt.Errorf("error creating v1 schema: %w", err)
		}
	}

	return nil
}

// ============ SCHEMA V2 (Migration) ============
// migrateToV2 upgrades from v1 to v2
// Changes:
// - Adds songbooks table for multi-songbook support
// - Adds song_book_id column to songs table
// - Adds schema_version table for tracking migrations
func (a *App) migrateToV2(db *sql.DB) error {
	slog.Info("Migrating to schema v2")

	// Create songbooks table
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS songbooks (
			acronym TEXT PRIMARY KEY NOT NULL,
			name TEXT NOT NULL,
			CHECK(length(acronym) <= 10)
		);
	`); err != nil {
		return fmt.Errorf("error creating songbooks table: %w", err)
	}

	// Add schema_version table
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return fmt.Errorf("error creating schema_version table: %w", err)
	}

	// Add song_book_id column to songs table
	if err := a.addColumnIfNotExists(db, "songs", "song_book_id", "TEXT"); err != nil {
		return fmt.Errorf("error adding song_book_id column: %w", err)
	}

	// Add index for song_book_id
	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_songs_song_book_id ON songs(song_book_id);
	`); err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}

	// Insert default songbook for existing data
	if _, err := db.Exec(`
		INSERT OR IGNORE INTO songbooks (acronym, name) VALUES (?, ?)
	`, "EZ", "Evangelický zpěvník 2021"); err != nil {
		return fmt.Errorf("error inserting default songbook: %w", err)
	}

	// Update existing songs to reference the default songbook
	if _, err := db.Exec(`
		UPDATE songs SET song_book_id = ? WHERE song_book_id IS NULL
	`, "EZ"); err != nil {
		return fmt.Errorf("error updating songs with default songbook: %w", err)
	}

	// Make song_book_id NOT NULL after populating it
	// Note: SQLite doesn't support ALTER COLUMN directly, so we keep it nullable or use a migration strategy
	// For now, we'll just add a constraint check in the application

	// Record that this version was applied
	_, err := db.Exec(`INSERT INTO schema_version (version) VALUES (2)`)
	return err
}

// ============ HELPER FUNCTIONS ============
// columnExists checks if a column exists in a table
func (a *App) columnExists(db *sql.DB, table, column string) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?
	`, table, column).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// addColumnIfNotExists adds a column to a table only if it doesn't already exist
// Use this in migration functions when adding new columns to existing tables
func (a *App) addColumnIfNotExists(db *sql.DB, table, column, columnDef string) error {
	exists, err := a.columnExists(db, table, column)
	if err != nil {
		return err
	}
	if exists {
		slog.Debug(fmt.Sprintf("Column %s already exists in table %s, skipping", column, table))
		return nil
	}

	alter := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, table, column, columnDef)
	_, err = db.Exec(alter)
	return err
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
		// Check if we're on v2 (has song_book_id column)
		hasMultiSongbook, err := a.columnExists(db, "songs", "song_book_id")
		if err != nil {
			slog.Error("Failed to check schema version", "error", err)
			return err
		}

		var songbookAcronym string
		if hasMultiSongbook {
			// V2+ schema - use songbook support
			acronym, err := a.getOrCreateSongbook(db, "EZ", "Evangelický zpěvník 2021")
			if err != nil {
				slog.Error("Failed to get or create songbook", "error", err)
				return err
			}
			songbookAcronym = acronym
		}

		for i, xmlFile := range xmlFiles {
			a.updateDatabaseProgress(i, totalFiles)
			if err := a.processSongFile(db, xmlFile, songbookAcronym); err != nil {
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
func (a *App) processSongFile(db *sql.DB, xmlFile os.DirEntry, songbookAcronym string) error {
	xmlFilePath := filepath.Join(a.songBookDir, xmlFile.Name())

	song, err := parseXmlSong(xmlFilePath)
	if err != nil {
		return fmt.Errorf("failed to parse XML: %w", err)
	}

	songID, err := a.insertSong(db, song, songbookAcronym)
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
// In V1, song_book_id is empty; in V2+, it contains the songbook acronym
func (a *App) insertSong(db *sql.DB, song *Song, songbookAcronym string) (int64, error) {
	title_d := removeDiacritics(song.Title)

	// Check if song_book_id column exists (V2+ schema)
	hasMultiSongbook, err := a.columnExists(db, "songs", "song_book_id")
	if err != nil {
		return 0, err
	}

	var result sql.Result
	if hasMultiSongbook {
		// V2+ - insert with song_book_id
		result, err = db.Exec(`INSERT INTO songs (song_book_id, title, title_d, verse_order, entry) VALUES (?, ?, ?, ?, ?)`,
			songbookAcronym, song.Title, title_d, song.VerseOrder, song.Songbook.Entry)
	} else {
		// V1 - insert without song_book_id
		result, err = db.Exec(`INSERT INTO songs (title, title_d, verse_order, entry) VALUES (?, ?, ?, ?)`,
			song.Title, title_d, song.VerseOrder, song.Songbook.Entry)
	}

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

// getOrCreateSongbook retrieves or creates a songbook by acronym (max 10 chars)
func (a *App) getOrCreateSongbook(db *sql.DB, acronym string, name string) (string, error) {
	if len(acronym) > 10 {
		return "", fmt.Errorf("songbook acronym must be 10 characters or less")
	}

	// Try to get existing songbook
	row := db.QueryRow(`SELECT acronym FROM songbooks WHERE acronym = ?`, acronym)
	var existingAcronym string
	err := row.Scan(&existingAcronym)
	if err == nil {
		return existingAcronym, nil // Found existing songbook
	}
	if err != sql.ErrNoRows {
		return "", err // Other error
	}

	// Create new songbook
	_, err = db.Exec(`INSERT INTO songbooks (acronym, name) VALUES (?, ?)`, acronym, name)
	if err != nil {
		return "", err
	}
	return acronym, nil
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

		// Build WHERE clause if searchPattern provided
		query_where := ""
		searchPatternD := ""
		trimmedPattern := strings.TrimSpace(searchPattern)

		if len(trimmedPattern) > 0 {
			// Check if it's a numeric search (for entry numbers)
			isNumeric := true
			for _, ch := range trimmedPattern {
				if ch < '0' || ch > '9' {
					isNumeric = false
					break
				}
			}

			if isNumeric {
				// Allow entry number search for any length
				query_where = `WHERE CAST(s.entry AS TEXT) = ?`
			} else if len(trimmedPattern) >= 3 {
				// Text search requires at least 3 characters
				searchPatternD = removeDiacritics(trimmedPattern)
				query_where = `
WHERE s.title_d LIKE ?
   OR EXISTS (SELECT 1 FROM authors a WHERE a.song_id = s.id AND a.author_value_d LIKE ?)
   OR v.lines_d LIKE ?
   OR CAST(s.entry AS TEXT) = ?
`
			}
		}

		sortOption := normalizeSortingOption(orderBy)
		orderColumn := orderColumnForSongs(sortOption)
		query_post := `
GROUP BY
			 s.id,
		 entry,
		 title
order by ` + orderColumn + `, v.name`

		fullQuery := query_pre + query_where + query_post

		var rows *sql.Rows
		var queryErr error
		if len(trimmedPattern) > 0 {
			// Check if numeric
			isNumeric := true
			for _, ch := range trimmedPattern {
				if ch < '0' || ch > '9' {
					isNumeric = false
					break
				}
			}

			if isNumeric {
				rows, queryErr = db.Query(fullQuery, trimmedPattern)
			} else if len(trimmedPattern) >= 3 {
				searchLike := "%" + searchPatternD + "%"
				rows, queryErr = db.Query(fullQuery, searchLike, searchLike, searchLike, trimmedPattern)
			} else {
				rows, queryErr = db.Query(fullQuery)
			}
		} else {
			rows, queryErr = db.Query(fullQuery)
		}
		if queryErr != nil {
			slog.Error(fmt.Sprintf("Error querying data: %s", queryErr))
			return queryErr
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
`

		// Build WHERE clause if searchPattern provided
		query_where := ""
		searchPatternD := ""
		trimmedPattern := strings.TrimSpace(searchPattern)

		if len(trimmedPattern) > 0 {
			// Check if it's a numeric search (for entry numbers)
			isNumeric := true
			for _, ch := range trimmedPattern {
				if ch < '0' || ch > '9' {
					isNumeric = false
					break
				}
			}

			if isNumeric {
				// Allow entry number search for any length
				query_where = `WHERE CAST(s.entry AS TEXT) = ?`
			} else if len(trimmedPattern) >= 3 {
				// Text search requires at least 3 characters
				searchPatternD = removeDiacritics(trimmedPattern)
				query_where = `
WHERE s.title_d LIKE ?
   OR EXISTS (SELECT 1 FROM authors a WHERE a.song_id = s.id AND a.author_value_d LIKE ?)
   OR EXISTS (SELECT 1 FROM verses v WHERE v.song_id = s.id AND v.lines_d LIKE ?)
   OR CAST(s.entry AS TEXT) = ?
`
			}
		}

		sortOption := normalizeSortingOption(orderBy)
		orderColumn := orderColumnForSongs2(sortOption)
		query_post := `
order by ` + orderColumn

		fullQuery := query_pre + query_where + query_post

		var rows *sql.Rows
		var queryErr error
		if len(trimmedPattern) > 0 {
			// Check if numeric
			isNumeric := true
			for _, ch := range trimmedPattern {
				if ch < '0' || ch > '9' {
					isNumeric = false
					break
				}
			}

			if isNumeric {
				rows, queryErr = db.Query(fullQuery, trimmedPattern)
			} else if len(trimmedPattern) >= 3 {
				searchLike := "%" + searchPatternD + "%"
				rows, queryErr = db.Query(fullQuery, searchLike, searchLike, searchLike, trimmedPattern)
			} else {
				rows, queryErr = db.Query(fullQuery)
			}
		} else {
			rows, queryErr = db.Query(fullQuery)
		}

		if queryErr != nil {
			slog.Error(fmt.Sprintf("Error querying data: %s", queryErr))
			return queryErr
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
