/*
	this was just supposed to be a "simple" server
	to use instead of polymer serve

	fml
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var counter *Counter

func main() {
	var (
		addr string
		port int
		prod bool
	)
	flag.StringVar(
		&addr,
		"addr",
		"",
		"leave blank for 0.0.0.0",
	)
	flag.IntVar(
		&port,
		"port",
		8080,
		"",
	)
	flag.BoolVar(
		&prod,
		"prod",
		false,
		"production mode",
	)
	flag.Parse()

	if flag.NArg() > 0 && flag.Arg(0) == "precache" {
		zipPrecache()
		return
	}

	counter = NewCounter(".cache/caffo.db")
	defer counter.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("->", r.Method, r.URL)
		file := "index.html"
		req := strings.TrimPrefix(r.URL.Path, "/")

		if !prod {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}
		if prod {
			if req == "app.min.js" {
				file = "./dist/app.min.js"
			} else {
				file = "./dist/index.html"
			}
		}

		if fileExists(req) && !isDir(req) {
			file = req
		} else if strings.HasPrefix(req, "bower_components") {
			notfoundHandler(w, r)
			return
		} else if strings.HasPrefix(req, "image/") && strings.HasSuffix(req, ".json") {
			imageHandler(w, r)
			return
		} else if strings.HasPrefix(req, "download/") {
			zipHandler(w, r)
			return
		} else if req == "stats.json" {
			statsHandler(w, r)
			return
		} else if strings.HasPrefix(req, "nfo/") {
			// XXX TEMP
			tracklistId, _ := strconv.Atoi(strings.TrimPrefix(req, "nfo/"))
			rel, _ := getReleaseByTracklist(tracklistId)
			io.WriteString(w, createNfo(tracklistId, rel))
			return
		} else if strings.HasSuffix(r.URL.Path, "--square.jpg") {
			f, err := os.Open("./image/imagefallback.jpg")
			checkErr(err)
			defer f.Close()
			w.Header().Set("Content-Type", "image/jpeg")
			io.Copy(w, f)
			return
		}

		log.Println("<- 200 OK")
		if strings.HasSuffix(req, ".mp3") {
			counter.IncrementPlays(req, r.RemoteAddr)
		}
		http.ServeFile(w, r, file)
	})

	listenAddr := fmt.Sprintf("%s:%d", addr, port)
	log.Println("listening on", listenAddr)
	log.Fatalln(http.ListenAndServe(listenAddr, nil))
}

func notfoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	log.Println("<- 404 Not Found")
	if w.Header().Get("Content-Type") == "application/json; charset=UTF-8" {
		io.WriteString(w, `{"error":"file not found"}`)
	} else {
		io.WriteString(w, "File Not Found")
	}
}
