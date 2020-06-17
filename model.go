package main

import (
	"time"
)

type Stats struct {
	RecordsRows  int64 `json:"records_rows"`
	TagsRows     int64 `json:"tags_rowss"`
	TagsDistinct int64 `json:"tag_distinct"`
}

type Record struct {
	Id        int64     `json:"id"`
	MessageId string    `json:"msg"`
	ImageId   string    `json:"img"`
	Time      time.Time `json:"time"`
	Path      string    `json:"path"`
	Caption   string    `json:"caption"`
	Tags      []string  `json:"tags"`
}

func TagsAuthor(author string) string {
	return "author: " + author
}
