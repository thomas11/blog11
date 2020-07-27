// Package blog11 is a static blog generator, because everyone needs to write
// one. With categories, markdown, atom feeds.
//
// To get started, copy example/exampleconf.go and customize it for
// your setup. Run something like example/build.sh to build your site.
//
// You need to provide your own templates.
//
// Thomas Kappler <http://www.thomaskappler.net/>
//
// This code is under BSD license. See license-bsd.txt.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/otiai10/copy"
)

type Site struct {
	posts       posts
	conf        *SiteConf
	renderCache map[string]string
}

func extractDateFromFilename(filename string, dateStampFormat string) (*time.Time, error) {
	if len(filename) < len(dateStampFormat)+1 {
		return nil, fmt.Errorf("skipping %v, name too short", filename)
	}

	dateStr := filename[:len(dateStampFormat)]
	date, err := time.Parse(dateStampFormat, dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date stamp in %v", dateStampFormat)
	}
	return &date, nil
}

func renderPostsListToFile(articles []*post, path string, tp templateParam, showTopicsLink bool, category category, engine templateEngine) error {
	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return engine.renderPostList(tp, articles, showTopicsLink, category.String(), outFile)
}

func ReadSite(conf *SiteConf, drafts bool) (*Site, error) {
	files, err := findPostFiles(conf.WritingDir, conf.WritingFileExtension)
	if err != nil {
		return nil, err
	}

	thisSite := Site{
		posts:       make(posts, 0, 100),
		conf:        conf,
		renderCache: make(map[string]string),
	}

	for _, f := range files {
		a, err := readPostFromFile(f, conf.WritingFileDateStampFormat)
		if err != nil {
			return nil, err
		}
		if drafts || !a.IsDraft() {
			thisSite.posts = append(thisSite.posts, a)
		}
	}

	// Order articles by date.
	sort.Slice(thisSite.posts, func(i, j int) bool { return thisSite.posts[i].Date.After(thisSite.posts[j].Date) })

	return &thisSite, nil
}

func (s *Site) RenderHtml() error {
	engine := newTemplateEngine(newMarkdownRenderer(), s.conf.TemplateDir)

	// Create a global template parameter holder. We'll re-use it for all
	// pages, overwriting the title.
	maxAgeForFrequentCategoriesInMonths := s.conf.MaxAgeForFrequentCategoriesInMonths
	if maxAgeForFrequentCategoriesInMonths == 0 {
		maxAgeForFrequentCategoriesInMonths = 24
	}

	minPostDate := time.Now().AddDate(0, -maxAgeForFrequentCategoriesInMonths, 0)
	postsRecentEnoughForFrequentCategories := s.posts.pruneOlderThan(minPostDate)
	globalTP := templateParam{
		FrequentCategories: groupByCategory(postsRecentEnoughForFrequentCategories).frequentCategories(
			s.conf.NumFrequentCategories,
			s.conf.MinArticlesForFrequentCategories),
	}

	// Render the articles.
	for _, a := range s.posts {
		outHtmlName := filepath.Join(s.conf.OutDir, a.ID+".html")
		var b bytes.Buffer
		globalTP.PageTitle = a.Title
		globalTP.FeedId = "index"
		globalTP.FileId = a.ID
		err, renderedBody := engine.renderPost(globalTP, a, &b)
		if err != nil {
			return err
		}
		ioutil.WriteFile(outHtmlName, b.Bytes(), os.FileMode(0664))

		s.renderCache[a.ID] = renderedBody
	}

	// Render the category pages.
	byCat := groupByCategory(s.posts)

	catDir := filepath.Join(s.conf.OutDir, s.conf.CategoriesOutDir)
	if _, err := os.Stat(catDir); os.IsNotExist(err) {
		err2 := os.Mkdir(catDir, os.FileMode(0775))
		if err2 != nil {
			log.Printf("Error creating directory %v: %v", catDir, err2)
		}
	}

	for _, c := range byCat {
		catId := c.Category.Id()
		outHtmlName := filepath.Join(catDir, catId+".html")
		globalTP.PageTitle = c.Category.String()
		globalTP.FeedId = catId
		globalTP.FileId = catId
		err := renderPostsListToFile(c.Posts, outHtmlName, globalTP, false, c.Category, engine)
		if err != nil {
			return err
		}
	}

	// Render the topics/categories overview page.
	var b bytes.Buffer
	globalTP.PageTitle = "Topics"
	globalTP.FeedId = "index"
	globalTP.FileId = "topics"
	err := engine.renderTopics(globalTP, byCat, &b)
	if err != nil {
		return err
	}
	outHtmlName := filepath.Join(s.conf.OutDir, globalTP.FileId+".html")
	ioutil.WriteFile(outHtmlName, b.Bytes(), os.FileMode(0664))

	// Render index.html with the last MaxArticlesOnIndex articles.
	articlesForIndex := s.posts
	haveMoreArticles := len(s.posts) > s.conf.MaxArticlesOnIndex
	if haveMoreArticles {
		articlesForIndex = articlesForIndex[:s.conf.MaxArticlesOnIndex]
	}
	globalTP.PageTitle = s.conf.SiteTitle
	globalTP.FeedId = "index"
	globalTP.FileId = "index"
	outHtmlName = filepath.Join(s.conf.OutDir, globalTP.FileId+".html")
	return renderPostsListToFile(articlesForIndex, outHtmlName, globalTP, haveMoreArticles, "", engine)
}

func (s *Site) RenderAll() error {
	err := s.RenderHtml()
	if err != nil {
		return err
	}
	return s.RenderAtom()
}

func (s *Site) CopyStaticFiles() error {
	srcDir := s.conf.StaticFilesDir
	dirName := filepath.Base(srcDir)
	dest := filepath.Join(s.conf.OutDir, dirName)
	log.Println("Recursively copying ", srcDir, " to ", dest)
	return copy.Copy(srcDir, dest)
}
