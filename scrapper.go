package main

import (
	"strings"
	"time"
)

const sleep = 1 * time.Second

var pipe chan chan bool = make(chan chan bool)

func ScrapperStart() {
	for guild, channels := range Config.Discord.Subs {
		for _, channel := range channels {
			earliest, latest := DatabaseSub(channel)
			go scrapeForward(guild, channel, latest)
			go scrapeBackwards(guild, channel, earliest)
		}
	}
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

func message(guild string, msg *Message) (bool, error) {
	kept := false

	ts, err := DiscordTimestamp(msg.ID)
	if err != nil {
		return kept, err
	}

	for _, attach := range msg.Attachments {
		if !validAttachment((*Attachment)(attach)) {
			continue
		}

		if msg.GuildID != "" {
			guild = msg.GuildID
		}

		rec := &Record{
			GuildId:   guild,
			ChannelId: msg.ChannelID,
			MessageId: msg.ID,
			ImageId:   attach.ID,
			Time:      ts,
			Path:      attach.URL,
			Caption:   msg.Content,
			Tags:      []string{TagsAuthor(msg.Author.Username)},
		}

		id := DatabaseRecordInsert(rec)
		Info("PUT: %v -> guild=%v, chan=%v, msg=%v",
			id, rec.GuildId, rec.ChannelId, rec.MessageId)

		kept = true
	}

	return kept, nil
}

func scrapeForward(guild, channel, from string) {
	it, err := DiscordMessageForwardIt(channel, from)
	if err != nil {
		Fatal("unable to create iterator for '%v': %v", channel, it)
	}

	for true {
		if bail() {
			break
		}

		msg, err := it.DiscordItNext()
		if err != nil {
			Fatal("unable to read message for '%v': %v", channel, err)
		}
		if msg == nil {
			time.Sleep(sleep)
			continue
		}

		if _, err := message(guild, msg); err != nil {
			Fatal("ERROR: failed to parse '%v': %v", msg.ID, err)
		}

		DatabaseSubUpdateLatest(channel, msg.ID)
	}
}

func scrapeBackwards(guild, channel, from string) {
	it, err := DiscordMessageBackwardsIt(channel, from)
	if err != nil {
		Fatal("unable to create iterator for '%v': %v", channel, it)
	}

	count := 0
	for true {
		if bail() {
			break
		}

		msg, err := it.DiscordItNext()
		if msg == nil {
			break
		} else if err != nil {
			Fatal("unable to read message for '%v': %v", channel, err)
			return
		}

		if kept, err := message(guild, msg); err != nil {
			Warning("failed to parse '%v': %v", msg.ID, err)
		} else if kept {
			count++
		}

		DatabaseSubUpdateEarliest(channel, msg.ID)
	}

	if from != it.Pos {
		Info("backfilled %v messages for '%v'", count, channel)
	}
}
