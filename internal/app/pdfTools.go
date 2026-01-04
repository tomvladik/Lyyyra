package app

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/oliverpool/unipdf/v3/creator"
	"github.com/oliverpool/unipdf/v3/extractor"
	"github.com/oliverpool/unipdf/v3/model"
)

// extractSongNumberFromPage tries to identify the song number from page content
func extractSongNumberFromPage(text string) (string, error) {
	lines := strings.Split(text, "\n")
	re := regexp.MustCompile(`^(\d{1,4})$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if matches := re.FindStringSubmatch(line); matches != nil {
			return matches[1], nil
		}
	}
	return "", fmt.Errorf("could not extract song number from page")
}

// splitPdfByPages splits a PDF into individual page files with naming like prefix_XXX.pdf
func splitPdfByPages(inputPath string, outputDir string, filePrefix string, skipFirstPages int, skipLastPages int) (map[int]string, error) {
	// Open the PDF file
	pdfFile, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("error opening PDF file: %w", err)
	}
	defer pdfFile.Close()

	// Read the PDF
	pdfReader, err := model.NewPdfReader(pdfFile)
	if err != nil {
		return nil, fmt.Errorf("error reading PDF: %w", err)
	}

	// Get total pages
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, fmt.Errorf("error getting page count: %w", err)
	}

	slog.Info(fmt.Sprintf("Processing PDF with %d pages (skip first: %d, skip last: %d)", numPages, skipFirstPages, skipLastPages))

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("error creating output directory: %w", err)
	}

	// Map of song number (entry) to filename
	songFiles := make(map[int]string)

	// Process pages (skip first and last as specified)
	for pageNum := skipFirstPages + 1; pageNum <= numPages-skipLastPages; pageNum++ {
		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			slog.Warn(fmt.Sprintf("Error getting page %d: %v", pageNum, err))
			continue
		}

		// Extract text to identify song number
		extraction, err := extractor.New(page)
		if err != nil {
			slog.Warn(fmt.Sprintf("Error creating extractor for page %d: %v", pageNum, err))
			continue
		}
		text, err := extraction.ExtractText()
		if err != nil {
			slog.Warn(fmt.Sprintf("Error extracting text from page %d: %v", pageNum, err))
			continue
		}

		// Try to extract song number
		songNumberStr, err := extractSongNumberFromPage(text)
		if err != nil {
			slog.Debug(fmt.Sprintf("Could not identify song number on page %d, skipping", pageNum))
			continue
		}

		// Parse song number as integer
		var songNumber int
		if _, err := fmt.Sscanf(songNumberStr, "%d", &songNumber); err != nil {
			slog.Warn(fmt.Sprintf("Could not parse song number '%s' on page %d", songNumberStr, pageNum))
			continue
		}

		// Create output filename with zero-padded 3-digit number
		outputFileName := fmt.Sprintf("%s_%03d.pdf", filePrefix, songNumber)
		outputPath := filepath.Join(outputDir, outputFileName)

		// Create a new PDF with just this page
		c := creator.New()
		if err := c.AddPage(page); err != nil {
			slog.Error(fmt.Sprintf("Error adding page %d to creator: %v", pageNum, err))
			continue
		}

		// Write the PDF file
		if err := c.WriteToFile(outputPath); err != nil {
			slog.Error(fmt.Sprintf("Error writing PDF for song %d: %v", songNumber, err))
			continue
		}

		slog.Info(fmt.Sprintf("Created PDF for song %d: %s", songNumber, outputFileName))
		songFiles[songNumber] = outputFileName
	}

	slog.Info(fmt.Sprintf("Successfully processed %d song pages into %s", len(songFiles), outputDir))
	return songFiles, nil
}

// ProcessKytaraPDF splits the kytara.pdf into individual song files and updates the database
func (a *App) ProcessKytaraPDF() error {
	if a.testRun {
		return nil
	}

	inputPath := filepath.Join(a.pdfDir, "kytara.pdf")
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("kytara.pdf not found at %s", inputPath)
	}

	outputDir := a.pdfDir
	songFiles, err := splitPdfByPages(inputPath, outputDir, "kytara", 10, 0)
	if err != nil {
		return err
	}

	// Update database with filenames
	return a.updateSongFilenames(songFiles)
}

// updateSongFilenames updates the database with the song PDF filenames
func (a *App) updateSongFilenames(songFiles map[int]string) error {
	return a.withDB(func(db *sql.DB) error {
		stmt, err := db.Prepare("UPDATE songs SET kytara_file = ? WHERE entry = ?")
		if err != nil {
			return err
		}
		defer stmt.Close()

		for entry, filename := range songFiles {
			if _, err := stmt.Exec(filename, entry); err != nil {
				slog.Warn(fmt.Sprintf("Error updating kytara_file for song %d: %v", entry, err))
			}
		}
		return nil
	})
}
