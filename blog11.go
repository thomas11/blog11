// A static blog generator, because everyone needs to write one. With
// categories, markdown, atom feeds.
//
// Thomas Kappler <http://www.thomaskappler.net/>
//
// This code is under BSD license. See license-bsd.txt.
package blog11

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type SiteConf struct {
	Author, AuthorUri string
	BaseUrl           string
	SiteTitle         string

	TemplateDir        string
	MaxArticlesOnIndex int

	WritingDir                 string
	WritingFileExtension       string
	WritingFileDateStampFormat string

	OutDir           string
	CategoriesOutDir string
	ImgOutDir        string
}

type category string

func (c category) String() string { return string(c) }

func (c category) Id() string { return strings.Replace(c.String(), " ", "_", -1) }

type article struct {
	Title, Id, Blurb string
	Date             time.Time
	Path             string
	Body             []byte
	Categories       []category
	Static           bool
}

func (a *article) String() string {
	b := new(bytes.Buffer)
	b.WriteString("title: ")
	b.WriteString(a.Title)
	b.WriteString("\ndate: ")
	b.WriteString(a.Date.String())
	b.WriteString("\nblurb: ")
	b.WriteString(a.Blurb)
	b.WriteString("\ncategories: ")
	fmt.Fprintln(b, a.Categories)

	body := a.Body
	if len(body) > 200 {
		body = append(body[:200], '.', '.', '.')
	}
	b.WriteString("\nbody: ")
	b.Write(body)

	return b.String()
}

type articles []*article

func (as articles) Len() int           { return len(as) }
func (as articles) Swap(i, j int)      { as[i], as[j] = as[j], as[i] }
func (as articles) Less(i, j int) bool { return as[i].Date.After(as[j].Date) }

type categoryArticles struct {
	category category
	articles []*article
}

type articlesByCategory []categoryArticles

// Order by number of articles per category, then by newest article.
func (s articlesByCategory) Len() int      { return len(s) }
func (s articlesByCategory) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s articlesByCategory) Less(i, j int) bool {
	li, lj := len(s[i].articles), len(s[j].articles)
	if li > lj {
		return true
	} else if lj > li {
		return false
	}

	var latestDate1, latestDate2 time.Time
	for _, article := range s[i].articles {
		if article.Date.After(latestDate1) {
			latestDate1 = article.Date
		}
	}
	for _, article := range s[j].articles {
		if article.Date.After(latestDate2) {
			latestDate2 = article.Date
		}
	}
	return latestDate1.After(latestDate2)
}

func (ac *articlesByCategory) addArticle(c category, a *article) {
	for i, cat := range *ac {
		if cat.category == c {
			cat.articles = append(cat.articles, a)
			(*ac)[i] = cat
			return
		}
	}

	newCA := categoryArticles{c, make([]*article, 1, 10)}
	newCA.articles[0] = a
	*ac = append(*ac, newCA)
}

func (ac articlesByCategory) String() string {
	b := new(bytes.Buffer)
	for _, c := range ac {
		b.WriteString(c.category.String())
		b.WriteString(": ")
		for i, a := range c.articles {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(a.Title)
		}
		b.WriteString("\n")
	}
	return b.String()
}

type Site struct {
	articles           articles
	articlesByCategory articlesByCategory
	conf               *SiteConf
	renderCache        map[string]string
}

// Return the most frequent n categories.
func (s *Site) frequentCategories(n, minArticles int) []category {
	frequent := make([]category, 0, n)
	for i, c := range s.articlesByCategory {
		if i == n || len(c.articles) < minArticles {
			break
		}
		frequent = append(frequent, c.category)
	}

	return frequent
}

func findArticleFiles(dir, fileExtension string) ([]string, error) {
	files := make([]string, 0, 100)

	myWalkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error at %v: %v\n", path, err)
			return nil
		}

		if !info.IsDir() && strings.HasSuffix(path, fileExtension) {
			files = append(files, path)
		}
		return nil
	}

	err := filepath.Walk(dir, myWalkFunc)
	return files, err
}

