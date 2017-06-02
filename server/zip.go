package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var re = regexp.MustCompile(`[^\w-.\(\)\[\]]+`)

func zipHandler(w http.ResponseWriter, r *http.Request) {

	tracklistId, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/download/"))

	if err != nil {
		log.Printf("strconv.Atoi(%v): %v", strings.TrimPrefix(r.URL.Path, "/download/"), err)
		notfoundHandler(w, r)
		return
	}

	rel, err := getReleaseByTracklist(tracklistId)

	if err != nil {
		log.Printf("getReleaseByTracklist(%v): %v", tracklistId, err)
		notfoundHandler(w, r)
		return
	}

	log.Println(req(r), "<- \033[32m200\033[0m OK")
	w.Header().Set("Content-Type", "application/zip")

	rname := releaseName(tracklistId, rel, true)

	zipFile := rname + ".zip"

	w.Header().Set("Content-Disposition", "attachment; filename=\""+zipFile+"\"")

	if !fileExists(".cache/" + zipFile) {
		f, err := os.Create(".cache/" + zipFile)
		checkErr(err)
		io.Copy(f, createZip(tracklistId, rel))
		f.Close()
	}
	f, err := os.Open(".cache/" + zipFile)
	checkErr(err)
	defer f.Close()
	io.Copy(w, f)
}

func releaseName(tracklistId int, rel release, prefix bool) string {
	special := ""
	tracklistTitle := rel.Tracklists[tracklistId].Title
	if tracklistTitle != "tracklist" {
		special = "(" + tracklistTitle + ") "
	}
	if !prefix {
		tracklistId = 0
	}
	rname := fmt.Sprintf("%02d %s - %s %s%d[music.dayvonjersen.com]", tracklistId, rel.Artist, rel.Title, special, rel.Year)
	rname = strings.ToLower(rname)
	rname = re.ReplaceAllString(rname, "_")
	return rname
}

func zipPrecache() {
	d := getData()
	for _, rel := range d.Releases {
		for _, tracklistId := range rel.TracklistIds {
			rname := releaseName(tracklistId, rel, true)
			zipFile := ".cache/" + rname + ".zip"
			log.Println("generating", zipFile)
			if fileExists(zipFile) {
				log.Println("removing existing", zipFile)
				checkErr(os.Remove(zipFile))
			}
			f, err := os.Create(zipFile)
			checkErr(err)
			n, _ := io.Copy(f, createZip(tracklistId, rel))
			log.Println(n, "bytes written [ OK ]")
			f.Close()
		}
	}
}

func createZipHeader(name string) *zip.FileHeader {
	h := &zip.FileHeader{
		Name:   name,
		Method: zip.Store,
	}
	h.SetModTime(time.Now())
	return h
}

func crc32sum(filename string) string {
	f, err := os.Open(filename)
	checkErr(err)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	checkErr(err)
	return fmt.Sprintf("%08x", crc32.ChecksumIEEE(b))
}

func createZip(tracklistId int, rel release) io.Reader {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	rname := releaseName(tracklistId, rel, false)

	nfo, err := zw.CreateHeader(createZipHeader(rname + ".nfo"))
	checkErr(err)
	io.WriteString(nfo, createNfo(tracklistId, rel))

	path := "./audio/" + rel.Url

	sfvText := ""
	m3uText := "#EXTM3U\n"

	i := 1
	for _, t := range rel.Tracklists[tracklistId].TrackIds {
		track := rel.Tracklists[tracklistId].Tracks[t]

		fname := fmt.Sprintf("%02d %s - %s.mp3", i, rel.Artist, track.Title)
		fname = strings.ToLower(fname)
		fname = re.ReplaceAllString(fname, "_")
		i++

		sfvText += fmt.Sprintf("%s\t%s\n", fname, crc32sum(path+"/"+track.File))

		m3uText += fmt.Sprintf("#EXTINF:%d,%s\n%s\n", track.Length, track.Title, fname)

		zf, err := zw.CreateHeader(createZipHeader(fname))
		checkErr(err)
		f, err := os.Open(path + "/" + track.File)
		checkErr(err)
		io.Copy(zf, f)
		f.Close()
	}

	if fileExists("./image/" + rel.Url + ".jpg") {
		zf, err := zw.CreateHeader(createZipHeader("AlbumArt.jpg"))
		checkErr(err)
		f, err := os.Open("./image/" + rel.Url + ".jpg")
		checkErr(err)
		io.Copy(zf, f)
		f.Close()
	}

	sfv, err := zw.CreateHeader(createZipHeader(rname + ".sfv"))
	checkErr(err)
	io.WriteString(sfv, sfvText)

	m3u, err := zw.CreateHeader(createZipHeader(rname + ".m3u"))
	checkErr(err)
	io.WriteString(m3u, m3uText)

	checkErr(zw.Close())
	return buf
}
