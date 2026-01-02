package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
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

// extractSongNumbersFromPage returns all plausible song numbers found on a page.
// It is more permissive than extractSongNumberFromPage because "noty" pages can
// contain multiple songs on the same page. Only matches explicit song number markers
// (like "1. Title" or "1 Title") to avoid capturing page numbers or other isolated digits.
func extractSongNumbersFromPage(text string) []int {
	lines := strings.Split(text, "\n")
	seen := make(map[int]bool)
	results := make([]int, 0)

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		// Match song numbers at line start with explicit markers: "1. Title" or "1 Title"
		// This avoids capturing isolated page numbers or headers.
		if re := regexp.MustCompile(`^(\d{1,4})(?:[\.)\s])`); re.MatchString(line) {
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.Atoi(matches[1]); err == nil {
					if val >= 1 && (ExpectedSongCount == 0 || val <= ExpectedSongCount) {
						if !seen[val] {
							seen[val] = true
							results = append(results, val)
						}
					}
				}
			}
		}
	}

	return results
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

// splitPdfBySongBoundaries tries to map song numbers to page ranges. If multiple
// songs live on the same page, each song will still get its own output PDF
// containing that shared page. Song ranges end just before the next detected
// song start (or the document end for the last one).
// skipFirstPages and skipLastPages allow ignoring offset (TOC, intro) and appendix pages.
func splitPdfBySongBoundaries(inputPath string, outputDir string, filePrefix string, skipFirstPages int, skipLastPages int) (map[int]string, error) {
	// Open the PDF file
	pdfFile, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("error opening PDF file: %w", err)
	}
	defer pdfFile.Close()

	reader, err := model.NewPdfReader(pdfFile)
	if err != nil {
		return nil, fmt.Errorf("error reading PDF: %w", err)
	}

	numPages, err := reader.GetNumPages()
	if err != nil {
		return nil, fmt.Errorf("error getting page count: %w", err)
	}

	slog.Info(fmt.Sprintf("Processing PDF with %d pages (skip first: %d, skip last: %d)", numPages, skipFirstPages, skipLastPages))

	pageStarts := make(map[int]int)

	for pageNum := skipFirstPages + 1; pageNum <= numPages-skipLastPages; pageNum++ {
		page, err := reader.GetPage(pageNum)
		if err != nil {
			slog.Warn(fmt.Sprintf("Error getting page %d: %v", pageNum, err))
			continue
		}

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

		numbers := extractSongNumbersFromPage(text)
		for _, entry := range numbers {
			if _, exists := pageStarts[entry]; !exists {
				pageStarts[entry] = pageNum
			}
		}
	}

	if len(pageStarts) == 0 {
		return nil, fmt.Errorf("no song numbers detected in %s", inputPath)
	}

	entries := make([]int, 0, len(pageStarts))
	for entry := range pageStarts {
		entries = append(entries, entry)
	}
	sort.Ints(entries)

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("error creating output directory: %w", err)
	}

	songFiles := make(map[int]string)

	for idx, entry := range entries {
		startPage := pageStarts[entry]
		endPage := numPages
		if idx+1 < len(entries) {
			nextStart := pageStarts[entries[idx+1]]
			if nextStart > startPage {
				endPage = nextStart - 1
			} else {
				endPage = startPage
			}
		}

		c := creator.New()
		pagesAdded := 0

		for pageNum := startPage; pageNum <= endPage; pageNum++ {
			page, err := reader.GetPage(pageNum)
			if err != nil {
				slog.Warn(fmt.Sprintf("Error reading page %d for song %d: %v", pageNum, entry, err))
				continue
			}
			if err := c.AddPage(page); err != nil {
				slog.Warn(fmt.Sprintf("Error adding page %d for song %d: %v", pageNum, entry, err))
				continue
			}
			pagesAdded++
		}

		if pagesAdded == 0 {
			slog.Warn(fmt.Sprintf("No pages added for song %d", entry))
			continue
		}

		filename := fmt.Sprintf("%s_%03d.pdf", filePrefix, entry)
		outputPath := filepath.Join(outputDir, filename)
		if err := c.WriteToFile(outputPath); err != nil {
			slog.Warn(fmt.Sprintf("Error writing PDF for song %d: %v", entry, err))
			continue
		}

		songFiles[entry] = filename
	}

	slog.Info(fmt.Sprintf("Successfully processed %d songs into %s", len(songFiles), outputDir))
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
	return a.updateSongFilenames(songFiles, "kytara_file")
}

// ProcessNotesPDF splits noty.pdf (standard/non-guitar notes) into song files and updates the database.
func (a *App) ProcessNotesPDF() error {
	if a.testRun {
		return nil
	}

	inputPath := filepath.Join(a.pdfDir, "noty.pdf")
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("noty.pdf not found at %s", inputPath)
	}

	outputDir := a.pdfDir
	// Skip first 26 pages (offset/intro) and last 1 page (appendix)
	// Adjust these values if needed based on actual PDF structure
	songFiles, err := splitPdfBySongBoundaries(inputPath, outputDir, "noty", 26, 1)
	if err != nil {
		return err
	}

	if len(songFiles) == 0 {
		return fmt.Errorf("no song files generated from %s", inputPath)
	}

	return a.updateSongFilenames(songFiles, "notes_file")
}

// updateSongFilenames updates the database with the song PDF filenames for the requested column.
func (a *App) updateSongFilenames(songFiles map[int]string, column string) error {
	column = strings.TrimSpace(strings.ToLower(column))
	switch column {
	case "kytara_file", "notes_file":
	default:
		return fmt.Errorf("unsupported column for filename update: %s", column)
	}

	return a.withDB(func(db *sql.DB) error {
		stmt, err := db.Prepare(fmt.Sprintf("UPDATE songs SET %s = ? WHERE entry = ?", column))
		if err != nil {
			return err
		}
		defer stmt.Close()

		for entry, filename := range songFiles {
			if _, err := stmt.Exec(filename, entry); err != nil {
				slog.Warn(fmt.Sprintf("Error updating %s for song %d: %v", column, entry, err))
			}
		}
		return nil
	})
}
