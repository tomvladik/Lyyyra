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

	"pdf-crop/pkg/crop"

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

// GetCombinedPdfWithOptions merges PDF files and optionally crops pages using pixel-perfect detection.
// When crop is true, uses pdf-crop library with MuPDF rendering for accurate bounds.
// marginRatio is ignored (kept for API compatibility) - uses library defaults.
func (a *App) GetCombinedPdfWithOptions(filenames []string, cropEnabled bool, marginRatio float64) (string, error) {
	return a.combinePdfs(filenames, combinePdfOptions{crop: cropEnabled, marginRatio: marginRatio})
}

func (a *App) combinePdfs(filenames []string, opts combinePdfOptions) (string, error) {
	if len(filenames) == 0 {
		return "", fmt.Errorf("no filenames provided")
	}

	// If cropping requested, use pdf-crop library for all files
	if opts.crop {
		return a.combinePdfsWithCrop(filenames)
	}

	// Otherwise, simple merge without cropping
	c := creator.New()
	pagesAdded := 0

	for _, name := range filenames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		count, err := a.addPdfToCreator(c, name)
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

// addPdfToCreator opens a PDF file and adds all pages to the creator (without cropping)
func (a *App) addPdfToCreator(c *creator.Creator, filename string) (int, error) {
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

		if err := c.AddPage(page); err != nil {
			return 0, fmt.Errorf("unable to add page %d from %s: %w", pageNum, filename, err)
		}
	}

	return numPages, nil
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

// combinePdfsWithCrop uses pdf-crop library to crop and merge PDFs with pixel-perfect detection
func (a *App) combinePdfsWithCrop(filenames []string) (string, error) {
	if len(filenames) == 0 {
		return "", fmt.Errorf("no filenames provided")
	}

	// Create temp file for merged output
	tempDir := os.TempDir()
	tempOutput := filepath.Join(tempDir, fmt.Sprintf("lyyyra_cropped_%d.pdf", os.Getpid()))
	defer os.Remove(tempOutput)

	// First merge all input PDFs into temp file using pdfcpu
	inputPaths := make([]string, 0, len(filenames))
	for _, name := range filenames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		inputPaths = append(inputPaths, filepath.Join(a.pdfDir, name))
	}

	if len(inputPaths) == 0 {
		return "", fmt.Errorf("no valid input files")
	}

	// Merge first (simple concatenation)
	tempMerged := filepath.Join(tempDir, fmt.Sprintf("lyyyra_merged_%d.pdf", os.Getpid()))
	defer os.Remove(tempMerged)

	c := creator.New()
	for _, inputPath := range inputPaths {
		file, err := os.Open(inputPath)
		if err != nil {
			return "", fmt.Errorf("unable to open %s: %w", inputPath, err)
		}

		reader, err := model.NewPdfReader(file)
		file.Close()
		if err != nil {
			return "", fmt.Errorf("unable to read %s: %w", inputPath, err)
		}

		numPages, _ := reader.GetNumPages()
		for pageNum := 1; pageNum <= numPages; pageNum++ {
			page, err := reader.GetPage(pageNum)
			if err != nil {
				continue
			}
			if err := c.AddPage(page); err != nil {
				slog.Warn("Error adding page to creator", "pageNum", pageNum, "error", err)
				continue
			}
		}
	}

	// Write merged PDF
	if err := c.WriteToFile(tempMerged); err != nil {
		return "", fmt.Errorf("failed to write merged PDF: %w", err)
	}

	// Now crop using pdf-crop library with the same defaults as crop_all_pdf CLI
	opts := crop.DefaultOptions()
	opts.Threshold = 0.1
	opts.Space = 5
	opts.DPI = 128
	opts.CropFrom = "center"
	slog.Info("Cropping PDF with pixel-perfect detection", "input", tempMerged, "dpi", opts.DPI, "threshold", opts.Threshold, "space", opts.Space, "cropFrom", opts.CropFrom)

	_, err := crop.CropAllPagesToSingleFile(tempMerged, tempOutput, opts)
	if err != nil {
		// If cropping fails (e.g., MuPDF not available on Windows), fall back to uncropped
		slog.Warn("PDF cropping failed, returning uncropped PDF", "error", err)
		return a.readPdfAsDataURL(tempMerged)
	}

	// Validate cropped size vs original to avoid over-cropping (missing lyrics/credits)
	origW, origH, errOrig := firstPageSize(tempMerged)
	cropW, cropH, errCrop := firstPageSize(tempOutput)
	if errOrig == nil && errCrop == nil {
		ratioW := cropW / origW
		ratioH := cropH / origH
		if ratioW < 0.8 || ratioH < 0.8 {
			slog.Warn("Cropped page significantly smaller than original; returning uncropped PDF", "ratioW", ratioW, "ratioH", ratioH)
			return a.readPdfAsDataURL(tempMerged)
		}
	}

	// Read cropped PDF and encode to data URL
	return a.readPdfAsDataURL(tempOutput)
}

func firstPageSize(path string) (float64, float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	r, err := model.NewPdfReader(f)
	if err != nil {
		return 0, 0, err
	}

	page, err := r.GetPage(1)
	if err != nil {
		return 0, 0, err
	}

	box, err := page.GetMediaBox()
	if err != nil {
		return 0, 0, err
	}

	width := box.Urx - box.Llx
	height := box.Ury - box.Lly
	return width, height, nil
}

func (a *App) readPdfAsDataURL(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return "data:application/pdf;base64," + encoded, nil
}

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
		stmt, err := db.Prepare("UPDATE songs SET kytara_file = ? WHERE entry = ? AND songbook_acronym = 'EZ'")
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
