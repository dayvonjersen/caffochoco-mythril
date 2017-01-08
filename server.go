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

func imageHandler(w http.ResponseWriter, r *http.Request) {
	path := "." + r.URL.Path
	path = strings.TrimSuffix(path, ".css")

	f, err := os.Open(path)
	checkErr(err)
	img, _, err := image.Decode(f)
	f.Close()
	checkErr(err)
	palette, err := vibrant.NewPaletteFromImage(img)
	checkErr(err)
	w.Header().Set("Content-Type", "text/css; charset=UTF-8")
	fmt.Fprintf(w, ":root {\n")
	for _, swatch := range palette.ExtractAwesome() {
		c := swatch.Color
		r, g, b := c.RGB()
		n := strings.ToLower(swatch.Name)
		fmt.Fprintf(w,
			"    --%s: rgba(%d,%d,%d,1);\n    --%s-text: %s;\n",
			n, r, g, b, n, c.TitleTextColor(),
		)
	}
	fmt.Fprintf(w, "}")
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
