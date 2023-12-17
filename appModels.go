package main

import "encoding/xml"

type SongFilesSources struct {
	Domain    string
	Url       string
	UrlScheme string
	Items     []FileItem
}

type FileItem struct {
	Href          string
	Title         string
	FileSize      string
	Download      string
	LocalFileName string
}

// Song represents the simplified structure of the XML document
type Song struct {
	XMLName    xml.Name `xml:"song"`
	Version    string   `xml:"version,attr"`
	Title      string   `xml:"properties>titles>title"`
	Songbook   Songbook `xml:"properties>songbooks>songbook"`
	VerseOrder string   `xml:"properties>verseOrder"`
	Authors    []Author `xml:"properties>authors>author"`
	Lyrics     struct {
		Verses []struct {
			Name  string `xml:"name,attr"`
			Lines string `xml:"lines"`
		} `xml:"verse"`
	} `xml:"lyrics"`
}

type Songbook struct {
	Name  string `xml:"name,attr"`
	Entry string `xml:"entry,attr"`
}

type Author struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}
