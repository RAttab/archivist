package main

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Stats struct {
	RecordsRows  int64            `json:"records_rows"`
	TagsRows     int64            `json:"tags_rows"`
	TagsDistinct int64            `json:"tags_distinct"`
	Tags         map[string]int64 `json:"tags"`
}

type Record struct {
	Id        int64     `json:"id"`
	GuildId   string    `json:"guild"`
	ChannelId string    `json:"channel"`
	MessageId string    `json:"message"`
	ImageId   string    `json:"image"`
	Time      time.Time `json:"time"`
	Path      string    `json:"path"`
	Caption   string    `json:"caption"`
	Tags      []string  `json:"tags"`
}

// /api/query/<guild>?from=<[msg|img]:<id>>&limit=<int>&tags=<tag,tag,tag>

type QueryId struct {
	Column string
	Id     string
}

func (id QueryId) IsSet() bool {
	return id.Column != ""
}

type Query struct {
	GuildId string
	Tag     string
	From    QueryId
	Since   QueryId
	Limit   int64
}

func (Query) parseGuild(path string) (bool, string) {
	items := strings.Split(path, "/")
	if len(items) != 4 || items[2] != "query" {
		return false, ""
	}

	guild := items[3]
	if _, err := strconv.ParseInt(guild, 10, 64); err != nil {
		return false, ""
	}

	return true, guild
}

func (Query) parseId(raw string) (bool, QueryId) {
	if raw == "" {
		return true, QueryId{}
	}

	items := strings.Split(raw, ":")
	if len(items) != 2 {
		return false, QueryId{}
	}

	var column string
	switch items[0] {
	case "msg":
		column = "msg_id"
	case "img":
		column = "img_id"
	case "chan":
		column = "chan_id"
	default:
		return false, QueryId{}
	}

	id := items[1]
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		return false, QueryId{}
	}

	return true, QueryId{column, id}
}

func (Query) parseInt(raw string) (bool, int64) {
	if raw == "" {
		return true, -1
	} else if val, err := strconv.ParseInt(raw, 10, 64); err != nil {
		return false, -1
	} else {
		return true, val
	}
}

func (Query) parseTag(tag string) (bool, string) {
	if tag == "" {
		return true, ""
	} else if n := len(tag); n == 0 || n > 50 {
		return false, ""
	}
	return true, tag
}

func NewQuery(url *url.URL) (bool, Query) {
	query := Query{}
	qs := url.Query()

	if ok, guild := query.parseGuild(url.Path); ok {
		query.GuildId = guild
	} else {
		return false, Query{}
	}

	if ok, id := query.parseId(qs.Get("from")); ok {
		query.From = id
	} else {
		return false, Query{}
	}

	if ok, id := query.parseId(qs.Get("since")); ok {
		query.From = id
	} else {
		return false, Query{}
	}

	if ok, limit := query.parseInt(qs.Get("limit")); ok && limit > 0 {
		query.Limit = limit
	} else {
		query.Limit = 20
	}

	if ok, tag := query.parseTag(qs.Get("tag")); ok {
		query.Tag = tag
	} else {
		return false, Query{}
	}

	return true, query
}

func (query Query) HasFrom() bool {
	return query.From.IsSet()
}
func (query Query) HasSince() bool {
	return query.From.IsSet()
}

func (query Query) HasTag() bool {
	return query.Tag != ""
}
