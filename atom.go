package blog11

import (
	atom "github.com/thomas11/atomgenerator"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

func (s *Site) renderFeed(title, relUrl string, articles []*article) ([]byte, error) {
	feedUrl := s.conf.BaseUrl
	if len(relUrl) > 0 {
		if relUrl[0] == '/' {
			relUrl = relUrl[1:]
		}
		feedUrl += relUrl
	}

	feed := atom.Feed{
		Title:   title,
		Link:    feedUrl,
		PubDate: time.Now(),
	}
	feed.AddAuthor(atom.Author{
		Name: s.conf.Author,
		Uri:  s.conf.AuthorUri,
	})

	for _, article := range articles {
		if article.Static {
			continue
		}
		e := &atom.Entry{
			Title:       article.Title,
			Description: article.Blurb,
			Link:        s.conf.BaseUrl + article.Id + ".html",
			PubDate:     article.Date,
		}

		for _, cat := range article.Categories {
			e.AddCategory(atom.Category{Term: string(cat)})
		}

		if renderedBody, ok := s.renderCache[article.Id]; ok {
			e.Content = renderedBody
		}
		feed.AddEntry(e)
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

func (s *Site) renderAndSaveFeed(title, relUrl, filePath string, articles []*article) error {
	atomXml, err := s.renderFeed(title, relUrl, articles)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, atomXml, os.FileMode(0664))
}

func (s *Site) RenderAtom() error {
	filePath := filepath.Join(s.conf.OutDir, "index.xml")
	err := s.renderAndSaveFeed(s.conf.SiteTitle, "", filePath, s.articles)
	if err != nil {
		return err
	}

	return s.renderAndSaveCategoriesAtom()
}

func (s *Site) renderAndSaveCategoriesAtom() error {
	for _, catArticles := range s.articles.byCategory() {
		cat := string(catArticles.Category)
		title := s.conf.SiteTitle + ` Category "` + cat + `."`
		urlPath := s.conf.CategoriesOutDir + "/" + cat + "/"
		filePath := filepath.Join(s.conf.OutDir, s.conf.CategoriesOutDir, cat+".xml")

		err := s.renderAndSaveFeed(title, urlPath, filePath, catArticles.Articles)
		if err != nil {
			return err
		}
	}
	return nil
}
