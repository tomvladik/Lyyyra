package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/oliverpool/unipdf/v3/extractor"
	"github.com/oliverpool/unipdf/v3/model"
)

func parsePdf(inputFilePath string) {

	// Open the PDF file.
	pdfFile, err := os.Open(inputFilePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Error opening PDF file: %s", err))
		return
	}
	defer pdfFile.Close()

	// Read the PDF file.
	pdfReader, err := model.NewPdfReader(pdfFile)
	if err != nil {
		slog.Error(fmt.Sprintf("Error reading PDF file: %s", err))
		return
	}

	// Get the total number of pages in the PDF.
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting number of pages: %s", err))
		return
	}
	// Iterate through each page and extract text content.
	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			slog.Error(err.Error())
		}
		// Create an extractor for the PDF content.
		extraction, err := extractor.New(page)
		text, err := extraction.ExtractText()
		if err != nil {
			slog.Error(fmt.Sprintf("Error extracting text from page %d: %v\n", pageNum, err))
			continue
		}

		slog.Debug(fmt.Sprintf("Text content from page %d:\n%s\n", pageNum, text))
	}
}
