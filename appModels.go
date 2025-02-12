package main

import "time"

type AppStatus struct {
	WebResourcesReady bool
	SongsReady        bool
	DatabaseReady     bool
	LastSave          time.Time
	Sorting           SortingOption
}

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
	Version    string   `xml:"version,attr"`
	Title      string   `xml:"properties>titles>title"`
	Songbook   Songbook `xml:"properties>songbooks>songbook"`
	VerseOrder string   `xml:"properties>verseOrder"`
	Authors    []Author `xml:"properties>authors>author"`
	Lyrics     Lyrics   `xml:"lyrics"`
}

type Songbook struct {
	Name  string `xml:"name,attr"`
	Entry string `xml:"entry,attr"`
}

type Author struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type Lyrics struct {
	Verses []Verse `xml:"verse"`
}

type Verse struct {
	Name  string `xml:"name,attr"`
	Lines string `xml:",innerxml"`
}

type dtoSong struct {
	Id     int
	Entry  int
	Title  string
	Verses string
}

type SortingOption string

const (
	Entry       SortingOption = "entry"
	Title       SortingOption = "title"
	AuthorMusic SortingOption = "authorMusic"
	AuthorLyric SortingOption = "authorLyric"
)
