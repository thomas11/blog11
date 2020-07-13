package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	atom "github.com/thomas11/atomgenerator"
)

func (s *Site) RenderAtom() error {
	filePath := filepath.Join(s.conf.OutDir, "index.xml")
	err := s.renderAndSaveFeed(s.conf.SiteTitle, "", filePath, s.posts)
	if err != nil {
		return err
	}

	return s.renderAndSaveCategoriesAtom()
}

func (s *Site) renderFeed(title, relURL string, articles []*post) ([]byte, error) {
	feedURL := s.conf.BaseURL
	if len(relURL) > 0 {
		if relURL[0] == '/' {
			relURL = relURL[1:]
		}
		feedURL += relURL
	}

	feed := atom.Feed{
		Title:   title,
		Link:    feedURL,
		PubDate: time.Now(),
	}
	feed.AddAuthor(atom.Author{
		Name: s.conf.Author,
		Uri:  s.conf.AuthorURI,
	})

	for _, article := range articles {
		if article.IsStatic() {
			continue
		}
		feed.AddEntry(s.entryForArticle(article))
	}

	errs := feed.Validate()
	if len(errs) > 0 {
		log.Println("Atom feed is not valid!")
		for _, e := range errs {
			log.Println(e.Error())
		}
		return nil, errs[0]
	}

	return feed.GenXml()
}

func (s *Site) entryForArticle(article *post) *atom.Entry {
	e := &atom.Entry{
		Title:       article.Title,
		Description: article.Blurb,
		Link:        s.conf.BaseURL + article.ID + ".html",
		PubDate:     article.Date,
	}

	for _, cat := range article.Categories {
		e.AddCategory(atom.Category{Term: string(cat)})
	}

	if renderedBody, ok := s.renderCache[article.ID]; ok {
		e.Content = renderedBody
	}

	return e
}

func (s *Site) renderAndSaveFeed(title, relURL, filePath string, articles []*post) error {
	atomXML, err := s.renderFeed(title, relURL, articles)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, atomXML, os.FileMode(0664))
}

func (s *Site) renderAndSaveCategoriesAtom() error {
	for _, catArticles := range groupByCategory(s.posts) {
		category := catArticles.Category
		title := s.conf.SiteTitle + ` Category "` + category.String() + `."`
		urlPath := s.conf.CategoriesOutDir + "/" + category.Id() + "/"
		filePath := filepath.Join(s.conf.OutDir, s.conf.CategoriesOutDir, category.Id()+".xml")

		err := s.renderAndSaveFeed(title, urlPath, filePath, catArticles.Posts)
		if err != nil {
			return err
		}
	}
	return nil
}
