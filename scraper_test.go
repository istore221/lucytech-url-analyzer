package analyzer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultHTMLScraper_Scrape(t *testing.T) {
	tests := []struct {
		name      string
		htmlInput string
		wantData  *ScrapeData
		wantErr   bool
	}{

		{
			name: "HTML5 Document with Title, Headings, Links, and Login Form",
			htmlInput: `<!DOCTYPE html>
<html>
<head>
    <title> Test Page Title </title>
    <link rel="stylesheet" href="styles.css">
</head>
<body>
   <h1>Heading 1</h1>
    <h1>Heading 1</h1>
<h2>Heading 2</h2>
<h3>Heading 3</h3>
<h4>Heading 4</h4>
<h5>Heading 5</h5>
<h6>Heading 6</h6>
<h6>Heading 6</h6>
<h6>Heading 6</h6>

    <a href="https://example.com">External Link</a>
    <a href="/about-us">Relative Link</a>
    <a href="#pricing">Fragment Link</a>
    <img src="https://example.com/logo.png" />
    <form action="/login" method="POST">
        <input type="text" name="username" />
        <input type="password" name="password" />
    </form>
</body>
</html>`,
			wantData: &ScrapeData{
				HTMLVersion: "HTML5",
				Title:       "Test Page Title",
				Headings: map[string]int{
					"h1": 2,
					"h2": 1,
					"h3": 1,
					"h4": 1,
					"h5": 1,
					"h6": 3,
				},
				RawLinks: []string{
					"styles.css",
					"https://example.com",
					"/about-us",
					"#pricing",
					"https://example.com/logo.png",
					"/login",
				},
				HasLoginForm: true,
			},
			wantErr: false,
		},
		{
			name: "HTML 4.01 Strict Doctype",
			htmlInput: `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">
<html>
<head><title>Legacy</title></head>
<body></body>
</html>`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML 4.01",
				Title:        "Legacy",
				Headings:     map[string]int{},
				RawLinks:     []string{},
				HasLoginForm: false,
			},
			wantErr: false,
		},
		{
			name: "XHTML 1.0 Doctype",
			htmlInput: `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html>
<head><title>XHTML</title></head>
<body></body>
</html>`,
			wantData: &ScrapeData{
				HTMLVersion:  "XHTML 1.0",
				Title:        "XHTML",
				Headings:     map[string]int{},
				RawLinks:     []string{},
				HasLoginForm: false,
			},
			wantErr: false,
		},
		{
			name: "Filter Javascript and Data URIs",
			htmlInput: `<html>
<body>
    <a href="javascript:void(0)">JS Link</a>
    <img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORG5CYII=" />
    <a href="https://valid.com">Valid Link</a>
</body>
</html>`,
			wantData: &ScrapeData{
				HTMLVersion: "HTML (Legacy/Unknown)",
				Title:       "",
				Headings:    map[string]int{},
				RawLinks: []string{
					"https://valid.com",
				},
				HasLoginForm: false,
			},
			wantErr: false,
		},
		{
			name: "Form Without Password Input",
			htmlInput: `<html>
<body>
    <form action="/search">
        <input type="text" name="q" />
        <button type="submit">Search</button>
    </form>
</body>
</html>`,
			wantData: &ScrapeData{
				HTMLVersion: "HTML (Legacy/Unknown)",
				Title:       "",
				Headings:    map[string]int{},
				RawLinks: []string{
					"/search",
				},
				HasLoginForm: false,
			},
			wantErr: false,
		},

		{
			name:      "tag <a> extracts href",
			htmlInput: `<a href="https://example.com/page">Link</a>`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []string{"https://example.com/page"},
				HasLoginForm: false,
			},
			wantErr: false,
		},
		{
			name:      "tag <link> extracts href",
			htmlInput: `<link href="/styles.css" rel="stylesheet">`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []string{"/styles.css"},
				HasLoginForm: false,
			},
			wantErr: false,
		},
		{
			name:      "tag <area> extracts href",
			htmlInput: `<area href="#map-section" shape="rect" coords="0,0,82,126">`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []string{"#map-section"},
				HasLoginForm: false,
			},
			wantErr: false,
		},
		{
			name:      "tag <base> extracts href",
			htmlInput: `<base href="https://example.com/">`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []string{"https://example.com/"},
				HasLoginForm: false,
			},
			wantErr: false,
		},

		{
			name:      "tag <script> extracts src",
			htmlInput: `<script src="/app.js"></script>`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []string{"/app.js"},
				HasLoginForm: false,
			},
			wantErr: false,
		},
		{
			name:      "tag <img> extracts src",
			htmlInput: `<img src="https://example.com/image.png" alt="Test">`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []string{"https://example.com/image.png"},
				HasLoginForm: false,
			},
			wantErr: false,
		},
		{
			name:      "tag <iframe> extracts src",
			htmlInput: `<iframe src="https://example.com/embed"></iframe>`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []string{"https://example.com/embed"},
				HasLoginForm: false,
			},
			wantErr: false,
		},
		{
			name:      "tag <embed> extracts src",
			htmlInput: `<embed src="/media/flash.swf">`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []string{"/media/flash.swf"},
				HasLoginForm: false,
			},
			wantErr: false,
		},
		{
			name:      "tag <source> extracts src",
			htmlInput: `<source src="/media/video.mp4" type="video/mp4">`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []string{"/media/video.mp4"},
				HasLoginForm: false,
			},
			wantErr: false,
		},
		{
			name:      "tag <track> extracts src",
			htmlInput: `<track src="/subtitles_en.vtt" kind="subtitles">`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []string{"/subtitles_en.vtt"},
				HasLoginForm: false,
			},
			wantErr: false,
		},

		{
			name:      "tag <form> extracts action",
			htmlInput: `<form action="/submit-form" method="POST"></form>`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []string{"/submit-form"},
				HasLoginForm: false,
			},
			wantErr: false,
		},

		// Group 4: targetAttr = "poster" (video)
		{
			name:      "tag <video> extracts poster",
			htmlInput: `<video poster="/images/thumbnail.jpg"></video>`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []string{"/images/thumbnail.jpg"},
				HasLoginForm: false,
			},
			wantErr: false,
		},

		{
			name:      "default tag <div> is ignored even with href",
			htmlInput: `<div href="/ignored">Content</div>`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []string{},
				HasLoginForm: false,
			},
			wantErr: false,
		},
		{
			name:      "default tag <p> is ignored even with src",
			htmlInput: `<p src="/ignored">Text</p>`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []string{},
				HasLoginForm: false,
			},
			wantErr: false,
		},
		{
			name:      "tag <video> ignores src (only targets poster attribute)",
			htmlInput: `<video src="/ignored-video.mp4"></video>`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []string{},
				HasLoginForm: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scraper := &DefaultHTMLScraper{}
			req := ScrapeRequest{
				Reader: strings.NewReader(tt.htmlInput),
			}

			got, err := scraper.Scrape(req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if got != nil && tt.wantData != nil {
				assert.Equal(t, tt.wantData.HTMLVersion, got.HTMLVersion)
				assert.Equal(t, tt.wantData.Title, got.Title)
				assert.Equal(t, tt.wantData.HasLoginForm, got.HasLoginForm)
				assert.Equal(t, tt.wantData.Headings, got.Headings)
				assert.Equal(t, tt.wantData.RawLinks, got.RawLinks)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Absolute URL",
			input:    "https://example.com/path",
			expected: "https://example.com/path",
		},
		{
			name:     "Relative URL",
			input:    "/about-us",
			expected: "/about-us",
		},
		{
			name:     "Fragment identifier",
			input:    "#section-1",
			expected: "#section-1",
		},
		{
			name:     "Query string",
			input:    "/search?q=test",
			expected: "/search?q=test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeURL(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
