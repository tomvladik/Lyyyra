package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Current database schema version
const CurrentDBVersion = 2

// InitializeDatabase checks schema version and applies migrations.
// This is called on every app startup to ensure the database schema is up-to-date.
func (a *App) InitializeDatabase() {
	slog.Info(fmt.Sprintf("InitializeDatabase: %s", a.dbFilePath))

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
// - Adds songbook_acronym column to songs table
// - Adds schema_version table for tracking migrations
func (a *App) migrateToV2(db *sql.DB) error {
	slog.Info("Migrating to schema v2")

	// Create songbooks table
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS songbooks (
			songbook_acronym TEXT PRIMARY KEY NOT NULL,
			name TEXT NOT NULL,
			CHECK(length(songbook_acronym) <= 10)
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

	// Add songbook_acronym column to songs table
	if err := a.addColumnIfNotExists(db, "songs", "songbook_acronym", "TEXT"); err != nil {
		return fmt.Errorf("error adding songbook_acronym column: %w", err)
	}

	// Add index for songbook_acronym
	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_songs_songbook_acronym ON songs(songbook_acronym);
	`); err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}

	// Insert default songbook for existing data
	if _, err := db.Exec(`
		INSERT OR IGNORE INTO songbooks (songbook_acronym, name) VALUES (?, ?)
	`, "EZ", "Evangelický zpěvník 2021"); err != nil {
		return fmt.Errorf("error inserting default songbook: %w", err)
	}

	// Update existing songs to reference the default songbook
	if _, err := db.Exec(`
		UPDATE songs SET songbook_acronym = ? WHERE songbook_acronym IS NULL
	`, "EZ"); err != nil {
		return fmt.Errorf("error updating songs with default songbook: %w", err)
	}

	// Make songbook_acronym NOT NULL after populating it
	// Note: SQLite doesn't support ALTER COLUMN directly, so we keep it nullable or use a migration strategy
	// For now, we'll just add a constraint check in the application

	// Add entry_text column to songs table to store original song numbers with characters
	if err := a.addColumnIfNotExists(db, "songs", "entry_text", "TEXT"); err != nil {
		return fmt.Errorf("error adding entry_text column: %w", err)
	}

	// Create index for entry_text for efficient searching
	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_songs_entry_text ON songs(entry_text);
	`); err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}

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
	a.updateProgress("Plním databázi...", 0)

	_ = a.withDB(func(db *sql.DB) error {
		// Process EZ songbook
		if err := a.fillEZSongs(db); err != nil {
			slog.Error("Failed to fill EZ songs", "error", err)
			return err
		}

		// Process KK songbook
		if err := a.fillKKSongs(db); err != nil {
			slog.Error("Failed to fill KK songs", "error", err)
			return err
		}

		return nil
	})
}

// fillEZSongs processes Evangelický zpěvník songs
func (a *App) fillEZSongs(db *sql.DB) error {
	ezDir := filepath.Join(a.songBookDir, "EZ")
	xmlFiles, err := a.getXmlFilesFromDir(ezDir)
	if err != nil {
		return err
	}

	songbookAcronym, err := a.getOrCreateSongbook(db, Acronym_EZ, "Evangelický zpěvník 2021")
	if err != nil {
		return fmt.Errorf("failed to get or create EZ songbook: %w", err)
	}

	totalFiles := len(xmlFiles)
	for i, xmlFile := range xmlFiles {
		if i%10 == 0 || i == totalFiles-1 {
			percent := int((float64(i+1) / float64(totalFiles)) * 50) // 0-50%
			message := fmt.Sprintf("Plním EZ databázi... (%d/%d)", i+1, totalFiles)
			a.updateProgress(message, percent)
		}

		if err := a.processEZSongFile(db, xmlFile, songbookAcronym); err != nil {
			slog.Error("Failed to process EZ song file", "file", xmlFile.Name(), "error", err)
		}
	}
	return nil
}

// fillKKSongs processes Katolický kancionál songs
func (a *App) fillKKSongs(db *sql.DB) error {
	kkDir := filepath.Join(a.songBookDir, "KK", "Kancional")
	if _, err := os.Stat(kkDir); os.IsNotExist(err) {
		slog.Info("KK directory not found, skipping KK import", "dir", kkDir)
		return nil // Not an error - KK is optional
	}

	xmlFiles, err := a.getXmlFilesFromDir(kkDir)
	if err != nil {
		return err
	}

	songbookAcronym, err := a.getOrCreateSongbook(db, Acronym_KK, "Katolický kancionál")
	if err != nil {
		return fmt.Errorf("failed to get or create KK songbook: %w", err)
	}

	totalFiles := len(xmlFiles)
	for i, xmlFile := range xmlFiles {
		if i%10 == 0 || i == totalFiles-1 {
			percent := 50 + int((float64(i+1)/float64(totalFiles))*50) // 50-100%
			message := fmt.Sprintf("Plním KK databázi... (%d/%d)", i+1, totalFiles)
			a.updateProgress(message, percent)
		}

		if err := a.processKKSongFile(db, xmlFile, songbookAcronym); err != nil {
			slog.Error("Failed to process KK song file", "file", xmlFile.Name(), "error", err)
		}
	}
	return nil
}

