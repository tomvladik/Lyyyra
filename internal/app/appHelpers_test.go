package app

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestWithDB(t *testing.T) {
	tests := []struct {
		name      string
		setupDB   bool
		wantError bool
	}{
		{
			name:      "successful database operation",
			setupDB:   true,
			wantError: false,
		},
		{
			name:      "invalid database path",
			setupDB:   false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var app *App
			if tt.setupDB {
				app = setupTestDB(t)
				defer teardownTestDB(app)
			} else {
				app = &App{dbFilePath: "/invalid/path/to/db.sqlite"}
			}

			err := app.withDB(func(db *sql.DB) error {
				// Simple query to test connection
				_, err := db.Exec("SELECT 1")
				return err
			})

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestUpdateProgress(t *testing.T) {
	// Create temporary file for status
	tmpDir, err := os.MkdirTemp("", "test_app_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	app := &App{
		appDir: tmpDir,
		status: AppStatus{},
	}

	tests := []struct {
		name            string
		message         string
		percent         int
		expectedMessage string
		expectedPercent int
	}{
		{
			name:            "update progress to 50%",
			message:         "Processing...",
			percent:         50,
			expectedMessage: "Processing...",
			expectedPercent: 50,
		},
		{
			name:            "update progress to 100%",
			message:         "Complete",
			percent:         100,
			expectedMessage: "Complete",
			expectedPercent: 100,
		},
		{
			name:            "update progress to 0%",
			message:         "Starting...",
			percent:         0,
			expectedMessage: "Starting...",
			expectedPercent: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app.updateProgress(tt.message, tt.percent)

			if app.status.ProgressMessage != tt.expectedMessage {
				t.Errorf("expected message %q, got %q", tt.expectedMessage, app.status.ProgressMessage)
			}
			if app.status.ProgressPercent != tt.expectedPercent {
				t.Errorf("expected percent %d, got %d", tt.expectedPercent, app.status.ProgressPercent)
			}
		})
	}
}

func TestClearProgress(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_app_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	app := &App{
		appDir: tmpDir,
		status: AppStatus{
			IsProgress:      true,
			ProgressMessage: "In progress",
			ProgressPercent: 75,
		},
	}

	app.clearProgress()

	if app.status.IsProgress {
		t.Error("expected IsProgress to be false")
	}
	if app.status.ProgressMessage != "" {
		t.Errorf("expected empty message, got %q", app.status.ProgressMessage)
	}
	if app.status.ProgressPercent != 0 {
		t.Errorf("expected percent 0, got %d", app.status.ProgressPercent)
	}
}

func TestStartProgress(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_app_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	app := &App{
		appDir: tmpDir,
		status: AppStatus{},
	}

	message := "Starting operation"
	app.startProgress(message)

	if !app.status.IsProgress {
		t.Error("expected IsProgress to be true")
	}
	if app.status.ProgressMessage != message {
		t.Errorf("expected message %q, got %q", message, app.status.ProgressMessage)
	}
	if app.status.ProgressPercent != 0 {
		t.Errorf("expected percent 0, got %d", app.status.ProgressPercent)
	}
}
