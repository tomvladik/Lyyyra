package app

const (
	// EZ (Evangelický zpěvník)
	XMLUrl_EZ            = "https://www.evangelickyzpevnik.cz.www.e-cirkev.cz/res/archive/001/000243.zip?download"
	PDFUrl               = "https://www.evangelickyzpevnik.cz/zpevnik/kapitoly-a-pisne/"
	ExpectedSongCount_EZ = 789
	Acronym_EZ           = "EZ"

	// KK (Katolický kancionál)
	XMLUrl_KK  = "https://stahuj.kancional.cz/opensong/pisne.zip"
	Acronym_KK = "KK"

	// Legacy constants (kept for backward compatibility)
	XMLUrl            = XMLUrl_EZ
	ExpectedSongCount = ExpectedSongCount_EZ
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
