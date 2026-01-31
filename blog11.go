package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/radovskyb/watcher"
)

var siteConfPath = flag.String("siteConfPath", "blog11.json", "Path to the site configuration file")
var serve = flag.Bool("serve", false, "Start a localhost:9999 server for the site")
var watch = flag.Bool("watch", false, "Keep running and re-render the site on changes to the input directory.")
var drafts = flag.Bool("drafts", false, "Include articles with the 'draft' flag.")

func main() {
	flag.Parse()

	conf := readConf(*siteConfPath)

	renderSite(conf, *drafts)

	if *watch && *serve {
		// Run watcher in background while serving
		go rerenderOnChange(conf, *drafts)
	}

	if *serve {
		serveSite(conf.OutDir)
	} else if *watch {
		// Watch mode without serve: block on the watcher
		rerenderOnChange(conf, *drafts)
	}
}

func renderSite(conf *SiteConf, drafts bool) {
	site, err := ReadSite(conf, drafts)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Writing site to " + conf.OutDir)
	if err = site.RenderAll(); err != nil {
		log.Fatal(err)
	}
	if err = site.CopyStaticFiles(); err != nil {
		log.Fatal(err)
	}
}

func serveSite(dir string) {
	port := ":9999"

	http.Handle("/", http.FileServer(http.Dir(dir)))
	log.Printf("Serving %v on %v.", dir, port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func rerenderOnChange(siteConf *SiteConf, drafts bool) {
	log.Println("Watching " + siteConf.WritingDir + " for changes...")

	watcher := watcher.New()
	watcher.SetMaxEvents(1)

	go func() {
		for {
			select {
			case <-watcher.Event:
				renderSite(siteConf, drafts)
			case err := <-watcher.Error:
				log.Println(err)
			case <-watcher.Closed:
				return
			}
		}
	}()

	if err := watcher.AddRecursive(siteConf.WritingDir); err != nil {
		log.Fatalln(err)
	}

	if err := watcher.Start(time.Millisecond * 200); err != nil {
		log.Fatalln(err)
	}
}
