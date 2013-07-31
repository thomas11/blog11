package main

import (
	"github.com/thomas11/blog11"
	"log"
)

const siteUrl = "http://example.com/"

var conf = blog11.SiteConf{
	Author:                       "Joe User",
	AuthorUri:                    siteUrl,
	BaseUrl:                      siteUrl,
	SiteTitle:                    "Joe User's site.",
	CategoriesOutDir:             "categories",
	WritingFileExtension:         ".text",
	WritingFileDateStampFormat:   "2006-01-02",
	ImgOutDir:                    "img",
	WritingDir:                   "../writing",
	OutDir:                       "out",
	TemplateDir:                  "tmpl",
	MaxArticlesOnIndex:           11,
	NumFreqCategories:            6,
	MinArticlesForFreqCategories: 2,
}

func main() {
	site, err := blog11.ReadSite(&conf)
	if err != nil {
		log.Fatal(err)
	}

	err = site.RenderAll()
	if err != nil {
		log.Fatal(err)
	}
}
