package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func findPostFiles(dir, fileExtension string) ([]string, error) {
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

func readPostFromFile(path, dateStampFormat string) (*post, error) {
	fileBaseName := filepath.Base(path)
	fileBaseName = fileBaseName[:len(fileBaseName)-len(filepath.Ext(fileBaseName))]

	fileContent, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	firstEmptyLine := bytes.Index(fileContent, []byte("\n\n"))
	if firstEmptyLine == -1 {
		firstEmptyLine = bytes.Index(fileContent, []byte("\r\n\r\n"))

		if firstEmptyLine == -1 {
			return nil, fmt.Errorf("weird article %v: no empty line", path)
		}
	}

	a := &post{
		ID:         fileBaseName,
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
			case "flags":
				a.Flags = strings.Split(string(val), ",")
			default:
				fmt.Printf("  Skipping unknown header field %s in article %v\n", key, fileBaseName)
			}
		} else {
			return nil, fmt.Errorf("invalid header line in article %v: %s", path, l)
		}
	}

	if !a.IsStatic() {
		date, err := extractDateFromFilename(fileBaseName, dateStampFormat)
		if err != nil {
			return nil, err
		}
		a.Date = *date
	}

	return a, nil
}
