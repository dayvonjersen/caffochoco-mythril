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
	"path/filepath"
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

		if req == "test" {
			testHandler(w, r)
			log.Println("<- 418 Teapot")
			return
		}

		if fileExists(req) {
			file = req
		} else if strings.HasPrefix(req, "bower_components/") {
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

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	d := getData()
	re := regexp.MustCompile(`[^\w-.]+`)
	sig := "dayvonjersen"
	for _, rel := range d.Releases {
		tdata := struct {
			Release, Artist, Title, Genre, Encoder, Quality, About, NumTracks, Length, Size string
			numTracks, length, size                                                         int
			HasArt                                                                          bool
			Tracks                                                                          []struct {
				Num, Title, Time string
			}
		}{
			Tracks: []struct {
				Num, Title, Time string
			}{},
			Size: strings.Repeat(" ", 56),
		}
		rname := fmt.Sprintf("%02d %s - %s-%d-%s", 0, rel.Artist, rel.Title, rel.Year, sig)
		tdata.Release = strwrap(fmt.Sprintf("dayvonjersen.com/releases/%s", rel.Url), 56, "║                     ", " ║", true, false)
		tdata.Artist = strpad(rel.Artist, 56)
		tdata.Title = strpad(rel.Title, 56)
		tdata.Genre = strpad(rel.Genre, 56)
		tdata.Encoder = strpad("LAME", 56)
		tdata.Quality = strpad("320kbps MP3", 56)
		tdata.About = strwrap(rel.About, 55, "║           ", "            ║", false, false)
		tdata.HasArt = fileExists("./image/" + rel.Url + ".jpg")
		rname = re.ReplaceAllString(rname, "_")
		fmt.Fprintln(w, rname+".m3u")
		fmt.Fprint(w, renderTemplate("m3u.tmpl", rel.Tracklists[rel.DefaultTracklistId]))
		fmt.Fprintln(w, rname+".nfo")
		fmt.Fprintln(w, rname+".sfv")
		i := 1
		for _, tl := range rel.TracklistIds {
			//fmt.Fprintf(w, "\t%s:\n", rel.Tracklists[tl].Title)
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
				fname := fmt.Sprintf("%02d %s - %s-%s.mp3", i, rel.Artist, track.Title, sig)
				i++
				fname = re.ReplaceAllString(fname, "_")
				fmt.Fprintln(w, fname)
			}
		}
		tdata.NumTracks = strpad(fmt.Sprintf("%d", tdata.numTracks), 56)
		tdata.Length = strpad(formattime(tdata.length), 56)
		fmt.Fprint(w, renderTemplate("awesome-tmpl.txt", tdata))
	}
	fmt.Fprintln(w)
}
