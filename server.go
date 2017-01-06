package main

import (
	"net/http"
	"os"
	"strings"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		file := "index.html"
		if req := strings.TrimPrefix(r.URL.Path, "/"); fileExists(req) {
			file = req
		}
		http.ServeFile(w, r, file)
	})
	println("Listening on :8000...")
	println(http.ListenAndServe(":8000", nil))
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
