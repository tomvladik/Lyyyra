package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_parseXmlSong(t *testing.T) {
	type args struct {
		xmlFilePath string
	}
	tests := []struct {
		name    string
		args    args
		want    Song
		wantErr bool
	}{
		{
			name: "Simple song file is properly parsed.",
			args: args{
				xmlFilePath: `testdata/song-1.xml`,
			},
			wantErr: false,
			want: Song{
				Version: "0.8",
				Title:   "ABCD",
				Songbook: Songbook{
					Name:  "EZ21",
					Entry: "288",
				},
				VerseOrder: "v1 v2",
				Authors: []Author{
					{Type: "words", Value: "Joachim, 1683\u00a0/ Harfa, 1915\u00a0/ zpěvník, 1979\u00a0/ Miloslav"},
					{Type: "music", Value: "Melica, 1674\u00a0/ Bundes-Psalmen, 1683"},
				},
				Lyrics: Lyrics{
					Verses: []Verse{
						{Name: "v1", Lines: "<lines>Chvaliž Hospodina, slávy vždy Krále mocného, ó\u00a0duše má, neboť tužba to srdce " +
							"je mého. Shromažďte se, harfy ať tón ozve se, zpívejte " +
							"chvalozpěv\u00a0jeho!</lines>"},
						{Name: "v2", Lines: "<lines>Chvaliž Hospodina, jenž všechno slavně spravuje, v\u00a0bezpečné náruči před " +
							"pádem tě ochraňuje a\u00a0vede tě Duchem své lásky v\u00a0světě, tvá duše to " +
							"pociťuje.</lines>"},
					},
				}},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentDir, _ := os.Getwd()
			fullPath := filepath.Join(currentDir, tt.args.xmlFilePath)
			got, err := parseXmlSong(fullPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseXmlSong() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(*got, tt.want) {
				t.Errorf("parseXmlSong() = \n%v\n, want:\n%v", *got, tt.want)
			}
		})
	}
}
