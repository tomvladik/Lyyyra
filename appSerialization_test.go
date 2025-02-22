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
				Title:   "ABCčDďE",
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
						{Name: "v1", Lines: "Chvaliž Hospodina, slávy vždy Krále mocného, ó\u00a0duše má, neboť tužba to srdce " +
							"je mého. Shromažďte se, harfy ať tón ozve se, zpívejte " +
							"chvalozpěv\u00a0jeho!"},
						{Name: "v2", Lines: "Chvaliž Hospodina, jenž všechno slavně spravuje, v\u00a0bezpečné náruči před " +
							"pádem tě ochraňuje a\u00a0vede tě Duchem své lásky v\u00a0světě, tvá duše to " +
							"pociťuje."},
					},
				}},
		},
		{
			name: "Simple song file - nonformatted XML -  is properly parsed.",
			args: args{
				xmlFilePath: `testdata/song-2.xml`,
			},
			wantErr: false,
			want: Song{
				Version: "0.8",
				Title:   "Kdo se vzdává cest, kterým vládne hřích",
				Songbook: Songbook{
					Name:  "EZ21",
					Entry: "1",
				},
				VerseOrder: "v1 v2 v3",
				Authors: []Author{
					{Type: "words", Value: "Miloslav Esterle"},
					{Type: "music", Value: "Pseaumes octante trois de David, 1551"},
				},
				Lyrics: Lyrics{
					Verses: []Verse{
						{Name: "v1", Lines: "Kdo se vzdává cest, kterým vládne hřích, kdo neřídí se radou bezbožných, ten, kdo se zříká lží a\u00a0posmívání, po Božích řádech ptá se bez přestání, kdo spravedlnost hledá den co den, požehnání smí nalézt v\u00a0díle\u00a0svém."},
						{Name: "v2", Lines: "S\u00a0vírou dál jak strom bude pevně stát a\u00a0z\u00a0čerstvých vod vždy novou sílu brát. Úrodou hojnou sytí hladového, v\u00a0žáru se stává stínem znaveného. Těm, kdo se trápí, chce být útěchou. Věrným dá Bůh znát cestu bezpečnou."},
						{Name: "v3", Lines: "Bezbožní však svůj život ztrácejí, když vlastní vůli věří raději. Jsou jako chmýří, kterým vítr zmítá, když Boží láska zůstává jim skrytá. Před soudem s\u00a0pýchou těžko obstojí, bez milosti svou duši nezhojí."},
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

// TestRemoveDiacritics tests the removeDiacritics function
func TestRemoveDiacritics(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"àèìòù", "aeiou"},
		{"âêîôû", "aeiou"},
		{"äëïöü", "aeiou"},
		{"ãñõ", "ano"},
		{"Çç", "Cc"},
		{"ÀÈÌÒÙ", "AEIOU"},
		{"ÂÊÎÔÛ", "AEIOU"},
		{"ÄËÏÖÜ", "AEIOU"},
		{"ÃÑÕ", "ANO"},
		{"ß", "ß"},
		{"Ææ", "Ææ"},
		{"Příliš žluťoučký kůň úpěl ďábelské ódy", "Prilis zlutoucky kun upel dabelske ody"},
		{"Loď čeří kýlem tůň - obzvlášť v Grónské úžině", "Lod ceri kylem tun - obzvlast v Gronske uzine"},
		{"Čtyři sta čtyřicet čtyři stříbrných stříkaček stříkalo přes čtyři sta čtyřicet čtyři stříbrných střech", "Ctyri sta ctyricet ctyri stribrnych strikacek strikalo pres ctyri sta ctyricet ctyri stribrnych strech"},
		{"Nezvyčajné kŕdle šťastných figliarskych ďatľov učia pri kótovanom ústí Váhu mĺkveho koňa Waldemara obžierať väčšie kusy exkluzívnej kôry s quesadillou.", "Nezvycajne krdle stastnych figliarskych datlov ucia pri kotovanom usti Vahu mlkveho kona Waldemara obzierat vacsie kusy exkluzivnej kory s quesadillou."},
		{"Vypätá dcéra grófa Maxwella s IQ nižším ako kôň núti čeľaď hrýzť hŕbu jabĺk.", "Vypata dcera grofa Maxwella s IQ nizsim ako kon nuti celad hryzt hrbu jablk."},
		{"Die heiße Zypernsonne quälte Max und Victoria ja böse auf dem Weg bis zur Küste.", "Die heiße Zypernsonne qualte Max und Victoria ja bose auf dem Weg bis zur Kuste."},
	}

	for _, test := range tests {
		result := removeDiacritics(test.input)
		if result != test.expected {
			t.Errorf("removeDiacritics(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

// TestIsMn tests the isMn function
func TestIsMn(t *testing.T) {
	tests := []struct {
		input    rune
		expected bool
	}{
		{'́', true},  // Combining Acute Accent
		{'̀', true},  // Combining Grave Accent
		{'̂', true},  // Combining Circumflex Accent
		{'̈', true},  // Combining Diaeresis
		{'̃', true},  // Combining Tilde
		{'a', false}, // Latin Small Letter A
		{'z', false}, // Latin Small Letter Z
		{'A', false}, // Latin Capital Letter A
		{'Z', false}, // Latin Capital Letter Z
	}

	for _, test := range tests {
		result := isMn(test.input)
		if result != test.expected {
			t.Errorf("isMn(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}
