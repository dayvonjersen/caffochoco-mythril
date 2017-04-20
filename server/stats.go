package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	day   = 86400
	week  = day * 7
	month = week * 4
	year  = month * 12
)

func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	log.Println("<- 200 OK")

	d := getData()

	todayPlays := map[string]int{}
	todayDownloads := map[string]int{}
	weekPlays := map[string]int{}
	weekDownloads := map[string]int{}
	monthPlays := map[string]int{}
	monthDownloads := map[string]int{}
	yearPlays := map[string]int{}
	yearDownloads := map[string]int{}
	allPlays := map[string]int{}
	allDownloads := map[string]int{}

	now := int(time.Now().Unix())

	for _, rel := range d.Releases {
		var d int
		d = counter.Downloads(rel.Url, 0)
		if d != 0 {
			allDownloads[rel.Url] = d
		}
		d = counter.Downloads(rel.Url, now-day)
		if d != 0 {
			todayDownloads[rel.Url] = d
		}
		d = counter.Downloads(rel.Url, now-week)
		if d != 0 {
			weekDownloads[rel.Url] = d
		}
		d = counter.Downloads(rel.Url, now-month)
		if d != 0 {
			monthDownloads[rel.Url] = d
		}
		d = counter.Downloads(rel.Url, now-year)
		if d != 0 {
			yearDownloads[rel.Url] = d
		}
		for _, tl := range rel.Tracklists {
			for _, t := range tl.Tracks {
				file := rel.Url + "/" + t.File
				var p int
				p = counter.Plays("audio/"+file, 0)
				if p != 0 {
					allPlays[file] = p
				}
				p = counter.Plays("audio/"+file, now-day)
				if p != 0 {
					todayPlays[file] = p
				}
				p = counter.Plays("audio/"+file, now-week)
				if p != 0 {
					weekPlays[file] = p
				}
				p = counter.Plays("audio/"+file, now-month)
				if p != 0 {
					monthPlays[file] = p
				}
				p = counter.Plays("audio/"+file, now-year)
				if p != 0 {
					yearPlays[file] = p
				}
			}
		}
	}

	type s struct {
		Plays     map[string]int `json:"plays"`
		Downloads map[string]int `json:"downloads"`
	}
	stats := struct {
		All   s `json:"all"`
		Today s `json:"today"`
		Week  s `json:"week"`
		Month s `json:"month"`
		Year  s `json:"year"`
	}{
		s{allPlays, allDownloads},
		s{todayPlays, todayDownloads},
		s{weekPlays, weekDownloads},
		s{monthPlays, monthDownloads},
		s{yearPlays, yearDownloads},
	}

	b, err := json.MarshalIndent(stats, "", "\t")
	checkErr(err)
	io.WriteString(w, string(b))
}
