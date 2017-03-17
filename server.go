package main

import (
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	"github.com/generaltso/vibrant"
)

func main() {
	var (
		addr string
		port int
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
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("->", r.Method, r.URL)
		file := "index.html"
		req := strings.TrimPrefix(r.URL.Path, "/")

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		if fileExists(req) {
			file = req
		} else if strings.HasPrefix(req, "bower_components/") {
			w.WriteHeader(404)
			log.Println("<- 404 Not Found")
			io.WriteString(w, "File Not Found")
			return
		} else if strings.HasPrefix(req, "image/") && strings.HasSuffix(req, ".css") {
			if !fileExists(strings.TrimSuffix(req, ".css")) {
				log.Println("<- 404 Not Found")
				io.WriteString(w, "File Not Found")
				return
			}
			imageHandler(w, r)
			log.Println("<- 200 OK")
			return
		}
		log.Println("<- 200 OK")
		http.ServeFile(w, r, file)
	})

	listenAddr := fmt.Sprintf("%s:%d", addr, port)
	log.Println("listening on", listenAddr)
	log.Fatalln(http.ListenAndServe(listenAddr, nil))

}

type file struct {
	modtime    int64
	stylesheet string
}

var imageCache = map[string]*file{}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=UTF-8")

	path := "." + r.URL.Path
	path = strings.TrimSuffix(path, ".css")

	f, err := os.Open(path)
	checkErr(err)
	defer f.Close()

	info, err := f.Stat()
	checkErr(err)
	modtime := info.ModTime().Unix()

	if cached, ok := imageCache[path]; ok {
		if cached.modtime == modtime {
			fmt.Fprintf(w, cached.stylesheet)
			return
		}
	}

	img, _, err := image.Decode(f)
	checkErr(err)
	palette, err := vibrant.NewPaletteFromImage(img)
	checkErr(err)
	vars := []string{}
	for _, swatch := range palette.ExtractAwesome() {
		c := swatch.Color
		r, g, b := c.RGB()
		n := strings.ToLower(swatch.Name)
		vars = append(vars, fmt.Sprintf(`"--%s":"rgba(%d,%d,%d,1)"`, n, r, g, b))
		vars = append(vars, fmt.Sprintf(`"--%s-text":"%s"`, n, c.TitleTextColor()))
	}
	stylesheet := "{" + strings.Join(vars, ",") + "}"
	fmt.Fprintf(w, stylesheet)
	imageCache[path] = &file{modtime: modtime, stylesheet: stylesheet}
}

func fileExists(filename string) bool {
	f, err := os.Open(filename)
	f.Close()
	if os.IsNotExist(err) {
		return false
	}
	checkErr(err)
	return true
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
