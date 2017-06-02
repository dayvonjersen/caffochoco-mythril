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
	"strings"
)

func main() {
	var (
		addr          string
		port          int
		prod, nocache bool
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
	flag.BoolVar(
		&nocache,
		"nocache",
		false,
		"disable caching",
	)
	flag.Parse()

	if flag.NArg() > 0 && flag.Arg(0) == "precache" {
		zipPrecache()
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		fallbackFile := "index.html"
		if prod {
			fallbackFile = "dist/" + fallbackFile
		}
		file := fallbackFile

		path := strings.TrimPrefix(r.URL.Path, "/")
		dir := strings.Split(path, "/")[0]

		if prod &&
			fileExists(dir) &&
			(dir == "audio" || dir == "css" || dir == "font" || dir == "image" || dir == "svg" || dir == "video" || dir == "data.json") &&
			fileExists(path) && !isDir(path) {
			file = path
		} else if prod && fileExists("dist/"+dir) {
			file = "dist/" + dir
		} else if !prod && fileExists(path) && !isDir(path) {
			file = path
		} else if strings.HasPrefix(path, "image/") && strings.HasSuffix(path, ".json") {
			imageHandler(w, r)
			return
		} else if strings.HasPrefix(path, "download/") {
			zipHandler(w, r)
			return
		} else if strings.HasSuffix(r.URL.Path, "--square.jpg") {
			log.Println(req(r), "<- \033[34m200\033[0m OK (fallback)")
			f, err := os.Open("./image/imagefallback.jpg")
			checkErr(err)
			defer f.Close()
			w.Header().Set("Content-Type", "image/jpeg")
			io.Copy(w, f)
			return
		}

		if nocache {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}

		if file == fallbackFile {
			log.Println(req(r), "<- \033[34m200\033[0m OK (fallback)")
		} else {
			log.Println(req(r), "<- \033[32m200\033[0m OK")
		}
		http.ServeFile(w, r, file)

	})

	listenAddr := fmt.Sprintf("%s:%d", addr, port)
	if prod {
		log.Println("listening on", listenAddr, "in \033[32mproduction mode\033[0m")
	} else {
		log.Println("listening on", listenAddr, "in development mode")
	}
	log.Fatalln(http.ListenAndServe(listenAddr, nil))
}

func notfoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	log.Println(req(r), "<- \033[33m404\033[0m Not Found")
	if w.Header().Get("Content-Type") == "application/json; charset=UTF-8" {
		io.WriteString(w, `{"error":"file not found"}`)
	} else {
		io.WriteString(w, "File Not Found")
	}
}

func req(r *http.Request) string {
	return fmt.Sprint(
		r.Host, " ", r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")], " ",
		r.Header.Get("User-Agent"),
		"\n -> ", r.Method, " ", r.URL, "\n",
	)
}
