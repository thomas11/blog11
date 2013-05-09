package blog11

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"os/exec"
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
	FileId             string
}

type articleTemplateParam struct {
	templateParam
	*article
	RenderedBody template.HTML
}

type articleListTemplateParam struct {
	templateParam
	Articles []*article
}

type topicsTemplateParam struct {
	templateParam
	ArticlesByCat articlesByCategory
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

func (te *templateEngine) renderArticle(tp templateParam, a *article, w io.Writer) (error, string) {
	body := highlightCode(a.Body)

	renderedBody := template.HTML(te.toHtml.render(body))
	p := articleTemplateParam{
		templateParam: tp,
		article:       a,
		RenderedBody:  renderedBody,
	}

	t := te.getTemplate("article.html")
	return t.Execute(w, p), string(renderedBody)
}

func (te *templateEngine) renderArticleList(tp templateParam, articles []*article, w io.Writer) error {
	p := articleListTemplateParam{templateParam: tp, Articles: articles}
	t := te.getTemplate("list.html")
	return t.Execute(w, p)
}

func (te *templateEngine) renderTopics(tp templateParam, topics articlesByCategory, w io.Writer) error {
	p := topicsTemplateParam{
		templateParam: tp,
		ArticlesByCat: topics,
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

//// EMACS SYNTAX HIGHLIGHTING, BROKEN ////

func elispToHtmlFontify(code, mode string) string {
	elisp := fmt.Sprintf(`(let ((buf (generate-new-buffer "highlight")))
	(set-buffer buf)
	(insert "%s")
	(%s)
	
	(htmlfontify-buffer)
	
	(set-buffer (concat (buffer-name buf) ".html"))
	(message (buffer-string)))`, code, mode)

	return elisp
}

func htmlFontifyWithEmacs(code, mode string) ([]byte, error) {
	elisp := elispToHtmlFontify(code, mode)
	cmd := exec.Command("/Users/thomas.kappler/software/emacs/trunk/nextstep/Emacs.app/Contents/MacOS/Emacs", "--batch", "--quick", "--eval", elisp)
	fmt.Println(cmd.Args)
	return cmd.CombinedOutput()
}
