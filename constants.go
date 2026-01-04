package main

const (
	XMLUrl            = "https://www.evangelickyzpevnik.cz.www.e-cirkev.cz/res/archive/001/000243.zip?download"
	PDFUrl            = "https://www.evangelickyzpevnik.cz/zpevnik/kapitoly-a-pisne/"
	ExpectedSongCount = 789
)

type SupplementalPDF struct {
	URL      string
	FileName string
}

var SupplementalPDFs = []SupplementalPDF{
	{
		URL:      "https://www.evangelickyzpevnik.cz.www.e-cirkev.cz/res/archive/001/000234.pdf",
		FileName: "kytara.pdf",
	},
	{
		URL:      "https://www.evangelickyzpevnik.cz.www.e-cirkev.cz/res/archive/001/000208.pdf",
		FileName: "choralnik.pdf",
	},
}