// getXmlFilesFromDir reads and validates XML files from a directory
func (a *App) getXmlFilesFromDir(dir string) ([]os.DirEntry, error) {
	xmlFiles, err := os.ReadDir(dir)
	if err != nil {
		slog.Error("Error reading XML files directory", "error", err, "dir", dir)
		return nil, err
	}
	if len(xmlFiles) == 0 {
		slog.Warn("No XML files in directory", "directory", dir)
		return nil, fmt.Errorf("no XML files found")
	}

	// Filter out directories
	var files []os.DirEntry
	for _, f := range xmlFiles {
		if !f.IsDir() {
			files = append(files, f)
		}
	}

	if a.testRun && len(files) > 25 {
		files = files[:25]
	}

	return files, nil
}

// getXmlFiles is kept for backward compatibility - delegates to root songbook dir
func (a *App) getXmlFiles() ([]os.DirEntry, error) {
	return a.getXmlFilesFromDir(a.songBookDir)
}

// processEZSongFile parses an EZ XML file and inserts song data
func (a *App) processEZSongFile(db *sql.DB, xmlFile os.DirEntry, songbookAcronym string) error {
	xmlFilePath := filepath.Join(a.songBookDir, "EZ", xmlFile.Name())

	song, err := parseXmlSong(xmlFilePath)
	if err != nil {
		return fmt.Errorf("failed to parse EZ XML: %w", err)
	}

	songID, err := a.insertSong(db, song, songbookAcronym)
	if err != nil {
		return fmt.Errorf("failed to insert EZ song: %w", err)
	}

	if err := a.insertAuthors(db, songID, song.Authors, xmlFile.Name()); err != nil {
		return fmt.Errorf("failed to insert EZ authors: %w", err)
	}

	if err := a.insertVerses(db, songID, song.Lyrics.Verses, xmlFile.Name()); err != nil {
		return fmt.Errorf("failed to insert EZ verses: %w", err)
	}

	slog.Debug("EZ data inserted", "entry", song.Songbook.Entry, "title", song.Title, "file", xmlFile.Name())
	return nil
}

// processKKSongFile parses a KK XML file and inserts song data
func (a *App) processKKSongFile(db *sql.DB, xmlFile os.DirEntry, songbookAcronym string) error {
	xmlFilePath := filepath.Join(a.songBookDir, "KK", "Kancional", xmlFile.Name())

	song, err := parseXmlSongKK(xmlFilePath)
	if err != nil {
		return fmt.Errorf("failed to parse KK XML: %w", err)
	}

	songID, err := a.insertSongKK(db, song, songbookAcronym)
	if err != nil {
		return fmt.Errorf("failed to insert KK song: %w", err)
	}

	if err := a.insertVersesKK(db, songID, song.Lyrics, xmlFile.Name()); err != nil {
		return fmt.Errorf("failed to insert KK verses: %w", err)
	}

	slog.Debug("KK data inserted", "entry", song.HymnNumber, "title", song.Title, "file", xmlFile.Name())
	return nil
}

// parseHymnNumber extracts the numeric part from hymn numbers that may contain letters or special characters
// Examples: "511A" -> 511, "067a" -> 67, "2.4" -> 2, "067b" -> 67
func parseHymnNumber(hymnStr string) int {
	hymnStr = strings.TrimSpace(hymnStr)

	// Extract leading digits only
	var numStr strings.Builder
	for _, ch := range hymnStr {
		if ch >= '0' && ch <= '9' {
			numStr.WriteRune(ch)
		} else {
			// Stop at first non-digit
			break
		}
	}

	if numStr.Len() == 0 {
		return 0
	}

	num, err := strconv.Atoi(numStr.String())
	if err != nil {
		return 0
	}
	return num
}

