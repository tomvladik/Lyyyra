package app

import (
	"bytes"
	"database/sql"
	"encoding/base64"
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

type combinePdfOptions struct {
	crop        bool
	marginRatio float64
}

// GetCombinedPdf merges the provided PDF files (in order) into a single PDF
// and returns it as a base64 data URL
func (a *App) GetCombinedPdf(filenames []string) (string, error) {
	return a.combinePdfs(filenames, combinePdfOptions{crop: false, marginRatio: 0})
}

// GetCombinedPdfWithOptions merges PDF files and optionally crops pages before returning as a data URL.
// marginRatio is a fraction (0-0.5) of page width/height trimmed from each side when crop is true.
func (a *App) GetCombinedPdfWithOptions(filenames []string, crop bool, marginRatio float64) (string, error) {
	return a.combinePdfs(filenames, combinePdfOptions{crop: crop, marginRatio: marginRatio})
}

func (a *App) combinePdfs(filenames []string, opts combinePdfOptions) (string, error) {
	if len(filenames) == 0 {
		return "", fmt.Errorf("no filenames provided")
	}

	if opts.marginRatio <= 0 {
		opts.marginRatio = 0.02
	}
	if opts.marginRatio >= 0.5 {
		opts.marginRatio = 0.49
	}

	c := creator.New()
	pagesAdded := 0

	for _, name := range filenames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		count, err := a.addPdfToCreator(c, name, opts)
		if err != nil {
			return "", err
		}
		pagesAdded += count
	}

	if pagesAdded == 0 {
		return "", fmt.Errorf("no pages added to compilation")
	}

	return a.encodeCreatorToPdf(c)
}

// addPdfToCreator opens a PDF file, optionally crops its pages, and adds them to the creator
func (a *App) addPdfToCreator(c *creator.Creator, filename string, opts combinePdfOptions) (int, error) {
	filePath := filepath.Join(a.pdfDir, filename)
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("unable to open %s: %w", filename, err)
	}
	defer file.Close()

	reader, err := model.NewPdfReader(file)
	if err != nil {
		return 0, fmt.Errorf("unable to read %s: %w", filename, err)
	}

	numPages, err := reader.GetNumPages()
	if err != nil {
		return 0, fmt.Errorf("unable to get page count for %s: %w", filename, err)
	}

	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page, err := reader.GetPage(pageNum)
		if err != nil {
			return 0, fmt.Errorf("unable to read page %d from %s: %w", pageNum, filename, err)
		}

		if opts.crop {
			if err := cropPageInPlace(page, opts.marginRatio); err != nil {
				slog.Warn("skipping crop for page", "page", pageNum, "file", filename, "err", err)
			}
		}

		if err := c.AddPage(page); err != nil {
			return 0, fmt.Errorf("unable to add page %d from %s: %w", pageNum, filename, err)
		}
	}

	return numPages, nil
}

// cropPageInPlace detects content bounds and sets a tight CropBox.
// marginRatio adds a small padding around detected content (0.01 = 1%).
func cropPageInPlace(page *model.PdfPage, marginRatio float64) error {
	mediaBox := page.MediaBox
	if mediaBox == nil {
		return fmt.Errorf("page missing MediaBox")
	}

	// Detect content bounds by analyzing text presence.
	contentBounds, err := detectPageContentBounds(page)
	if err != nil {
		// Fallback: apply conservative uniform margin if detection fails.
		slog.Debug("content detection failed, falling back to uniform margin", "err", err)
		return fallbackUniformCrop(page, 0.05)
	}

	// Apply small padding around detected content.
	pageWidth := mediaBox.Urx - mediaBox.Llx
	pageHeight := mediaBox.Ury - mediaBox.Lly
	padX := marginRatio * pageWidth
	padY := marginRatio * pageHeight

	// Clamp padding.
	if padX > contentBounds.Llx-mediaBox.Llx {
		padX = (contentBounds.Llx - mediaBox.Llx) * 0.5
	}
	if padY > contentBounds.Lly-mediaBox.Lly {
		padY = (contentBounds.Lly - mediaBox.Lly) * 0.5
	}

	page.CropBox = &model.PdfRectangle{
		Llx: contentBounds.Llx - padX,
		Lly: contentBounds.Lly - padY,
		Urx: contentBounds.Urx + padX,
		Ury: contentBounds.Ury + padY,
	}

	return nil
}

// detectPageContentBounds estimates the bounding box of page content.
// Uses text extraction to determine content region and applies heuristic shrinking.
func detectPageContentBounds(page *model.PdfPage) (*model.PdfRectangle, error) {
	mediaBox := page.MediaBox
	if mediaBox == nil {
		return nil, fmt.Errorf("no MediaBox")
	}

	w := mediaBox.Urx - mediaBox.Llx
	h := mediaBox.Ury - mediaBox.Lly

	// Extract text to check if page has content.
	extraction, err := extractor.New(page)
	if err == nil && extraction != nil {
		text, _ := extraction.ExtractText()
		if len(strings.TrimSpace(text)) > 0 {
			// Page has text content.
			// Heuristic: typical sheet music/text has margins ~10-15% on sides, ~5-10% on bottom, ~5% on top.
			return &model.PdfRectangle{
				Llx: mediaBox.Llx + w*0.12, // ~12% from left
				Lly: mediaBox.Lly + h*0.08, // ~8% from bottom
				Urx: mediaBox.Urx - w*0.12, // ~12% from right
				Ury: mediaBox.Ury - h*0.08, // ~8% from top
			}, nil
		}
	}

	// Fallback: conservative estimate for blank pages.
	return &model.PdfRectangle{
		Llx: mediaBox.Llx + w*0.15,
		Lly: mediaBox.Lly + h*0.1,
		Urx: mediaBox.Urx - w*0.15,
		Ury: mediaBox.Ury - h*0.15,
	}, nil
}

// fallbackUniformCrop applies a conservative uniform margin when content detection fails.
func fallbackUniformCrop(page *model.PdfPage, marginRatio float64) error {
	box := page.MediaBox
	if box == nil {
		return fmt.Errorf("no MediaBox")
	}

	width := box.Urx - box.Llx
	height := box.Ury - box.Lly
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid dimensions")
	}

	dx := marginRatio * width
	dy := marginRatio * height

	page.CropBox = &model.PdfRectangle{
		Llx: box.Llx + dx,
		Lly: box.Lly + dy,
		Urx: box.Urx - dx,
		Ury: box.Ury - dy,
	}

	return nil
}

// encodeCreatorToPdf writes the creator content to a buffer and returns as base64 data URL
func (a *App) encodeCreatorToPdf(c *creator.Creator) (string, error) {
	buf := &bytes.Buffer{}
	if err := c.Write(buf); err != nil {
		return "", fmt.Errorf("unable to write compiled PDF: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:application/pdf;base64," + encoded, nil
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
