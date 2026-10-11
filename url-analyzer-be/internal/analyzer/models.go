package analyzer

type PageAnalysis struct {
	URL               string
	HTMLVersion       string
	Title             string
	Headings          map[string]int
	InternalLinks     []string
	ExternalLinks     []string
	InaccessibleLinks []string
	HasLoginForm      bool
}

type ScrapeData struct {
	HTMLVersion  string
	Title        string
	Headings     map[string]int
	RawLinks     []Link
	HasLoginForm bool
}
