/*
	this was just supposed to be a "simple" server
	to use instead of polymer serve

	fml
*/
package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"hash/crc32"
	"image"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"
	"unicode/utf8"

	_ "image/jpeg"
	_ "image/png"

	"github.com/generaltso/vibrant"

	"./strip"
)

var counter *Counter

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

	counter = NewCounter(".cache/caffo.db")
	defer counter.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("->", r.Method, r.URL)
		file := "index.html"
		req := strings.TrimPrefix(r.URL.Path, "/")

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		if fileExists(req) && !isDir(req) {
			file = req
		} else if strings.HasPrefix(req, "bower_components") {
			notfoundHandler(w, r)
			return
		} else if strings.HasPrefix(req, "image/") && strings.HasSuffix(req, ".css") {
			if !fileExists(strings.TrimSuffix(req, ".css")) {
				notfoundHandler(w, r)
				return
			}
			imageHandler(w, r)
			log.Println("<- 200 OK")
			return
		} else if strings.HasPrefix(req, "audio/") && strings.HasSuffix(req, ".zip") {
			if !fileExists(strings.TrimSuffix(req, ".zip")) {
				notfoundHandler(w, r)
				return
			}
			zipHandler(w, r)
			log.Println("<- 200 OK")
			return
		} else if strings.HasPrefix(req, "plays/") {
			file := strings.TrimPrefix(req, "plays/")
			fmt.Fprintf(w, "%s - %d play(s)", file, counter.Plays(file))
			log.Println("<- 200 OK")
			return
		}
		log.Println("<- 200 OK")
		if strings.HasSuffix(req, ".mp3") {
			counter.Increment(req, r.RemoteAddr)
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
	io.WriteString(w, "File Not Found")
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

func crc32sum(filename string) string {
	f, err := os.Open(filename)
	checkErr(err)
	b, err := ioutil.ReadAll(f)
	checkErr(err)
	return fmt.Sprintf("%08x", crc32.ChecksumIEEE(b))
}

var re = regexp.MustCompile(`[^\w-.]+`)

func zipHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/zip")

	path := "." + r.URL.Path
	path = strings.TrimSuffix(path, ".zip")

	rel := getReleaseByURL(strings.TrimPrefix(path, "./audio/"))

	rname := fmt.Sprintf("00 %s - %s-%d", rel.Artist, rel.Title, rel.Year)
	rname = re.ReplaceAllString(rname, "_")

	zipFile := rname + ".zip"

	w.Header().Set("Content-Disposition", "attachment; filename=\""+zipFile+"\"")

	if !fileExists(".cache/" + zipFile) {
		f, err := os.Create(".cache/" + zipFile)
		checkErr(err)
		io.Copy(f, createZip(path, rname, rel))
		f.Close()
	}
	f, err := os.Open(".cache/" + zipFile)
	checkErr(err)
	defer f.Close()
	io.Copy(w, f)
}

func createZip(path, rname string, rel release) io.Reader {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	nfo, err := zw.Create(rname + ".nfo")
	checkErr(err)
	io.WriteString(nfo, createNfo(rel))

	sfv, err := zw.Create(rname + ".sfv")
	checkErr(err)

	m3u, err := zw.Create(rname + ".m3u")
	checkErr(err)

	io.WriteString(m3u, "#EXTM3U\n")
	i := 1
	for _, tl := range rel.TracklistIds {
		for _, t := range rel.Tracklists[tl].TrackIds {
			track := rel.Tracklists[tl].Tracks[t]
			fname := fmt.Sprintf("%02d %s - %s.mp3", i, rel.Artist, track.Title)
			fname = re.ReplaceAllString(fname, "_")
			i++

			io.WriteString(sfv, fmt.Sprintf("%s\t%s\n", fname, crc32sum(path+"/"+track.File)))

			fmt.Fprintf(m3u, "#EXTINF:%d,%s\n%s\n", track.Length, track.Title, fname)

			zf, err := zw.Create(fname)
			checkErr(err)
			f, err := os.Open(path + "/" + track.File)
			checkErr(err)
			io.Copy(zf, f)
			f.Close()

			track.File = fname
			rel.Tracklists[tl].Tracks[t] = track
		}
	}

	if fileExists("./image/" + rel.Url + ".jpg") {
		zf, err := zw.Create("AlbumArt.jpg")
		checkErr(err)
		f, err := os.Open("./image/" + rel.Url + ".jpg")
		checkErr(err)
		io.Copy(zf, f)
		f.Close()
	}

	checkErr(zw.Close())
	return buf
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

func isDir(filename string) bool {
	finfo, err := os.Stat(filename)
	checkErr(err)
	return finfo.IsDir()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

type track struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	File   string `json:"file"`
	Length int    `json:"length"`
}

type tracklist struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	TrackIds []int  `json:"tracks"`
	Tracks   map[int]track
}

type release struct {
	Id                 int    `json:"id"`
	Artist             string `json:"artist"`
	Title              string `json:"title"`
	Year               int    `json:"year"`
	Genre              string `json:"genre"`
	Url                string `json:"url"`
	TracklistIds       []int  `json:"tracklists"`
	DefaultTracklistId int    `json:"defaultTracklist"`
	Category           string `json:"category"`
	About              string `json:"about"`
	Tracklists         map[int]tracklist
}

