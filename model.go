package main

import (
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
