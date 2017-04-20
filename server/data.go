package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

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

func getReleaseByTracklist(tracklistId int) (release, error) {
	data := getData()
	for _, rel := range data.Releases {
		for _, id := range rel.TracklistIds {
			if id == tracklistId {
				return rel, nil
			}
		}
	}
	return release{}, fmt.Errorf("tracklist id %d does not exist", tracklistId)
}
