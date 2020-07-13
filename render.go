package main

import (
	"bufio"
	"bytes"
	"html/template"
	"io"
	"log"
	"time"
)

func formatDate(d time.Time) string {
	return d.Format("January 2, 2006")
}

func formatDateShort(d time.Time) string {
	return d.Format("Jan 2, 2006")
}

type templateParam struct {
	PageTitle          string
	FrequentCategories []category
	// A short id such as a category name or "About"
	FileId string
	FeedId string
}

func (t templateParam) IdIs(id string) bool {
	return t.FileId == id
}

type postTemplateParam struct {
	templateParam
	*post
	RenderedBody template.HTML
}

type postListTemplateParam struct {
	templateParam
	PageHeading    string
	Posts          []*post
	ShowTopicsLink bool
}

type topicsTemplateParam struct {
	templateParam
	PostsByCategory postsByCategory
}

func (t topicsTemplateParam) Eq(a, b int) bool {
	return a == b
}

type renderer interface {
	render(in []byte) string
}

type templateEngine struct {
	toHtml        renderer
	templateDir   string
	templateCache map[string]*template.Template
}

func newTemplateEngine(r renderer, dir string) templateEngine {
	return templateEngine{
		toHtml:        r,
		templateDir:   dir,
		templateCache: make(map[string]*template.Template),
	}
}

func (te *templateEngine) renderPost(tp templateParam, a *post, w io.Writer) (error, string) {
	body := highlightCode(a.Body)

	renderedBody := template.HTML(te.toHtml.render(body))
	p := postTemplateParam{
		templateParam: tp,
		post:          a,
		RenderedBody:  renderedBody,
	}

	t := te.getTemplate("post.html")
	return t.Execute(w, p), string(renderedBody)
}

func (te *templateEngine) renderPostList(tp templateParam, posts []*post, showTopicsLink bool, pageHeading string, w io.Writer) error {
	p := postListTemplateParam{
		templateParam:  tp,
		PageHeading:    pageHeading,
		Posts:          posts,
		ShowTopicsLink: showTopicsLink,
	}
	t := te.getTemplate("list.html")
	return t.Execute(w, p)
}

func (te *templateEngine) renderTopics(tp templateParam, topics postsByCategory, w io.Writer) error {
	p := topicsTemplateParam{
		templateParam:   tp,
		PostsByCategory: topics,
	}
	t := te.getTemplate("topics.html")
	return t.Execute(w, p)
}

func (te *templateEngine) getTemplate(filename string) *template.Template {
	t, ok := te.templateCache[filename]
	if !ok {
		t = template.Must(template.ParseFiles(te.templateDir+"/global.html", te.templateDir+"/"+filename))
		te.templateCache[filename] = t
	}
	return t
}

// For now, just strip the highlighting directives.
func highlightCode(text []byte) []byte {
	newText := bytes.NewBuffer(make([]byte, 0, len(text)))
	r := bufio.NewReader(bytes.NewReader(text))

	for {
		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if !bytes.HasPrefix(bytes.TrimSpace(line), []byte("!highlight")) {
			_, err = newText.Write(line)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	return newText.Bytes()
}
