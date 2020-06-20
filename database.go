package main

import (
	"database/sql"
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
    guild_id text not null,
    chan_id text not null,
    msg_id text not null,
    img_id text not null,
    time datetime not null,
    path text not null,
    caption text
);`, `
create unique index index_records_img on records(img_id);
`, `
create index index_records_msg on records(msg_id);
`, `
create index index_records_time on records(time);
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
create table subs (
    chan_id text not null primary key,
    earliest_msg_id text not null,
    latest_msg_id text not null
);`,
}

func check(err error, query string) {
	if err != nil {
		Fatal("unable to query '%v': %v", query, err)
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
		Fatal("unable to open '%v': %v", file, err)
	}

	if init {
		for _, query := range schema {
			if _, err := db.Exec(query); err != nil {
				Fatal("unable to initialize db '%v': %v\n%v", file, err, query)
			}
		}
	}
}

func DatabaseClose() {
	db.Close()
}

func DatabaseSub(channel string) (string, string) {
	lock.RLock()
	defer lock.RUnlock()

	const query = "select earliest_msg_id, latest_msg_id from subs where chan_id = ?;"
	rows, err := db.Query(query, channel)
	defer rows.Close()
	check(err, query)

	if !rows.Next() {
		const query = "insert into subs (chan_id, earliest_msg_id, latest_msg_id) values (?, \"\", \"\");"
		_, err := db.Exec(query, channel)
		check(err, query)
		return "", ""
	}

	var earliest, latest string
	check(rows.Scan(&earliest, &latest), query)
	return earliest, latest
}

func DatabaseSubUpdateLatest(channel, pos string) {
	lock.Lock()
	defer lock.Unlock()

	const query = "update subs set latest_msg_id = ? where chan_id = ?;"
	_, err := db.Exec(query, pos, channel)
	check(err, query)
}

func DatabaseSubUpdateEarliest(channel, pos string) {
	lock.Lock()
	defer lock.Unlock()

	const query = "update subs set earliest_msg_id = ? where chan_id = ?;"
	_, err := db.Exec(query, pos, channel)
	check(err, query)
}

func DatabaseRecordInsert(rec *Record) int64 {
	lock.Lock()
	defer lock.Unlock()

	tx, err := db.Begin()
	check(err, "begin")

	{
		const query = "insert into records(guild_id, chan_id, msg_id, img_id, time, path, caption)" +
			"  values(?, ?, ?, ?, ?, ?, ?) on conflict do nothing;"
		_, err := tx.Exec(
			query,
			rec.GuildId,
			rec.ChannelId,
			rec.MessageId,
			rec.ImageId,
			rec.Time,
			rec.Path,
			rec.Caption)
		check(err, query)
	}

	// result.LastInsertId() is unreliable in the case of an upsert so can't use it.
	var id int64
	{
		const query = "select id from records where img_id = ?;"
		rows, err := tx.Query(query, rec.ImageId)

		if !rows.Next() {
			Fatal("unable to get id of upserted record '%v': %v", rec.ImageId, err)
		}
		check(rows.Scan(&id), query)
	}

	{
		const query = "insert into tags(record_id, tag) values(?, ?) on conflict do nothing;"
		for _, tag := range rec.Tags {
			_, err := tx.Exec(query, id, tag)
			check(err, query)
		}
	}

	check(tx.Commit(), "commit")
	return id
}

func DatabaseRecord(guild string, id int64) *Record {
	lock.RLock()
	defer lock.RUnlock()

	tx, err := db.Begin()
	check(err, "begin")

	rec := &Record{}

	{
		const query = "select id, guild_id, chan_id, msg_id, img_id, time, path, caption" +
			"  from records where id = ? and guild_id = ?;"
		rows, err := tx.Query(query, id, guild)
		defer rows.Close()
		check(err, query)

		if !rows.Next() {
			return nil
		}

		check(rows.Scan(
			&rec.Id,
			&rec.GuildId,
			&rec.ChannelId,
			&rec.MessageId,
			&rec.ImageId,
			&rec.Time,
			&rec.Path,
			&rec.Caption), query)
	}

	{
		const query = "select tag from tags where record_id = ? order by tag asc;"
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

func DatabaseRecords(guild string, limit int64) []int64 {
	lock.RLock()
	defer lock.RUnlock()

	Debug("records guild:%v limit:%v", guild, limit)

	const query = "select id from records where guild_id = ? order by time desc limit ?;"
	rows, err := db.Query(query, guild, limit)
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

func DatabaseTag(guild, tag string) []int64 {
	lock.RLock()
	defer lock.RUnlock()

	const query = "select record_id from tags, records" +
		"  where tag = ? and tags.record_id = records.id and records.guild_id = ?" +
		"  order by records.time desc;"
	rows, err := db.Query(query, tag, guild)
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

func DatabaseTags(guild string) []string {
	lock.RLock()
	defer lock.RUnlock()

	const query = "select distinct tag from tags, records" +
		"  where tags.record_id = records.id and records.guild_id = ?" +
		"  order by tag asc;"
	rows, err := db.Query(query, guild)
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

		rows.Next()
		check(rows.Scan(&stats.RecordsRows), query)
	}

	{
		const query = "select count(*) from tags;"
		rows, err := db.Query(query)
		defer rows.Close()
		check(err, query)

		rows.Next()
		check(rows.Scan(&stats.TagsRows), query)
	}

	{
		const query = "select count(distinct tag) from tags;"
		rows, err := db.Query(query)
		defer rows.Close()
		check(err, query)

		rows.Next()
		check(rows.Scan(&stats.TagsDistinct), query)
	}

	{
		const query = "select tag, count(record_id) from tags group by tag;"
		rows, err := db.Query(query)
		defer rows.Close()
		check(err, query)

		stats.Tags = make(map[string]int64)
		for rows.Next() {
			var tag string
			var count int64
			check(rows.Scan(&tag, &count), query)
			stats.Tags[tag] = count
		}
	}

	return stats
}
