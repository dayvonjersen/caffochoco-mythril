package main

import (
	"./strip"
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"unicode/utf8"
)

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

func createNfo(tracklistId int, rel release) string {
	tdata := struct {
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
		Size: strings.Repeat(" ", 56),
	}
	tdata.TracklistTitle = strpad(strings.ToUpper(rel.Tracklists[tracklistId].Title), 45)
	tdata.Release = strwrap(fmt.Sprintf("%03d dayvonjersen.com/releases/%s", tracklistId, rel.Url), 56, "║                     ", " ║", true, false)
	tdata.Artist = strpad(rel.Artist, 56)
	tdata.Title = strpad(rel.Title, 56)
	tdata.Genre = strpad(rel.Genre, 56)
	tdata.Encoder = strpad("LAME", 56)
	tdata.Quality = strpad("320kbps MP3", 56)
	tdata.About = strwrap(rel.About, 55, "║           ", "            ║", false, false)
	tdata.HasArt = fileExists("./image/" + rel.Url + ".jpg")
	i := 1
	for _, t := range rel.Tracklists[tracklistId].TrackIds {
		track := rel.Tracklists[tracklistId].Tracks[t]
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
	tdata.NumTracks = strpad(fmt.Sprintf("%d", tdata.numTracks), 56)
	tdata.Length = strpad(formattime(tdata.length), 56)
	return renderTemplate("nfo.tmpl", tdata)
}
