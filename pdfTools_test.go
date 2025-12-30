package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestSplitPdfByPages(t *testing.T) {
	if os.Getenv("RUN_REMOTE_INTEGRATION_TESTS") == "" {
		t.Skip("skipping remote PDF test; set RUN_REMOTE_INTEGRATION_TESTS=1 to enable")
	}

	// Create temporary directories for test
	tempDir, err := os.MkdirTemp("", "pdf_split_test_")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Download kytara.pdf
	kytaraURL := "https://www.evangelickyzpevnik.cz.www.e-cirkev.cz/res/archive/001/000234.pdf"
	pdfPath := filepath.Join(tempDir, "kytara.pdf")

	t.Logf("Downloading kytara.pdf from %s...", kytaraURL)
	resp, err := http.Get(kytaraURL)
	if err != nil {
		t.Fatalf("Failed to download PDF: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to download PDF: HTTP %d", resp.StatusCode)
	}

	pdfFile, err := os.Create(pdfPath)
	if err != nil {
		t.Fatalf("Failed to create PDF file: %v", err)
	}
	defer pdfFile.Close()

	_, err = io.Copy(pdfFile, resp.Body)
	if err != nil {
		t.Fatalf("Failed to write PDF file: %v", err)
	}
	t.Logf("Downloaded PDF to %s", pdfPath)

	// Output directory for split PDFs
	outputDir := filepath.Join(tempDir, "output")

	// Split PDF with skip parameters
	// Note: Adjust skipFirstPages and skipLastPages based on actual PDF structure
	skipFirst := 1 // Skip cover/first page
	skipLast := 1  // Skip back cover/last page

	t.Logf("Splitting PDF (skip first %d, skip last %d pages)...", skipFirst, skipLast)
	songFiles, err := splitPdfByPages(pdfPath, outputDir, "kytara", skipFirst, skipLast)
	if err != nil {
		t.Fatalf("Failed to split PDF: %v", err)
	}

	t.Logf("Successfully split PDF into %d song files", len(songFiles))

	// Verify output files
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}

	if len(entries) == 0 {
		t.Error("No output files created")
	} else {
		t.Logf("Created %d PDF files", len(entries))
		for _, entry := range entries {
			t.Logf("  - %s", entry.Name())
		}
	}

	// Copy results to workspace for inspection
	workspaceOutputDir := filepath.Join("/workspaces/Lyyyra/build/pdf_split_results")
	if err := os.MkdirAll(workspaceOutputDir, os.ModePerm); err != nil {
		t.Fatalf("Failed to create results directory: %v", err)
	}

	t.Logf("Copying results to %s", workspaceOutputDir)
	for _, entry := range entries {
		srcPath := filepath.Join(outputDir, entry.Name())
		dstPath := filepath.Join(workspaceOutputDir, entry.Name())

		srcFile, err := os.Open(srcPath)
		if err != nil {
			t.Errorf("Failed to open source file %s: %v", srcPath, err)
			continue
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			t.Errorf("Failed to create destination file %s: %v", dstPath, err)
			continue
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			t.Errorf("Failed to copy file %s: %v", entry.Name(), err)
			continue
		}
		t.Logf("Copied: %s", entry.Name())
	}

	t.Logf("Results available at: %s", workspaceOutputDir)
}

func TestSplitPdfByPagesWithSkip(t *testing.T) {
	if os.Getenv("RUN_REMOTE_INTEGRATION_TESTS") == "" {
		t.Skip("skipping remote PDF test; set RUN_REMOTE_INTEGRATION_TESTS=1 to enable")
	}

	// This test allows you to experiment with different skip values
	// Run with: go test -v -run TestSplitPdfByPagesWithSkip -tags webkit2_41

	tempDir, err := os.MkdirTemp("", "pdf_split_skip_test_")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Download PDF
	kytaraURL := "https://www.evangelickyzpevnik.cz.www.e-cirkev.cz/res/archive/001/000234.pdf"
	pdfPath := filepath.Join(tempDir, "kytara.pdf")

	t.Logf("Downloading PDF...")
	resp, err := http.Get(kytaraURL)
	if err != nil {
		t.Fatalf("Failed to download PDF: %v", err)
	}
	defer resp.Body.Close()

	pdfFile, err := os.Create(pdfPath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer pdfFile.Close()

	io.Copy(pdfFile, resp.Body)

	// Try different skip configurations
	skipConfigs := []struct {
		name      string
		skipFirst int
		skipLast  int
		maxSongs  int // Just process first N songs
	}{
		{"minimal_skip", 1, 1, 5},
		{"more_skip_start", 3, 1, 5},
		{"more_skip_both", 3, 3, 5},
	}

	for _, config := range skipConfigs {
		t.Run(config.name, func(t *testing.T) {
			outputDir := filepath.Join(tempDir, fmt.Sprintf("split_%s", config.name))

			t.Logf("Testing: skipFirst=%d, skipLast=%d", config.skipFirst, config.skipLast)
			songFiles, err := splitPdfByPages(pdfPath, outputDir, "kytara", config.skipFirst, config.skipLast)
			if err != nil {
				t.Fatalf("Failed to split PDF: %v", err)
			}

			entries, _ := os.ReadDir(outputDir)
			t.Logf("Created %d songs, mapped %d entries", len(entries), len(songFiles))

			// Copy best result to workspace
			if config.name == "minimal_skip" {
				workspaceDir := filepath.Join("/workspaces/Lyyyra/build/pdf_split_results")
				os.MkdirAll(workspaceDir, os.ModePerm)
				for _, entry := range entries {
					src := filepath.Join(outputDir, entry.Name())
					dst := filepath.Join(workspaceDir, entry.Name())
					srcFile, _ := os.Open(src)
					defer srcFile.Close()
					dstFile, _ := os.Create(dst)
					defer dstFile.Close()
					io.Copy(dstFile, srcFile)
				}
			}
		})
	}
}