// insertSongKK inserts a KK song record and returns the song ID
func (a *App) insertSongKK(db *sql.DB, song *SongKK, songbookAcronym string) (int64, error) {
	// Remove song number prefix from title (e.g., "065 Litanie..." -> "Litanie...")
	title := song.Title
	if idx := strings.Index(title, " "); idx > 0 {
		// Check if prefix starts with numeric (handles "511A Title", "067b Title", etc.)
		prefix := title[:idx]
		if len(prefix) > 0 && prefix[0] >= '0' && prefix[0] <= '9' {
			// Strip the number prefix (includes letters like "511A")
			title = strings.TrimSpace(title[idx+1:])
		}
	}

	title_d := removeDiacritics(title)

	// Convert hymn number to integer, handling letters and special characters
	entryNum := parseHymnNumber(song.HymnNumber)

	// V3+ - insert with original entry_text (supports "511A", "067b", etc.)
	result, err := db.Exec(`INSERT INTO songs (songbook_acronym, title, title_d, verse_order, entry, entry_text) VALUES (?, ?, ?, ?, ?, ?)`,
		songbookAcronym, title, title_d, "", entryNum, song.HymnNumber)

	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// insertVersesKK parses and inserts KK verses from plain text with [V1], [V2] markers
func (a *App) insertVersesKK(db *sql.DB, songID int64, lyrics string, filename string) error {
	// Split lyrics by verse markers like [V1], [V2], etc.
	lines := strings.Split(lyrics, "\n")
	var currentVerseName string
	var currentVerseLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line is a verse marker
		if strings.HasPrefix(line, "[V") && strings.Contains(line, "]") {
			// Save previous verse if exists
			if currentVerseName != "" && len(currentVerseLines) > 0 {
				verseText := strings.Join(currentVerseLines, "\n")
				lines_d := removeDiacritics(verseText)
				_, err := db.Exec(`INSERT INTO verses (song_id, name, lines, lines_d) VALUES (?, ?, ?, ?)`,
					songID, currentVerseName, verseText, lines_d)
				if err != nil {
					slog.Error("Error inserting KK verse", "file", filename, "verse", currentVerseName, "error", err)
				}
			}

			// Start new verse
			endIdx := strings.Index(line, "]")
			currentVerseName = strings.ToLower(line[1:endIdx])
			currentVerseLines = []string{}
		} else if currentVerseName != "" {
			// Add line to current verse
			currentVerseLines = append(currentVerseLines, line)
		}
	}

	// Insert last verse
	if currentVerseName != "" && len(currentVerseLines) > 0 {
		verseText := strings.Join(currentVerseLines, "\n")
		lines_d := removeDiacritics(verseText)
		_, err := db.Exec(`INSERT INTO verses (song_id, name, lines, lines_d) VALUES (?, ?, ?, ?)`,
			songID, currentVerseName, verseText, lines_d)
		if err != nil {
			slog.Error("Error inserting KK verse (last)", "file", filename, "verse", currentVerseName, "error", err)
		}
	}

	return nil
}

// insertSong inserts a song record and returns the song ID
func (a *App) insertSong(db *sql.DB, song *Song, songbookAcronym string) (int64, error) {
	title_d := removeDiacritics(song.Title)

	// Check if songbook_acronym column exists (V2+ schema)

	// Convert entry string to integer for storage in entry column
	entryNum, _ := strconv.Atoi(song.Songbook.Entry)

	result, err := db.Exec(`INSERT INTO songs (songbook_acronym, title, title_d, verse_order, entry, entry_text) VALUES (?, ?, ?, ?, ?, ?)`,
		songbookAcronym, song.Title, title_d, song.VerseOrder, entryNum, song.Songbook.Entry)

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
	row := db.QueryRow(`SELECT songbook_acronym FROM songbooks WHERE songbook_acronym = ?`, acronym)
	var existingAcronym string
	err := row.Scan(&existingAcronym)
	if err == nil {
		return existingAcronym, nil // Found existing songbook
	}
	if err != sql.ErrNoRows {
		return "", err // Other error
	}

	// Create new songbook
	_, err = db.Exec(`INSERT INTO songbooks (songbook_acronym, name) VALUES (?, ?)`, acronym, name)
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
    COALESCE(kytara_file, '') AS kytara_file,
    COALESCE(s.songbook_acronym, '') AS songbook_acronym
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
				title, allVerses, authorMusic, authorLyric, kytaraFile, songbookAcronym string
				id, entry                                                               int
			)
			err := rows.Scan(&id, &entry, &title, &allVerses, &authorMusic, &authorLyric, &kytaraFile, &songbookAcronym)
			if err != nil {
				slog.Error(fmt.Sprintf("Error scanning row: %s", err))
				return err
			}

			result = append(result, dtoSong{Id: id, Entry: entry, Title: title, Verses: allVerses, AuthorMusic: authorMusic, AuthorLyric: authorLyric, KytaraFile: kytaraFile, SongbookAcronym: songbookAcronym})
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
