package main

import (
	"github.com/russross/blackfriday"
)

var htmlRenderer renderer

var htmlFlags, extensions int

func init() {
	htmlFlags |= blackfriday.HTML_USE_XHTML
	htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES

	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH
}

func newMarkdownRenderer() renderer {
	r := blackfriday.HtmlRenderer(htmlFlags, "", "")
	return &blackfridayHtmlRenderer{r, extensions}
}

type blackfridayHtmlRenderer struct {
	r          blackfriday.Renderer
	extensions int
}

func (b *blackfridayHtmlRenderer) render(in []byte) string {
	return string(blackfriday.Markdown(in, b.r, b.extensions))
}
