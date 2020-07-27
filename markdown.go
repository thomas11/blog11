package main

import (
	"regexp"

	"github.com/russross/blackfriday"
)

var extensions int

var emptyTopLevelInTocStart, emptyTopLevelInTocEnd *regexp.Regexp

func init() {
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH

	emptyTopLevelInTocStart = regexp.MustCompile(`<nav>\s*<ul>\s*<li>\s*<ul>\s*<li>`)
	emptyTopLevelInTocEnd = regexp.MustCompile(`</li>\s*</ul>\s*</nav>`)
}

func createHTMLFlags(generateToc bool) int {
	var htmlFlags int
	htmlFlags |= blackfriday.HTML_USE_XHTML
	htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES

	if generateToc {
		htmlFlags |= blackfriday.HTML_TOC
	}

	return htmlFlags
}

func newMarkdownRenderer() renderer {
	return &blackfridayHTMLRenderer{}
}

type blackfridayHTMLRenderer struct {
}

func (b *blackfridayHTMLRenderer) render(in []byte, generateToc bool) string {
	htmlFlags := createHTMLFlags(generateToc)
	r := blackfriday.HtmlRenderer(htmlFlags, "", "")
	html := string(blackfriday.Markdown(in, r, extensions))

	// Replace unnecessary nesting in ToC when we don't have an h1 heading
	if generateToc && emptyTopLevelInTocStart.MatchString(html) {
		html = emptyTopLevelInTocStart.ReplaceAllLiteralString(html, `<nav><ul><li>`)
		html = emptyTopLevelInTocEnd.ReplaceAllLiteralString(html, `</nav>`)
	}

	return html
}
