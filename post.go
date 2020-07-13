package main

import (
	"bytes"
	"fmt"
	"time"
)

type post struct {
	Title, ID, Blurb string
	Date             time.Time
	Path             string
	Flags            []string
	Body             []byte
	Categories       []category
}

func (p *post) IsStatic() bool {
	return p.hasFlag("static")
}

func (p *post) IsDraft() bool {
	return p.hasFlag("draft")
}

func (p *post) hasFlag(flag string) bool {
	for _, f := range p.Flags {
		if f == flag {
			return true
		}
	}
	return false
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

func (ps posts) pruneOlderThan(minDate time.Time) posts {
	pruned := make(posts, 0, 20)

	for _, p := range ps {
		if !p.Date.Before(minDate) {
			pruned = append(pruned, p)
		}
	}
	return pruned
}