func readArticleFromFile(path, dateStampFormat string) (*article, error) {
	fileBaseName := filepath.Base(path)
	fileBaseName = fileBaseName[:len(fileBaseName)-len(filepath.Ext(fileBaseName))]

	fileContent, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	firstEmptyLine := bytes.Index(fileContent, []byte("\n\n"))
	if firstEmptyLine == -1 {
		return nil, fmt.Errorf("Weird article %v: no empty line.", path)
	}

	a := &article{
		Id:         fileBaseName,
		Body:       fileContent[firstEmptyLine+2:],
		Categories: make([]category, 0, 5),
	}

	headerLines := bytes.Split(fileContent[:firstEmptyLine], []byte("\n"))
	for _, l := range headerLines {
		if colon := bytes.Index(l, []byte(":")); colon != -1 {
			key, val := l[:colon], bytes.TrimSpace(l[colon+1:])
			switch string(key) {
			case "title":
				a.Title = string(val)
			case "blurb":
				a.Blurb = string(val)
			case "categories":
				for _, c := range bytes.Split(val, []byte(",")) {
					a.Categories = append(a.Categories, category(string(bytes.TrimSpace(c))))
				}
			case "static":
				a.Static = true
			default:
				fmt.Printf("  Skipping unknown header field %s in article %v\n", key, fileBaseName)
			}
		} else {
			return nil, fmt.Errorf("Invalid header line in article %v: %v", path, l)
		}
	}

	// Extract the date from the filename, if it's not a static page.
	if !a.Static {
		if len(fileBaseName) < len(dateStampFormat)+1 {
			return nil, fmt.Errorf("Skipping %v, name too short.", fileBaseName)
		}

		date, err := time.Parse(dateStampFormat, fileBaseName[:len(dateStampFormat)])
		if err != nil {
			return nil, fmt.Errorf("Invalid date stamp in %v", dateStampFormat)
		}
		a.Date = date
	}

	return a, nil
}

func renderArticleListToFile(articles []*article, path string, tp templateParam, engine templateEngine) error {
	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return engine.renderArticleList(tp, articles, outFile)
}

func ReadSite(conf *SiteConf) (*Site, error) {
	files, err := findArticleFiles(conf.WritingDir, conf.WritingFileExtension)
	if err != nil {
		return nil, err
	}

	thisSite := Site{
		articles:           make(articles, 0, 100),
		articlesByCategory: make(articlesByCategory, 0, 10),
		conf:               conf,
		renderCache:        make(map[string]string),
	}

	for _, f := range files {
		fmt.Println(f)
		a, err := readArticleFromFile(f, conf.WritingFileDateStampFormat)
		if err != nil {
			return nil, err
		}

		thisSite.articles = append(thisSite.articles, a)

		for _, cat := range a.Categories {
			thisSite.articlesByCategory.addArticle(cat, a)
		}
	}

	// Order articles by date.
	sort.Sort(thisSite.articles)
	// Order categories by the number of articles in them.
	sort.Sort(thisSite.articlesByCategory)

	return &thisSite, nil
}

func (s *Site) RenderHtml() error {
	engine := newTemplateEngine(newMarkdownRenderer(), s.conf.TemplateDir)

	// Create a global template parameter holder. We'll re-use for all
	// pages, overwriting the title.
	globalTP := templateParam{FrequentCategories: s.frequentCategories(6, 2)}

	// Render the articles.
	for _, a := range s.articles {
		outHtmlName := filepath.Join(s.conf.OutDir, a.Id+".html")
		var b bytes.Buffer
		globalTP.PageTitle = a.Title
		globalTP.FileId = "index"
		err, renderedBody := engine.renderArticle(globalTP, a, &b)
		if err != nil {
			return err
		}
		ioutil.WriteFile(outHtmlName, b.Bytes(), os.FileMode(0664))

		s.renderCache[a.Id] = renderedBody
	}

	// Render the category pages.
	catDir := filepath.Join(s.conf.OutDir, s.conf.CategoriesOutDir)
	if _, err := os.Stat(catDir); os.IsNotExist(err) {
		err2 := os.Mkdir(catDir, os.FileMode(0775))
		if err2 != nil {
			log.Printf("Error creating directory %v: %v", catDir, err2)
		}
	}

	for _, c := range s.articlesByCategory {
		catId := c.category.Id()
		outHtmlName := filepath.Join(catDir, catId+".html")
		globalTP.PageTitle = c.category.String()
		globalTP.FileId = catId
		err := renderArticleListToFile(c.articles, outHtmlName, globalTP, engine)
		if err != nil {
			return err
		}
	}

	// Render index.html with the last MaxArticlesOnIndex articles.
	articlesForIndex := s.articles
	if len(s.articles) > s.conf.MaxArticlesOnIndex {
		articlesForIndex = articlesForIndex[:s.conf.MaxArticlesOnIndex]
	}
	outHtmlName := filepath.Join(s.conf.OutDir, "index.html")
	globalTP.PageTitle = "index"
	globalTP.FileId = "index"
	return renderArticleListToFile(articlesForIndex, outHtmlName, globalTP, engine)
}

func (s *Site) RenderAll() error {
	err := s.RenderHtml()
	if err != nil {
		return err
	}
	return s.RenderAtom()
}
