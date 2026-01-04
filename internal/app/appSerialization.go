package app

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/antchfx/htmlquery"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v2"
)

func (a *App) serializeToYaml(fileName string, data interface{}) {

	// Serialize to YAML
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	yamlPath := filepath.Join(a.appDir, fileName)

	// Write YAML data to a file
	err = os.WriteFile(yamlPath, yamlData, 0644)
	if err != nil {
		slog.Error(fmt.Sprintf("Error writing to file: %s", err))
		return
	}
	slog.Info(fmt.Sprintf("YAML Serialization to file completed. %v %s", data, yamlPath))
}

func (a *App) deserializeFromYaml(dataStruct interface{}, fileName string) error {
	yamlPath := filepath.Join(a.appDir, fileName)
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, dataStruct)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) parseHtml(fileName string) []FileItem {
	// Open the HTML file
	file, err := os.Open(fileName)
	if err != nil {
		slog.Error(err.Error())
	}
	defer file.Close()
	// Parse the HTML document
	doc, err := htmlquery.Parse(file)
	if err != nil {
		slog.Error(err.Error())
	}

	// Use XPath to select a list of nodes /html/body/section[1]/div/div/div/h4[xxxxxx]/a
	nodes, err := htmlquery.QueryAll(doc, "/html/body/section[1]/div/div/div/h4/a")
	if err != nil {
		slog.Error(err.Error())
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

func parseXmlSong(xmlFilePath string) (*Song, error) {
	xmlData, err := os.ReadFile(xmlFilePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Error reading XML file %s: %v\n", xmlFilePath, err))
		return nil, err
	}

	var song Song
	err = xml.Unmarshal(xmlData, &song)
	if err != nil {
		slog.Error(fmt.Sprintf("Error unmarshalling XML in file %s: %v\n", xmlFilePath, err))
		return nil, err
	}

	// Only collapse multiple spaces/tabs, but preserve newlines for verse formatting.
	reWhiteSpaces := regexp.MustCompile(`[ \t]{2,}`)
	reToRemove := regexp.MustCompile(`<lines>|</lines>`)
	for i, verse := range song.Lyrics.Verses {
		// Trim leading and trailing whitespace from each line
		trimmed := strings.TrimSpace(verse.Lines)

		// Replace <br /> and <br/> with actual newlines to preserve line breaks
		trimmed = strings.ReplaceAll(trimmed, "<br />", "\n")
		trimmed = strings.ReplaceAll(trimmed, "<br/>", "\n")

		// Remove the lines tags
		trimmed = reToRemove.ReplaceAllString(trimmed, "")

		// Replace multiple spaces/tabs with a single space (newlines are preserved)
		trimmed = reWhiteSpaces.ReplaceAllString(trimmed, " ")

		song.Lyrics.Verses[i].Lines = trimmed
	}
	return &song, nil
}

// removeDiacritics removes accents from a UTF-8 string
func removeDiacritics(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

// isMn checks if a rune is a non-spacing mark (diacritic)
func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}
