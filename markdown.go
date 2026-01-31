package main

import (
	"regexp"

	"github.com/russross/blackfriday/v2"
)

var extensions = blackfriday.NoIntraEmphasis |
	blackfriday.Tables |
	blackfriday.FencedCode |
	blackfriday.Autolink |
	blackfriday.Strikethrough

var emptyTopLevelInTocStart = regexp.MustCompile(`<nav>\s*<ul>\s*<li>\s*<ul>\s*<li>`)
var emptyTopLevelInTocEnd = regexp.MustCompile(`</li>\s*</ul>\s*</nav>`)

func createHTMLFlags(generateToc bool) blackfriday.HTMLFlags {
	htmlFlags := blackfriday.UseXHTML |
		blackfriday.Smartypants |
		blackfriday.SmartypantsFractions |
		blackfriday.SmartypantsLatexDashes

	if generateToc {
		htmlFlags |= blackfriday.TOC
	}

	return htmlFlags
}

func newMarkdownRenderer() renderer {
	return &blackfridayHTMLRenderer{}
}

type blackfridayHTMLRenderer struct{}

func (b *blackfridayHTMLRenderer) render(in []byte, generateToc bool) string {
	htmlFlags := createHTMLFlags(generateToc)
	r := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: htmlFlags,
	})
	html := string(blackfriday.Run(in,
		blackfriday.WithExtensions(extensions),
		blackfriday.WithRenderer(r)))

	// Replace unnecessary nesting in ToC when we don't have an h1 heading
	if generateToc && emptyTopLevelInTocStart.MatchString(html) {
		html = emptyTopLevelInTocStart.ReplaceAllLiteralString(html, `<nav><ul><li>`)
		html = emptyTopLevelInTocEnd.ReplaceAllLiteralString(html, `</nav>`)
	}

	return html
}
