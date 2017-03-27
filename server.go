package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"hash/crc32"
	"image"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
		} else if strings.HasPrefix(req, "audio/") && strings.HasSuffix(req, ".zip") {
			if !fileExists(strings.TrimSuffix(req, ".zip")) {
				log.Println("<- 404 Not Found")
				io.WriteString(w, "File Not Found")
				return
			}
			zipHandler(w, r)
			log.Println("<- 200 OK")
			return
		}
		log.Println("<- 200 OK")
		// if strings.HasSuffix(req, ".mp3") {
		// 	<-time.After(time.Second * 2)
		// }
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

var vibrantFallback = map[string]string{
	"vibrant":           "#acaaaa",
	"vibrant-text":      "#000",
	"lightvibrant":      "#fff",
	"lightvibrant-text": "#000",
	"darkvibrant":       "#2b2b2b",
	"darkvibrant-text":  "#fff",
	"muted":             "#6d6a6a",
	"muted-text":        "#fff",
	"lightmuted":        "#6d6a6a",
	"lightmuted-text":   "#fff",
	"darkmuted":         "#32312f",
	"darkmuted-text":    "#fff",
}

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
	vibrantColors := vibrantFallback
	for _, swatch := range palette.ExtractAwesome() {
		c := swatch.Color
		r, g, b := c.RGB()
		n := strings.ToLower(swatch.Name)
		vibrantColors[n] = fmt.Sprintf(`rgba(%d,%d,%d,1)`, r, g, b)
		vibrantColors[n+"-text"] = c.TitleTextColor().RGBHex()
	}
	vars := []string{}
	for k, v := range vibrantColors {
		vars = append(vars, fmt.Sprintf(`"--%s":"%s"`, k, v))
	}
	stylesheet := "{" + strings.Join(vars, ",") + "}"
	fmt.Fprintf(w, stylesheet)
	imageCache[path] = &file{modtime: modtime, stylesheet: stylesheet}
}

func crc32sum(f io.Reader) string {
	b, err := ioutil.ReadAll(f)
	checkErr(err)
	return fmt.Sprintf("%08x", crc32.ChecksumIEEE(b))
}

func zipHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: cache created zips on disk
	// TODO: add license, nfo, m3u, album art
	// TODO: add tracknumbers, artist, etc (from data.json maybe?)
	w.Header().Set("Content-Type", "application/zip")

	path := "." + r.URL.Path
	path = strings.TrimSuffix(path, ".zip")

	w.Header().Set("Content-Disposition", "attachment; filename=\""+path+".zip\"")

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	filepath.Walk(path, func(path string, file os.FileInfo, err error) error {
		if strings.HasSuffix(file.Name(), ".mp3") {
			zf, err := zw.Create(file.Name())
			checkErr(err)
			f, err := os.Open(path)
			checkErr(err)
			io.Copy(zf, f)
			f.Close()
		}
		return nil
	})

	checkErr(zw.Close())
	io.Copy(w, buf)
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
