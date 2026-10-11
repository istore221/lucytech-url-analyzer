package analyzer

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type errReader struct{}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestDefaultHTMLScraper_Scrape_MoreCases(t *testing.T) {
	tests := []struct {
		name      string
		htmlInput string
		wantData  *ScrapeData
		assertErr assert.ErrorAssertionFunc
	}{
		{
			name:      "Empty Document",
			htmlInput: "",
			wantData: &ScrapeData{
				HTMLVersion:  "HTML (Legacy/Unknown)",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []Link{},
				HasLoginForm: false,
			},
			assertErr: assert.NoError,
		},
		{
			name: "Title Outside Head Tag",
			htmlInput: `<!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="utf-8">
        </head>
        <body>
            <title>Body Title</title>
        </body>
        </html>`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML5",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []Link{},
				HasLoginForm: false,
			},
			assertErr: assert.NoError,
		},
		{
			name: "Head without Title Tag",
			htmlInput: `<!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="utf-8">
        </head>
        <body></body>
        </html>`,
			wantData: &ScrapeData{
				HTMLVersion:  "HTML5",
				Title:        "",
				Headings:     map[string]int{},
				RawLinks:     []Link{},
				HasLoginForm: false,
			},
			assertErr: assert.NoError,
		},
		{
			name: "Headings Beyond H6 ignored",
			htmlInput: `<!DOCTYPE html>
                                <html lang="en">
                                <head>
                                    <meta charset="utf-8">
                                    <title>Headings Test</title>
                                </head>
                                <body>
                                    <h7>Not a Heading</h7>
                                    <h1>Valid H1</h1>
                                    <h0>Not a Heading</h0>
                                </body>
                            </html>`,
			wantData: &ScrapeData{
				HTMLVersion: "HTML5",
				Title:       "Headings Test",
				Headings: map[string]int{
					"h1": 1,
				},
				RawLinks:     []Link{},
				HasLoginForm: false,
			},
			assertErr: assert.NoError,
		},
		{
			name: "Links with Empty or Whitespace Attr",
			htmlInput: `<!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="utf-8">
            <title>Links Test</title>
        </head>
        <body>
            <a href="">Empty Link</a>
            <a href="   ">Whitespace Link</a>
            <a href="  https://example.com/trimmed  ">Trimmed Link</a>
        </body>
        </html>`,
			wantData: &ScrapeData{
				HTMLVersion: "HTML5",
				Title:       "Links Test",
				Headings:    map[string]int{},
				RawLinks: []Link{
					ScrapedLink("%20%20%20"),
					ScrapedLink("  https://example.com/trimmed  "),
				},
				HasLoginForm: false,
			},
			assertErr: assert.NoError,
		},
		{
			name: "Case Insensitivity for Tags, Attributes, and Doctype",
			htmlInput: `<!DOCTYPE html>
                <HTML lang="en">
                <HEAD>
                    <META CHARSET="utf-8">
                    <TITLE>Uppercase Title</TITLE>
                </HEAD>
                <BODY>
                    <H1>Uppercase Heading</H1>
                    <A HREF="https://example.com/upper">Link</A>
                    <FORM ACTION="/login-upper">
                        <INPUT TYPE="PASSWORD" NAME="pass">
                    </FORM>
                </BODY>
                </HTML>`,
			wantData: &ScrapeData{
				HTMLVersion: "HTML5",
				Title:       "Uppercase Title",
				Headings: map[string]int{
					"h1": 1,
				},
				RawLinks: []Link{
					ScrapedLink("https://example.com/upper"),
					ScrapedLink("/login-upper"),
				},
				HasLoginForm: true,
			},
			assertErr: assert.NoError,
		},
		{
			name: "Nested Forms with Password Input in Second Form",
			htmlInput: `<!DOCTYPE html>
                <html lang="en">
                <head>
                    <meta charset="utf-8">
                    <title>Forms Test</title>
                </head>
                <body>
                    <form action="/search">
                        <input type="text" name="q">
                    </form>
                    <form action="/login">
                        <div>
                            <div>
                                <input type="password" name="pwd">
                            </div>
                        </div>
                    </form>
                </body>
                </html>`,
			wantData: &ScrapeData{
				HTMLVersion: "HTML5",
				Title:       "Forms Test",
				Headings:    map[string]int{},
				RawLinks: []Link{
					ScrapedLink("/search"),
					ScrapedLink("/login"),
				},
				HasLoginForm: true,
			},
			assertErr: assert.NoError,
		},
		{
			name: "Form containing Input non-password",
			htmlInput: `<!DOCTYPE html>
                <html lang="en">
                <head>
                    <meta charset="utf-8">
                    <title>Register Test</title>
                </head>
                <body>
                    <form action="/register">
                        <input type="email" name="email">
                        <input type="text" name="username">
                    </form>
                </body>
                </html>`,
			wantData: &ScrapeData{
				HTMLVersion: "HTML5",
				Title:       "Register Test",
				Headings:    map[string]int{},
				RawLinks: []Link{
					ScrapedLink("/register"),
				},
				HasLoginForm: false,
			},
			assertErr: assert.NoError,
		},
		{
			name: "JavaScript and Data scheme Case Insensitivity",
			htmlInput: `<!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="utf-8">
            <title>Schemes Test</title>
        </head>
        <body>
            <a href="JAVASCRIPT:alert(1)">Upper JS</a>
            <img src="DATA:image/png;base64,123" alt="Data URI Image">
        </body>
        </html>`,
			wantData: &ScrapeData{
				HTMLVersion: "HTML5",
				Title:       "Schemes Test",
				Headings:    map[string]int{},
				RawLinks: []Link{
					ScrapedLink("javascript:alert(1)"),
					ScrapedLink("data:image/png;base64,123"),
				},
				HasLoginForm: false,
			},
			assertErr: assert.NoError,
		},
		{
			name: "Cover href targetAttr elements (link, area, base)",
			htmlInput: `<!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="utf-8">
            <base href="https://example.com/api/">
            <link rel="stylesheet" href="/styles.css">
            <title>Href Attributes Test</title>
        </head>
        <body>
            <map name="primary">
                <area shape="rect" coords="0,0,82,126" href="/details">
            </map>
        </body>
        </html>`,
			wantData: &ScrapeData{
				HTMLVersion: "HTML5",
				Title:       "Href Attributes Test",
				Headings:    map[string]int{},
				RawLinks: []Link{
					ScrapedLink("https://example.com/api/"),
					ScrapedLink("/styles.css"),
					ScrapedLink("/details"),
				},
				HasLoginForm: false,
			},
			assertErr: assert.NoError,
		},
		{
			name: "Cover src targetAttr elements (script, img, iframe, embed, source, track)",
			htmlInput: `<!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="utf-8">
            <script src="/app.js"></script>
            <title>Src Attributes Test</title>
        </head>
        <body>
            <img src="/logo.png" alt="Logo">
            <iframe src="https://example.com/embed"></iframe>
            <embed src="/media.swf">
            <video>
                <source src="/video.mp4" type="video/mp4">
                <track src="/subtitles.vtt" kind="subtitles">
            </video>
        </body>
        </html>`,
			wantData: &ScrapeData{
				HTMLVersion: "HTML5",
				Title:       "Src Attributes Test",
				Headings:    map[string]int{},
				RawLinks: []Link{
					ScrapedLink("/app.js"),
					ScrapedLink("/logo.png"),
					ScrapedLink("https://example.com/embed"),
					ScrapedLink("/media.swf"),
					ScrapedLink("/video.mp4"),
					ScrapedLink("/subtitles.vtt"),
				},
				HasLoginForm: false,
			},
			assertErr: assert.NoError,
		},
		{
			name: "Cover poster targetAttr element (video)",
			htmlInput: `<!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="utf-8">
            <title>Poster Attribute Test</title>
        </head>
        <body>
            <video poster="/thumbnail.jpg">
                <source src="/video.mp4" type="video/mp4">
            </video>
        </body>
        </html>`,
			wantData: &ScrapeData{
				HTMLVersion: "HTML5",
				Title:       "Poster Attribute Test",
				Headings:    map[string]int{},
				RawLinks: []Link{
					ScrapedLink("/thumbnail.jpg"),
					ScrapedLink("/video.mp4"),
				},
				HasLoginForm: false,
			},
			assertErr: assert.NoError,
		},
		{
			name: "Ignored attributes on non-matching elements",
			htmlInput: `<!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="utf-8">
            <title>Ignored Attributes Test</title>
        </head>
        <body>
            <div href="/ignored-href" src="/ignored-src">Div</div>
            <span action="/ignored-action">Span</span>
            <a src="/ignored-src-on-a" href="/valid-href">Link</a>
        </body>
        </html>`,
			wantData: &ScrapeData{
				HTMLVersion: "HTML5",
				Title:       "Ignored Attributes Test",
				Headings:    map[string]int{},
				RawLinks: []Link{
					ScrapedLink("/valid-href"),
				},
				HasLoginForm: false,
			},
			assertErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scraper := &DefaultHTMLScraper{}
			req := ScrapeRequest{
				Reader: strings.NewReader(tt.htmlInput),
			}

			got, err := scraper.Scrape(req)
			tt.assertErr(t, err)

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

func TestDefaultHTMLScraper_Scrape_ReaderError(t *testing.T) {
	scraper := &DefaultHTMLScraper{}
	req := ScrapeRequest{
		Reader: &errReader{},
	}

	got, err := scraper.Scrape(req)
	assert.Error(t, err)
	assert.Nil(t, got)
}
