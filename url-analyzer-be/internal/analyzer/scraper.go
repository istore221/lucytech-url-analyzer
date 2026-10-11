package analyzer

import (
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type ScrapeRequest struct {
	Reader    io.Reader
	TargetURL *url.URL
}

type Scraper interface {
	Scrape(req ScrapeRequest) (*ScrapeData, error)
}

type DefaultHTMLScraper struct{}

func (s *DefaultHTMLScraper) Scrape(req ScrapeRequest) (*ScrapeData, error) {
	doc, err := html.Parse(req.Reader)
	if err != nil {
		return nil, err
	}

	data := &ScrapeData{
		HTMLVersion: "Unknown",
		Headings:    make(map[string]int),
		RawLinks:    make([]string, 0),
	}

	data.HTMLVersion = extractDoctype(doc)
	data.Title = extractHeadTitle(doc)

	s.walk(doc, data)

	return data, nil
}

func (s *DefaultHTMLScraper) walk(n *html.Node, data *ScrapeData) {
	if n.Type == html.ElementNode {
		tagName := strings.ToLower(n.Data)

		// 1. Record Headings (h1..h6)
		if len(tagName) == 2 && tagName[0] == 'h' && tagName[1] >= '1' && tagName[1] <= '6' {
			data.Headings[tagName]++
		}

		// 2. Extract, clean, and resolve links
		if linkURL := extractLinkFromNode(n, tagName); linkURL != "" {
			data.RawLinks = append(data.RawLinks, linkURL)
		}

		// 3. Check for Password Input inside <form>
		if tagName == "form" && !data.HasLoginForm {
			data.HasLoginForm = hasPasswordInput(n)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		s.walk(c, data)
	}
}

// Helper to target key attributes (href, src, action, poster) per HTML specification
func extractLinkFromNode(n *html.Node, tagName string) string {
	var targetAttr string

	switch tagName {
	case "a", "link", "area", "base":
		targetAttr = "href"
	case "script", "img", "iframe", "embed", "source", "track":
		targetAttr = "src"
	case "form":
		targetAttr = "action"
	case "video":
		targetAttr = "poster"
	default:
		return ""
	}

	for _, attr := range n.Attr {
		if strings.ToLower(attr.Key) == targetAttr {
			if attr.Val != "" {
				return normalizeURL(attr.Val)
			}
		}
	}

	return ""
}

func normalizeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	return parsed.String()
}

func extractHeadTitle(n *html.Node) string {
	headNode := findElement(n, "head")
	if headNode == nil {
		return ""
	}

	titleNode := findElement(headNode, "title")
	if titleNode == nil {
		return ""
	}

	return strings.TrimSpace(extractTextContent(titleNode))
}

func extractTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(extractTextContent(c))
	}
	return sb.String()
}

func findElement(n *html.Node, targetTag string) *html.Node {
	if n.Type == html.ElementNode && strings.ToLower(n.Data) == targetTag {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if res := findElement(c, targetTag); res != nil {
			return res
		}
	}
	return nil
}

func hasPasswordInput(formNode *html.Node) bool {
	if formNode.Type == html.ElementNode && strings.ToLower(formNode.Data) == "input" {
		for _, attr := range formNode.Attr {
			if strings.ToLower(attr.Key) == "type" && strings.ToLower(attr.Val) == "password" {
				return true
			}
		}
	}
	for c := formNode.FirstChild; c != nil; c = c.NextSibling {
		if hasPasswordInput(c) {
			return true
		}
	}
	return false
}

func extractDoctype(doc *html.Node) string {
	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.DoctypeNode {
			return parseDoctype(c)
		}
	}
	return "HTML (Legacy/Unknown)"
}

func parseDoctype(node *html.Node) string {
	var sb strings.Builder
	sb.WriteString(node.Data)

	for _, attr := range node.Attr {
		sb.WriteString(" ")
		sb.WriteString(attr.Val)
	}

	dt := strings.ToLower(sb.String())

	if strings.Contains(dt, "html 4.01") {
		return "HTML 4.01"
	}
	if strings.Contains(dt, "xhtml 1.0") {
		return "XHTML 1.0"
	}
	if strings.TrimSpace(strings.ToLower(node.Data)) == "html" {
		return "HTML5"
	}

	return "HTML (Legacy/Unknown)"
}
