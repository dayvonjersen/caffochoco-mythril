package main

import "os"

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

func fileSize(filename string) int {
	f, err := os.Open(filename)
	checkErr(err)
	defer f.Close()
	finfo, err := f.Stat()
	checkErr(err)
	return int(finfo.Size())
}
