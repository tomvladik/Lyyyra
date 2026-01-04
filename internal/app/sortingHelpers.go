package app

import "strings"

func normalizeSortingOption(raw string) SortingOption {
	trimmed := strings.TrimSpace(raw)
	option := SortingOption(trimmed)
	switch option {
	case Entry, Title, AuthorMusic, AuthorLyric:
		return option
	default:
		return Entry
	}
}

func isValidSortingOption(option SortingOption) bool {
	if option == "" {
		return false
	}
	return normalizeSortingOption(string(option)) == option
}

func orderColumnForSongs(option SortingOption) string {
	switch option {
	case Title:
		return "title"
	case AuthorMusic:
		return "authorMusic"
	case AuthorLyric:
		return "authorLyric"
	default:
		return "entry"
	}
}

func orderColumnForSongs2(option SortingOption) string {
	switch option {
	case Title:
		return "title"
	default:
		return "entry"
	}
}
