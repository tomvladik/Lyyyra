package main

import (
	"fmt"
	"log"
	"os"

	"github.com/oliverpool/unipdf/v3/extractor"
	"github.com/oliverpool/unipdf/v3/model"
)

func parsePdf(inputFilePath string) {

	// Open the PDF file.
	pdfFile, err := os.Open(inputFilePath)
	if err != nil {
		fmt.Println("Error opening PDF file:", err)
		return
	}
	defer pdfFile.Close()

	// Read the PDF file.
	pdfReader, err := model.NewPdfReader(pdfFile)
	if err != nil {
		fmt.Println("Error reading PDF file:", err)
		return
	}

	// Get the total number of pages in the PDF.
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		fmt.Println("Error getting number of pages:", err)
		return
	}
	// Iterate through each page and extract text content.
	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			log.Fatal(err)
		}
		// Create an extractor for the PDF content.
		extraction, err := extractor.New(page)
		text, err := extraction.ExtractText()
		if err != nil {
			fmt.Printf("Error extracting text from page %d: %v\n", pageNum, err)
			continue
		}

		fmt.Printf("Text content from page %d:\n%s\n", pageNum, text)
	}
}
