package app

import (
	"testing"
)

func TestNormalizeSortingOption(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected SortingOption
	}{
		{
			name:     "valid entry option",
			input:    "entry",
			expected: Entry,
		},
		{
			name:     "valid title option",
			input:    "title",
			expected: Title,
		},
		{
			name:     "valid author music option",
			input:    "authorMusic",
			expected: AuthorMusic,
		},
		{
			name:     "valid author lyric option",
			input:    "authorLyric",
			expected: AuthorLyric,
		},
		{
			name:     "invalid option defaults to entry",
			input:    "invalid",
			expected: Entry,
		},
		{
			name:     "empty string defaults to entry",
			input:    "",
			expected: Entry,
		},
		{
			name:     "whitespace trimmed",
			input:    "  title  ",
			expected: Title,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeSortingOption(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIsValidSortingOption(t *testing.T) {
	tests := []struct {
		name     string
		option   SortingOption
		expected bool
	}{
		{
			name:     "entry is valid",
			option:   Entry,
			expected: true,
		},
		{
			name:     "title is valid",
			option:   Title,
			expected: true,
		},
		{
			name:     "authorMusic is valid",
			option:   AuthorMusic,
			expected: true,
		},
		{
			name:     "authorLyric is valid",
			option:   AuthorLyric,
			expected: true,
		},
		{
			name:     "invalid option",
			option:   SortingOption("invalid"),
			expected: false,
		},
		{
			name:     "empty string is invalid",
			option:   SortingOption(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidSortingOption(tt.option)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestOrderColumnForSongs(t *testing.T) {
	tests := []struct {
		name     string
		option   SortingOption
		expected string
	}{
		{
			name:     "entry option",
			option:   Entry,
			expected: "entry",
		},
		{
			name:     "title option",
			option:   Title,
			expected: "title",
		},
		{
			name:     "authorMusic option",
			option:   AuthorMusic,
			expected: "authorMusic",
		},
		{
			name:     "authorLyric option",
			option:   AuthorLyric,
			expected: "authorLyric",
		},
		{
			name:     "invalid option defaults to entry",
			option:   SortingOption("invalid"),
			expected: "entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := orderColumnForSongs(tt.option)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestOrderColumnForSongs2(t *testing.T) {
	tests := []struct {
		name     string
		option   SortingOption
		expected string
	}{
		{
			name:     "title option",
			option:   Title,
			expected: "title",
		},
		{
			name:     "entry option defaults to entry",
			option:   Entry,
			expected: "entry",
		},
		{
			name:     "authorMusic defaults to entry",
			option:   AuthorMusic,
			expected: "entry",
		},
		{
			name:     "authorLyric defaults to entry",
			option:   AuthorLyric,
			expected: "entry",
		},
		{
			name:     "invalid option defaults to entry",
			option:   SortingOption("invalid"),
			expected: "entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := orderColumnForSongs2(tt.option)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
