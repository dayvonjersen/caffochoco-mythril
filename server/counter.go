package main

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Counter struct {
	db *sql.DB
}

func NewCounter(dbname string) *Counter {
	c := &Counter{}
	var err error
	c.db, err = sql.Open("sqlite3", dbname)
	checkErr(err)
	return c
}

func (c *Counter) Close() {
	c.db.Close()
}

func (c *Counter) IncrementPlays(file, ip string) {
	stmt, err := c.db.Prepare("INSERT INTO `plays` (`file`, `ip`, `time`) VALUES (?, ?, ?)")
	checkErr(err)
	defer stmt.Close()

	_, err = stmt.Exec(file, ip, time.Now().Unix())
	checkErr(err)
}

func (c *Counter) Plays(file string, since int) int {
	rows, err := c.db.Query("SELECT COUNT(*) FROM `plays` WHERE `file` = ? AND `time` >= ?", file, since)
	checkErr(err)
	defer rows.Close()

	var p int
	rows.Next()
	rows.Scan(&p)
	return p
}
func (c *Counter) IncrementDownloads(file, ip string) {
	stmt, err := c.db.Prepare("INSERT INTO `downloads` (`file`, `ip`, `time`) VALUES (?, ?, ?)")
	checkErr(err)
	defer stmt.Close()

	_, err = stmt.Exec(file, ip, time.Now().Unix())
	checkErr(err)
}

func (c *Counter) Downloads(file string, since int) int {
	rows, err := c.db.Query("SELECT COUNT(*) FROM `downloads` WHERE `file` = ? AND `time` >= ?", file, since)
	checkErr(err)
	defer rows.Close()

	var p int
	rows.Next()
	rows.Scan(&p)
	return p
}
