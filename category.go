package main

import (
	"bytes"
	"cmp"
	"slices"
	"strings"
)

type category string

func (c category) String() string { return string(c) }

func (c category) Id() string { return strings.ReplaceAll(c.String(), " ", "_") }

type categoryWithPosts struct {
	Category category
	Posts    posts
}

func (c categoryWithPosts) EarliestDateFormatted() string {
	return formatDateShort(c.Posts.earliestDate())
}

func (c categoryWithPosts) LatestDateFormatted() string {
	return formatDateShort(c.Posts.latestDate())
}

// Posts grouped by category. Sort sorts by number of articles per category, then by newest article.
// Create using the groupByCategory methods which sorts like this.
type postsByCategory []categoryWithPosts

func (pc *postsByCategory) addPost(c category, a *post) {
	for i, cat := range *pc {
		if cat.Category == c {
			cat.Posts = append(cat.Posts, a)
			(*pc)[i] = cat
			return
		}
	}

	newCategoryWithPosts := categoryWithPosts{c, make([]*post, 1, 10)}
	newCategoryWithPosts.Posts[0] = a
	*pc = append(*pc, newCategoryWithPosts)
}

func (pc postsByCategory) String() string {
	b := new(bytes.Buffer)
	for _, c := range pc {
		b.WriteString(c.Category.String())
		b.WriteString(": ")
		for i, a := range c.Posts {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(a.Title)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// Return the most frequent n categories.
func (pc postsByCategory) frequentCategories(n, minPosts int) []category {
	frequent := make([]category, 0, n)
	for i, c := range pc {
		if i == n || len(c.Posts) < minPosts {
			break
		}
		frequent = append(frequent, c.Category)
	}

	return frequent
}

func groupByCategory(posts posts) postsByCategory {
	byCat := make(postsByCategory, 0, 20)

	for _, post := range posts {
		for _, cat := range post.Categories {
			byCat.addPost(cat, post)
		}
	}

	// Order categories by the number of articles in them, then by newest article.
	slices.SortFunc(byCat, func(a, b categoryWithPosts) int {
		// More posts = comes first (descending order)
		if c := cmp.Compare(len(b.Posts), len(a.Posts)); c != 0 {
			return c
		}
		// If equal post count, newer comes first
		return b.Posts.latestDate().Compare(a.Posts.latestDate())
	})

	return byCat
}
