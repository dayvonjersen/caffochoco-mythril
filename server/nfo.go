package main

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/dayvonjersen/caffochoco-mythril/server/strip"
)

func strpad(s string, l int) string {
	amt := l - utf8.RuneCountInString(s)
	if amt > 0 {
		return s + strings.Repeat(" ", amt)
	}
	return s
}

func strcenter(s string, max int) string {
	amt := max - utf8.RuneCountInString(s)
	var left, right int
	if amt%2 == 0 {
		left = amt / 2
		right = left
	} else {
		amt := float64(amt) / 2
		left = int(math.Floor(amt))
		right = int(math.Ceil(amt))
	}
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
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

func formatTime(t int) string {
	m := t / 60
	s := t % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

var byteUnits = [4]string{"", "K", "M", "G"}

func formatBytes(bytes int) string {
	b := float64(bytes)
	u := 0
	for b >= 1024 {
		u++
		b /= 1024
	}
	return fmt.Sprintf("%.1f %sB", b, byteUnits[u])
}

func renderTemplate(filename string, data interface{}) string {
	t, err := template.ParseFiles(filename)
	checkErr(err)
	buf := new(bytes.Buffer)
	checkErr(t.Execute(buf, data))
	return buf.String()
}

var htmlentitiesre = regexp.MustCompile(`&.*?;`)

func createNfo(tracklistId int, rel release) string {
	tdata := struct {
		Url,
		Release, Artist, Title, Genre, Encoder,
		Quality, About, NumTracks, Length, Size string
		numTracks, length, size int
		HasArt                  bool
		TracklistTitle          string
		Tracks                  []struct {
			Num, Title, Time string
		}
	}{
		Tracks: []struct {
			Num, Title, Time string
		}{},
	}
	tdata.TracklistTitle = strcenter(strings.ToUpper(rel.Tracklists[tracklistId].Title), 78)
	tdata.Url = strcenter("https://music.dayvonjersen.com/release/"+rel.Url, 78)
	tdata.Release = strwrap(fmt.Sprintf("%02d", tracklistId), 56, "║                     ", " ║", true, false)
	tdata.Artist = strpad(rel.Artist, 56)
	tdata.Title = strpad(rel.Title, 56)
	tdata.Genre = strpad(rel.Genre, 56)
	tdata.Encoder = strpad("LAME", 56)
	tdata.Quality = strpad("320kbps MP3", 56)
	tdata.About = strwrap(htmlentitiesre.ReplaceAllString(rel.About, ""), 55, "║           ", "            ║", false, false)
	tdata.HasArt = fileExists("./image/" + rel.Url + ".jpg")
	i := 1
	size := 0
	for _, t := range rel.Tracklists[tracklistId].TrackIds {
		track := rel.Tracklists[tracklistId].Tracks[t]
		tdata.numTracks++
		tdata.length += track.Length

		title := track.Title
		if len(title) > 44 {
			title = title[:44] + " " + formatTime(track.Length) + " ║           ║\n" +
				strwrap(title[44:], 45, "║          ║    ", "      ║           ║", false, false)
		} else {
			title = strpad(title, 45) + formatTime(track.Length) + " ║           ║"
		}

		tdata.Tracks = append(tdata.Tracks, struct {
			Num, Title, Time string
		}{
			fmt.Sprintf("%02d", i),
			title,
			formatTime(track.Length),
		})
		i++
		size += fileSize("./audio/" + rel.Url + "/" + track.File)
	}
	tdata.Size = strpad(formatBytes(size), 56)
	tdata.NumTracks = strpad(fmt.Sprintf("%d", tdata.numTracks), 56)
	tdata.Length = strpad(formatTime(tdata.length), 56)
	return renderTemplate("./server/nfo.tmpl", tdata)
}
