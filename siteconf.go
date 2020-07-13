package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
)

type SiteConf struct {
	Author, AuthorUri string
	BaseUrl           string
	SiteTitle         string

	TemplateDir string

	WritingDir                 string
	WritingFileExtension       string
	WritingFileDateStampFormat string
	StaticFilesDir             string

	OutDir           string
	CategoriesOutDir string

	MaxArticlesOnIndex                  int
	NumFrequentCategories               int
	MinArticlesForFrequentCategories    int
	MaxAgeForFrequentCategoriesInMonths int
}

func readConf(fileName string) *SiteConf {
	rawConf, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	conf := SiteConf{}

	if err = json.Unmarshal([]byte(rawConf), &conf); err != nil {
		log.Fatal(err)
	}

	// Populate with defaults
	if len(conf.StaticFilesDir) == 0 {
		conf.StaticFilesDir = filepath.Join(conf.WritingDir, "static")
	}
	if len(conf.TemplateDir) == 0 {
		conf.TemplateDir = "tmpl"
	}
	if len(conf.CategoriesOutDir) == 0 {
		conf.CategoriesOutDir = "categories"
	}

	// Normalize relative paths because the executable can be called from anywhere
	baseDir := filepath.Dir(fileName)
	conf.TemplateDir = normalizePath(conf.TemplateDir, baseDir)
	conf.WritingDir = normalizePath(conf.WritingDir, baseDir)
	conf.StaticFilesDir = normalizePath(conf.StaticFilesDir, baseDir)
	conf.OutDir = normalizePath(conf.OutDir, baseDir)
	conf.CategoriesOutDir = normalizePath(conf.CategoriesOutDir, baseDir)

	conf.TemplateDir, err = filepath.Abs(conf.TemplateDir)
	if err != nil {
		log.Fatal(err)
	}

	return &conf
}

func normalizePath(path, baseDir string) string {
	if !filepath.IsAbs(path) {
		absPath := filepath.Join(baseDir, path)
		log.Println("Normalizing ", path, " to ", absPath)
		return absPath
	}
	return path
}
