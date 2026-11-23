package app

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
)

// withDB opens a database connection, executes the provided function, and ensures cleanup
func (a *App) withDB(fn func(*sql.DB) error) error {
	db, err := sql.Open("sqlite3", a.dbFilePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Error opening database: %s", err))
		return err
	}
	defer db.Close()
	return fn(db)
}

// updateProgress updates the progress message and percentage, then saves status
func (a *App) updateProgress(message string, percent int) {
	a.status.ProgressMessage = message
	a.status.ProgressKey = ""
	a.status.ProgressCurrent = 0
	a.status.ProgressTotal = 0
	a.status.ProgressPercent = percent
	a.saveStatus()
}

// updateProgressKey updates progress using a localization key and optional counts.
func (a *App) updateProgressKey(key string, percent int, current int, total int) {
	a.status.ProgressMessage = ""
	a.status.ProgressKey = key
	a.status.ProgressCurrent = current
	a.status.ProgressTotal = total
	a.status.ProgressPercent = percent
	a.saveStatus()
}

// clearProgress resets all progress indicators and saves status
func (a *App) clearProgress() {
	a.status.IsProgress = false
	a.status.ProgressMessage = ""
	a.status.ProgressKey = ""
	a.status.ProgressCurrent = 0
	a.status.ProgressTotal = 0
	a.status.ProgressPercent = 0
	a.saveStatus()
}

// startProgress sets the IsProgress flag and initializes progress tracking
func (a *App) startProgress(message string) {
	a.status.IsProgress = true
	a.updateProgress(message, 0)
}

// startProgressKey sets IsProgress and initializes localized progress tracking.
func (a *App) startProgressKey(key string) {
	a.status.IsProgress = true
	a.updateProgressKey(key, 0, 0, 0)
}
