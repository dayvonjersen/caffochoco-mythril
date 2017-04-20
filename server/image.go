package main

import (
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/generaltso/vibrant"

	_ "image/jpeg"
	_ "image/png"
)

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
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	path := "." + r.URL.Path
	path = strings.TrimSuffix(path, ".json")

	if !fileExists(path) {
		notfoundHandler(w, r)
		return
	}
	log.Println("<- 200 OK")

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
