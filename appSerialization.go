package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/antchfx/htmlquery"
	"gopkg.in/yaml.v2"
)

func (a *App) serializeToYaml(book SongFilesSources) {

	// Serialize to YAML
	yamlData, err := yaml.Marshal(book)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	yamlPath := filepath.Join(a.appDir, "data.yaml")

	// Write YAML data to a file
	err = os.WriteFile(yamlPath, yamlData, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	fmt.Println("YAML Serialization to file completed.")
}

func (a *App) parseHtml(fileName string) []FileItem {
	// Open the HTML file
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// Parse the HTML document
	doc, err := htmlquery.Parse(file)
	if err != nil {
		log.Fatal(err)
	}

	// Use XPath to select a list of nodes /html/body/section[1]/div/div/div/h4[xxxxxx]/a
	nodes, err := htmlquery.QueryAll(doc, "/html/body/section[1]/div/div/div/h4/a")
	if err != nil {
		log.Fatal(err)
	}

	// Create a slice to store the list of items
	var dataList []FileItem

	// Iterate over the nodes and fill the slice
	for _, node := range nodes {
		// Extract values from each <a> element
		item := FileItem{
			Href:     htmlquery.SelectAttr(node, "href"),
			Title:    htmlquery.SelectAttr(node, "title"),
			Download: htmlquery.SelectAttr(node, "download"),
		}

		dataList = append(dataList, item)
	}

	return dataList
}
