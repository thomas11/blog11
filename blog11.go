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

func main() {
	flag.Parse()

	conf := readConf(*siteConfPath)

	renderSite(conf)

	if *watch {
		go rerenderOnChange(conf)
	}

	if *serve {
		serveSite(conf.OutDir)
	}
}

func renderSite(conf *SiteConf) {
	site, err := ReadSite(conf)
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
	http.ListenAndServe(port, nil)
}

func rerenderOnChange(siteConf *SiteConf) {
	log.Println("Watching " + siteConf.WritingDir + " for changes...")

	watcher := watcher.New()
	watcher.SetMaxEvents(1)

	go func() {
		for {
			select {
			case _ = <-watcher.Event:
				renderSite(siteConf)
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
