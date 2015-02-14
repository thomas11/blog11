package blog11

import (
	"bytes"
	"fmt"
	"sort"
	"time"
)

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