type data struct {
	Releases   []release   `json:"releases"`
	Tracklists []tracklist `json:"tracklists"`
	Tracks     []track     `json:"tracks"`
}

func getData() data {
	f, err := os.Open("data.json")
	checkErr(err)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	checkErr(err)

	var d data
	checkErr(json.Unmarshal(b, &d))

	tracklists := map[int]tracklist{}
	tracks := map[int]track{}

	for _, t := range d.Tracks {
		tracks[t.Id] = t
	}
	for i := range d.Tracklists {
		d.Tracklists[i].Tracks = map[int]track{}
	}
	for _, tl := range d.Tracklists {
		for _, t := range tl.TrackIds {
			tl.Tracks[t] = tracks[t]
		}
		tracklists[tl.Id] = tl
	}
	for i := range d.Releases {
		d.Releases[i].Tracklists = map[int]tracklist{}
	}
	for _, r := range d.Releases {
		for _, tl := range r.TracklistIds {
			r.Tracklists[tl] = tracklists[tl]
		}
	}
	return d
}

func getReleaseByURL(url string) release {
	data := getData()
	for _, rel := range data.Releases {
		if rel.Url == url {
			return rel
		}
	}
	return release{}
}

func strpad(s string, l int) string {
	amt := l - utf8.RuneCountInString(s)
	if amt > 0 {
		return s + strings.Repeat(" ", amt)
	}
	return s
}

func strwrap(s string, l int, prefix, postfix string, noprefixfirst, nopostfixfirst bool) string {
	s = strip.StripTags(s)
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return prefix + strpad("[NO TEXT]", l) + postfix
	}
	parts := []string{}
	sect := 0
	j := 0
	for i := 0; i < utf8.RuneCountInString(s); i++ {
		if (i < utf8.RuneCountInString(s)-1 && s[i] == '\n') || (j > 0 && j%l == 0) || i == utf8.RuneCountInString(s)-1 {
			part := s[sect:i]
			if i == utf8.RuneCountInString(s)-1 {
				part = s[sect:]
			}
			part = strings.TrimSpace(part)
			part = strpad(part, l)
			if part[utf8.RuneCountInString(part)-1] != ' ' {
				for s[i-1] != ' ' && s[i-1] != '\n' {
					i--
				}
				part = s[sect:i]
				part = strings.TrimSpace(part)
				part = strpad(part, l)
			}
			sect = i
			if !noprefixfirst || len(parts) > 0 {
				part = prefix + part
			}
			if !nopostfixfirst || len(parts) > 0 {
				part += postfix
			}
			parts = append(parts, part)
			j = 0
		} else {
			j++
		}
	}
	return strings.Join(parts, "\n")
}

func formattime(t int) string {
	m := t / 60
	s := t % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func renderTemplate(filename string, data interface{}) string {
	t, err := template.ParseFiles(filename)
	checkErr(err)
	buf := new(bytes.Buffer)
	checkErr(t.Execute(buf, data))
	return buf.String()
}

func createNfo(rel release) string {
	tdata := struct {
		Release, Artist, Title, Genre, Encoder,
		Quality, About, NumTracks, Length, Size string
		numTracks, length, size int
		HasArt                  bool
		Tracks                  []struct {
			Num, Title, Time string
		}
	}{
		Tracks: []struct {
			Num, Title, Time string
		}{},
		Size: strings.Repeat(" ", 56),
	}
	tdata.Release = strwrap(fmt.Sprintf("dayvonjersen.com/releases/%s", rel.Url), 56, "║                     ", " ║", true, false)
	tdata.Artist = strpad(rel.Artist, 56)
	tdata.Title = strpad(rel.Title, 56)
	tdata.Genre = strpad(rel.Genre, 56)
	tdata.Encoder = strpad("LAME", 56)
	tdata.Quality = strpad("320kbps MP3", 56)
	tdata.About = strwrap(rel.About, 55, "║           ", "            ║", false, false)
	tdata.HasArt = fileExists("./image/" + rel.Url + ".jpg")
	i := 1
	for _, tl := range rel.TracklistIds {
		for _, t := range rel.Tracklists[tl].TrackIds {
			track := rel.Tracklists[tl].Tracks[t]
			tdata.numTracks++
			tdata.length += track.Length

			title := track.Title
			if len(title) > 44 {
				title = title[:44] + " " + formattime(track.Length) + " ║           ║\n" + strwrap(title[44:], 45, "║          ║    ",
					"      ║           ║", false, false)
			} else {
				title = strpad(title, 45) + formattime(track.Length) + " ║           ║"
			}

			tdata.Tracks = append(tdata.Tracks, struct {
				Num, Title, Time string
			}{
				fmt.Sprintf("%02d", i),
				title,
				formattime(track.Length),
			})
		}
	}
	tdata.NumTracks = strpad(fmt.Sprintf("%d", tdata.numTracks), 56)
	tdata.Length = strpad(formattime(tdata.length), 56)
	return renderTemplate("nfo.tmpl", tdata)
}
