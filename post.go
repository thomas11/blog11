package blog11

import (
	"bytes"
	"fmt"
	"time"
)

type post struct {
	Title, Id, Blurb string
	Date             time.Time
	Path             string
	Body             []byte
	Categories       []category
	Static           bool
}

// Called from templates
func (p *post) FormatDateShort() string {
	return formatDateShort(p.Date)
}

func (p *post) String() string {
	b := new(bytes.Buffer)
	b.WriteString("title: ")
	b.WriteString(p.Title)
	b.WriteString("\ndate: ")
	b.WriteString(p.Date.String())
	b.WriteString("\nblurb: ")
	b.WriteString(p.Blurb)
	b.WriteString("\ncategories: ")
	fmt.Fprintln(b, p.Categories)

	body := p.Body
	if len(body) > 200 {
		body = append(body[:200], '.', '.', '.')
	}
	b.WriteString("\nbody: ")
	b.Write(body)

	return b.String()
}

type posts []*post

func (ps posts) Len() int           { return len(ps) }
func (ps posts) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }
func (ps posts) Less(i, j int) bool { return ps[i].Date.After(ps[j].Date) }

func (ps posts) earliestDate() time.Time {
	t := time.Now()
	for _, a := range ps {
		if a.Date.Before(t) {
			t = a.Date
		}
	}
	return t
}

func (ps posts) latestDate() time.Time {
	var t time.Time
	for _, a := range ps {
		if a.Date.After(t) {
			t = a.Date
		}
	}
	return t
}
