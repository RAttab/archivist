package main

import (
	"database/sql"
	"log"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db   *sql.DB
	lock sync.RWMutex
)

var schema = []string{`
create table records (
    id integer not null primary key autoincrement,
    msg_id text not null,
    img_id text not null,
    time datetime not null,
    path text not null,
    caption text
);`, `
create unique index index_records_img on records(img_id);
`, `
create table tags(
    record_id integer not null references records(id),
    tag text not null,
    constraint pk primary key (record_id, tag)
);`, `
create index index_tags_record on tags(record_id);
`, `
create index index_tags_tag on tags(tag);
`, `
create table iterators (
    chan_id text not null primary key,
    msg_id text not null
);`,
}

func check(err error, query string) {
	if err != nil {
		log.Fatalf("unable to query '%v': %v", query, err)
	}
}

func DatabaseOpen(file string) {
	init := false
	if _, err := os.Stat(file); err != nil {
		init = true
	}

	var err error
	db, err = sql.Open("sqlite3", "file:"+file+"?mode=rwc")
	if err != nil {
		log.Fatalf("unable to open '%v': %v", file, err)
	}

	if init {
		for _, query := range schema {
			if _, err := db.Exec(query); err != nil {
				log.Fatalf("unable to initialize db '%v': %v\n%v", file, err, query)
			}
		}
	}
}

func DatabaseClose() {
	db.Close()
}

func DatabaseIt(channel string) string {
	lock.RLock()
	defer lock.RUnlock()

	const query = "select msg_id from iterators where chan_id = ?;"
	rows, err := db.Query(query, channel)
	defer rows.Close()
	check(err, query)

	if !rows.Next() {
		return ""
	}

	var msg string
	check(rows.Scan(&msg), query)

	return msg
}

func DatabaseItSet(channel, pos string) {
	lock.Lock()
	defer lock.Unlock()

	const query = "insert into iterators (chan_id, msg_id) values(?, ?)" +
		"  on conflict(chan_id) do update set msg_id = ?;"

	_, err := db.Exec(query, channel, pos, pos)
	check(err, query)
}

func DatabaseRecordInsert(rec *Record) int64 {
	lock.Lock()
	defer lock.Unlock()

	tx, err := db.Begin()
	check(err, "begin")

	const queryRecs = "insert into records(msg_id, img_id, time, path, caption) values(?, ?, ?, ?, ?);"
	result, err := tx.Exec(queryRecs, rec.MessageId, rec.ImageId, rec.Time, rec.Path, rec.Caption)
	check(err, queryRecs)

	id, err := result.LastInsertId()
	check(err, "last insert id")

	const queryTag = "insert into tags(record_id, tag) values(?, ?);"
	for _, tag := range rec.Tags {
		_, err := tx.Exec(queryTag, id, tag)
		check(err, queryTag)
	}

	check(tx.Commit(), "commit")
	return id
}

func DatabaseRecord(id int64) *Record {
	lock.RLock()
	defer lock.RUnlock()

	tx, err := db.Begin()
	check(err, "begin")

	rec := &Record{}

	{
		const query = "select id, msg_id, img_id, time, path, caption from records where id = ?;"
		rows, err := tx.Query(query, id)
		defer rows.Close()
		check(err, query)

		if !rows.Next() {
			return nil
		}

		check(rows.Scan(&rec.Id, &rec.MessageId, &rec.ImageId, &rec.Time, &rec.Path, &rec.Caption), query)
	}

	{
		const query = "select tag from tags where record_id = ?;"
		rows, err := tx.Query(query, rec.Id)
		defer rows.Close()
		check(err, query)

		for rows.Next() {
			var tag string
			check(rows.Scan(&tag), query)
			rec.Tags = append(rec.Tags, tag)
		}
	}

	check(tx.Commit(), "commit")
	return rec
}

func DatabaseRecords(limit int64) []int64 {
	lock.RLock()
	defer lock.RUnlock()

	const query = "select id from records order by id desc limit ?;"
	rows, err := db.Query(query, limit)
	defer rows.Close()
	check(err, query)

	ids := []int64{}
	for rows.Next() {
		var id int64
		check(rows.Scan(&id), query)
		ids = append(ids, id)
	}

	return ids
}

func DatabaseTag(tag string) []int64 {
	lock.RLock()
	defer lock.RUnlock()

	const query = "select record_id from tags where tag = ?;"
	rows, err := db.Query(query, tag)
	defer rows.Close()
	check(err, query)

	ids := []int64{}
	for rows.Next() {
		var id int64
		check(rows.Scan(&id), query)
		ids = append(ids, id)
	}

	return ids
}

func DatabaseTags() []string {
	lock.RLock()
	defer lock.RUnlock()

	const query = "select distinct tag from tags;"
	rows, err := db.Query(query)
	defer rows.Close()
	check(err, query)

	tags := []string{}
	for rows.Next() {
		var tag string
		check(rows.Scan(&tag), query)
		tags = append(tags, tag)
	}

	return tags
}

func DatabaseTagsSet(id int64, tags []string) {
	lock.Lock()
	defer lock.Unlock()

	tx, err := db.Begin()
	check(err, "begin")

	for _, tag := range tags {
		const query = "insert into tags(record_id, tag) values(?, ?);"
		_, err := db.Exec(query, id, tag)
		check(err, query)
	}

	check(tx.Commit(), "commit")
}

func DatabaseTagsUnset(id int64, tags []string) {
	lock.Lock()
	defer lock.Unlock()

	tx, err := db.Begin()
	check(err, "begin")

	for _, tag := range tags {
		const query = "delete from tags where record_id = ? and tag = ?;"
		_, err := db.Exec(query, id, tag)
		check(err, query)
	}

	check(tx.Commit(), "commit")
}

func DatabaseStats() *Stats {
	lock.RLock()
	defer lock.RUnlock()

	stats := &Stats{}

	{
		const query = "select count(id) from records;"
		rows, err := db.Query(query)
		defer rows.Close()
		check(err, query)

		for rows.Next() {
			check(rows.Scan(&stats.RecordsRows), query)
		}
	}

	{
		const query = "select count(*) from tags;"
		rows, err := db.Query(query)
		defer rows.Close()
		check(err, query)

		for rows.Next() {
			check(rows.Scan(&stats.TagsRows), query)
		}
	}

	{
		const query = "select count(distinct tag) from tags;"
		rows, err := db.Query(query)
		defer rows.Close()
		check(err, query)

		for rows.Next() {
			check(rows.Scan(&stats.TagsDistinct), query)
		}
	}

	return stats
}
