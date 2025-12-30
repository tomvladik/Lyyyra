package main

import (
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// reconcileStoredStatus inspects existing files on disk so the UI reflects the real state
// even after application restarts or manual data copies.
func (a *App) reconcileStoredStatus() {
	songsReady := a.hasDownloadedSongs()
	databaseReady := a.hasDatabaseContent()

	updated := false

	if songsReady != a.status.SongsReady || databaseReady != a.status.DatabaseReady {
		slog.Info("Reconciling stored status flags", "songsReady", songsReady, "databaseReady", databaseReady)
		a.status.SongsReady = songsReady
		a.status.DatabaseReady = databaseReady
		updated = true
	}

	if !isValidSortingOption(a.status.Sorting) {
		slog.Info("Resetting invalid sorting option", "value", a.status.Sorting)
		a.status.Sorting = Entry
		updated = true
	}

	if updated {
		a.saveStatus()
	}
}

func (a *App) hasDownloadedSongs() bool {
	entries, err := os.ReadDir(a.songBookDir)
	if err != nil {
		slog.Warn("Failed to inspect songBookDir", "error", err)
		return false
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".xml") {
			count++
		}
	}
	if a.testRun {
		return count > 0
	}
	if count != ExpectedSongCount {
		slog.Warn("Unexpected number of XML files", "found", count, "expected", ExpectedSongCount)
		return false
	}
	return true
}

func (a *App) hasDatabaseContent() bool {
	if _, err := os.Stat(a.dbFilePath); err != nil {
		return false
	}

	var count int
	err := a.withDB(func(db *sql.DB) error {
		return db.QueryRow("SELECT COUNT(*) FROM songs").Scan(&count)
	})
	if err != nil {
		slog.Warn("Failed to verify database contents", "error", err)
		return false
	}
	if a.testRun {
		return count > 0
	}
	if count != ExpectedSongCount {
		slog.Warn("Unexpected number of songs in database", "found", count, "expected", ExpectedSongCount)
		return false
	}
	return true
}
