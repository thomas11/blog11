package main

import (
	"bytes"
	"sort"
	"strings"
)

type category string

func (c category) String() string { return string(c) }

func (c category) Id() string { return strings.Replace(c.String(), " ", "_", -1) }

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

// Order
func (pc postsByCategory) Len() int      { return len(pc) }
func (pc postsByCategory) Swap(i, j int) { pc[i], pc[j] = pc[j], pc[i] }
func (pc postsByCategory) Less(i, j int) bool {
	li, lj := len(pc[i].Posts), len(pc[j].Posts)
	if li > lj {
		return true
	} else if lj > li {
		return false
	}

	latestDate1 := pc[i].Posts.latestDate()
	latestDate2 := pc[j].Posts.latestDate()
	return latestDate1.After(latestDate2)
}

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

	// Order categories by the number of articles in them.
	sort.Sort(byCat)

	return byCat
}
