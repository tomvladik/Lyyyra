package app

import (
	"database/sql"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// reconcileStoredStatus inspects existing files on disk so the UI reflects the real state
// even after application restarts or manual data copies.
func (a *App) reconcileStoredStatus() {
	songsReady := a.hasDownloadedSongs()
	databaseReady := a.hasDatabaseContent()
	webResourcesReady := a.hasPdfSources()

	updated := false

	if songsReady != a.status.SongsReady || databaseReady != a.status.DatabaseReady || webResourcesReady != a.status.WebResourcesReady {
		slog.Info("Reconciling stored status flags", "songsReady", songsReady, "databaseReady", databaseReady, "webResourcesReady", webResourcesReady)
		a.status.SongsReady = songsReady
		a.status.DatabaseReady = databaseReady
		a.status.WebResourcesReady = webResourcesReady
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

	if a.status.SongsReady && !a.status.WebResourcesReady {
		a.ensureBackgroundSupplementalDownload()
	}
}

func (a *App) hasDownloadedSongs() bool {
	searchDir := filepath.Join(a.songBookDir, "EZ")
	entries, err := os.ReadDir(searchDir)
	if err != nil {
		// Backward compatibility with legacy layout where XML files were at songBookDir root.
		searchDir = a.songBookDir
		entries, err = os.ReadDir(searchDir)
		if err != nil {
			slog.Warn("Failed to inspect songBookDir", "error", err)
			return false
		}
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
		slog.Warn("Unexpected number of XML files", "dir", searchDir, "found", count, "expected", ExpectedSongCount)
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

func (a *App) hasPdfSources() bool {
	if len(SupplementalPDFs) == 0 {
		return true
	}

	if a.pdfDir == "" {
		return false
	}

	for _, pdf := range SupplementalPDFs {
		fileName := pdf.FileName
		if fileName == "" {
			fileName = path.Base(pdf.URL)
		}
		targetPath := filepath.Join(a.pdfDir, fileName)
		if _, err := os.Stat(targetPath); err != nil {
			if os.IsNotExist(err) {
				slog.Warn("Missing supplemental PDF", "file", targetPath)
			} else {
				slog.Warn("Failed to inspect supplemental PDF", "file", targetPath, "error", err)
			}
			return false
		}
	}

	return true
}
