package main

import (
	"log"
	"strings"
	"time"
)

const sleep = 1 * time.Second

var pipe chan chan bool = make(chan chan bool)

func ScrapperStart(channel, since string) {
	go scrape(channel, since)
}

func ScrapperStop() {
	wait := make(chan bool)
	pipe <- wait
	<-wait
}

func bail() bool {
	select {
	case done := <-pipe:
		done <- true
		return true
	default:
		return false
	}
}

func validAttachment(attach *Attachment) bool {
	file := attach.Filename
	return strings.HasSuffix(file, ".jpg") ||
		strings.HasSuffix(file, ".png") ||
		strings.HasSuffix(file, ".gif")
}

func message(msg *Message) error {
	log.Printf("DBG: message id=%v, id=%v, content=%v", msg.ID, msg.Author.Username, msg.Content)

	ts, err := DiscordTimestamp(msg.ID)
	if err != nil {
		return err
	}

	for _, attach := range msg.Attachments {
		log.Printf("DBG:   attach %v", attach)

		if !validAttachment((*Attachment)(attach)) {
			continue
		}

		art := &Record{
			MessageId: msg.ID,
			ImageId:   attach.ID,
			Time:      ts,
			Path:      attach.URL,
			Caption:   msg.Content,
			Tags:      []string{TagsAuthor(msg.Author.Username)},
		}

		id := DatabaseRecordInsert(art)
		log.Printf("PUT: %v -> %v", msg.ID, id)
	}

	return nil
}

func scrape(channel, since string) {
	it, err := DiscordMessageIterator(channel, since)
	if err != nil {
		log.Fatalf("unable to create iterator for '%v': %v", channel, it)
	}

	for true {
		if bail() {
			break
		}

		msg, err := it.DiscordIteratorNext()
		if err != nil {
			log.Fatalf("unable to read message for '%v': %v", channel, err)
		}
		if msg == nil {
			time.Sleep(sleep)
			continue
		}

		if err := message(msg); err != nil {
			log.Printf("ERROR: failed to parse '%v': %v", msg.ID, err)
		}

		DatabaseItSet(channel, msg.ID)
	}
}
