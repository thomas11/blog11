// A static blog generator, because everyone needs to write one. With
// categories, markdown, atom feeds.
//
// To get started, copy example/exampleconf.go and customize it for
// your setup. Run something like example/build.sh to build your site.
//
// You need to provide your own templates.
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

	TemplateDir string

	WritingDir                 string
	WritingFileExtension       string
	WritingFileDateStampFormat string

	OutDir           string
	CategoriesOutDir string
	ImgOutDir        string

	MaxArticlesOnIndex           int
	NumFreqCategories            int
	MinArticlesForFreqCategories int
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

func (a *article) FormatDate() string {
	return formatDate(a.Date)
}

func (a *article) FormatDateShort() string {
	return formatDateShort(a.Date)
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

func (as articles) earliestDate() time.Time {
	t := time.Now()
	for _, a := range as {
		if a.Date.Before(t) {
			t = a.Date
		}
	}
	return t
}

func (as articles) latestDate() time.Time {
	var t time.Time
	for _, a := range as {
		if a.Date.After(t) {
			t = a.Date
		}
	}
	return t
}

func (as articles) byCategory() articlesByCategory {
	var t time.Time
	return as.byCategoryFrom(t)
}

func (as articles) byCategoryFrom(from time.Time) articlesByCategory {
	byCat := make(articlesByCategory, 0, 20)

	for _, art := range as {
		if art.Date.Before(from) {
			continue
		}
		for _, cat := range art.Categories {
			byCat.addArticle(cat, art)
		}
	}

	// Order categories by the number of articles in them.
	sort.Sort(byCat)

	return byCat
}

type categoryArticles struct {
	Category category
	Articles articles
}

func (c categoryArticles) EarliestDateFormatted() string {
	return formatDateShort(c.Articles.earliestDate())
}

func (c categoryArticles) LatestDateFormatted() string {
	return formatDateShort(c.Articles.latestDate())
}

type articlesByCategory []categoryArticles

// Order by number of articles per category, then by newest article.
func (s articlesByCategory) Len() int      { return len(s) }
func (s articlesByCategory) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s articlesByCategory) Less(i, j int) bool {
	li, lj := len(s[i].Articles), len(s[j].Articles)
	if li > lj {
		return true
	} else if lj > li {
		return false
	}

	latestDate1 := s[i].Articles.latestDate()
	latestDate2 := s[j].Articles.latestDate()
	return latestDate1.After(latestDate2)
}

func (ac *articlesByCategory) addArticle(c category, a *article) {
	for i, cat := range *ac {
		if cat.Category == c {
			cat.Articles = append(cat.Articles, a)
			(*ac)[i] = cat
			return
		}
	}

	newCA := categoryArticles{c, make([]*article, 1, 10)}
	newCA.Articles[0] = a
	*ac = append(*ac, newCA)
}

func (ac articlesByCategory) String() string {
	b := new(bytes.Buffer)
	for _, c := range ac {
		b.WriteString(c.Category.String())
		b.WriteString(": ")
		for i, a := range c.Articles {
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
	articles    articles
	conf        *SiteConf
	renderCache map[string]string
}

// Return the most frequent n categories.
func (ac articlesByCategory) frequentCategories(n, minArticles int) []category {
	frequent := make([]category, 0, n)
	for i, c := range ac {
		if i == n || len(c.Articles) < minArticles {
			break
		}
		frequent = append(frequent, c.Category)
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
			return nil, fmt.Errorf("Invalid header line in article %v: %s", path, l)
		}
	}

	// Extract the date from the filename, if it's not a static page.
	if !a.Static {
		if len(fileBaseName) < len(dateStampFormat)+1 {
			return nil, fmt.Errorf("Skipping %v, name too short.", fileBaseName)
		}

		dateStr := fileBaseName[:len(dateStampFormat)]
		date, err := time.Parse(dateStampFormat, dateStr)
		if err != nil {
			return nil, fmt.Errorf("Invalid date stamp in %v", dateStampFormat)
		}
		a.Date = date
	}

	return a, nil
}

func renderArticleListToFile(articles []*article, path string, tp templateParam, showTopicsLink bool, engine templateEngine) error {
	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return engine.renderArticleList(tp, articles, showTopicsLink, outFile)
}

func ReadSite(conf *SiteConf) (*Site, error) {
	files, err := findArticleFiles(conf.WritingDir, conf.WritingFileExtension)
	if err != nil {
		return nil, err
	}

	thisSite := Site{
		articles:    make(articles, 0, 100),
		conf:        conf,
		renderCache: make(map[string]string),
	}

	for _, f := range files {
		a, err := readArticleFromFile(f, conf.WritingFileDateStampFormat)
		if err != nil {
			return nil, err
		}
		thisSite.articles = append(thisSite.articles, a)
	}

	// Order articles by date.
	sort.Sort(thisSite.articles)

	return &thisSite, nil
}

func (s *Site) RenderHtml() error {
	engine := newTemplateEngine(newMarkdownRenderer(), s.conf.TemplateDir)

	// Create a global template parameter holder. We'll re-use for all
	// pages, overwriting the title.
	frequentCategoriesFrom := time.Now().Add(-1 * time.Hour * 24 * 365 * 2)
	globalTP := templateParam{
		FrequentCategories: s.articles.byCategoryFrom(frequentCategoriesFrom).frequentCategories(
			s.conf.NumFreqCategories,
			s.conf.MinArticlesForFreqCategories),
	}

	// Render the articles.
	for _, a := range s.articles {
		outHtmlName := filepath.Join(s.conf.OutDir, a.Id+".html")
		var b bytes.Buffer
		globalTP.PageTitle = a.Title
		globalTP.FeedId = "index"
		globalTP.FileId = a.Id
		err, renderedBody := engine.renderArticle(globalTP, a, &b)
		if err != nil {
			return err
		}
		ioutil.WriteFile(outHtmlName, b.Bytes(), os.FileMode(0664))

		s.renderCache[a.Id] = renderedBody
	}

	// Render the category pages.
	byCat := s.articles.byCategory()

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
		err := renderArticleListToFile(c.Articles, outHtmlName, globalTP, false, engine)
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
	articlesForIndex := s.articles
	haveMoreArticles := len(s.articles) > s.conf.MaxArticlesOnIndex
	if haveMoreArticles {
		articlesForIndex = articlesForIndex[:s.conf.MaxArticlesOnIndex]
	}
	globalTP.PageTitle = s.conf.SiteTitle
	globalTP.FeedId = "index"
	globalTP.FileId = "index"
	outHtmlName = filepath.Join(s.conf.OutDir, globalTP.FileId+".html")
	return renderArticleListToFile(articlesForIndex, outHtmlName, globalTP, haveMoreArticles, engine)
}

func (s *Site) RenderAll() error {
	err := s.RenderHtml()
	if err != nil {
		return err
	}
	return s.RenderAtom()
}
