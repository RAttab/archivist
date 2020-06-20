package main

import (
	"strings"
	"time"
)

const sleep = 1 * time.Second

func ScrapperStart() {
	for guild, channels := range Config.Discord.Subs {
		for _, channelId := range channels {
			earliest, latest := DatabaseSub(channelId)
			Info("scrape guild:%v chan:%v fwd:%v back:%v", guild, channelId, latest, earliest)

			channelName := DiscordChannelName(channelId)
			go scrapeForward(guild, channelId, channelName, latest)
			go scrapeBackwards(guild, channelId, channelName, earliest)
		}
	}
}

func validAttachment(attach *Attachment) bool {
	file := attach.Filename
	return strings.HasSuffix(file, ".jpg") ||
		strings.HasSuffix(file, ".jpeg") ||
		strings.HasSuffix(file, ".png") ||
		strings.HasSuffix(file, ".gif")
}

func message(guild, channelName string, msg *Message) (bool, error) {
	kept := false
	// Debug("filter guild:%v chan:%v msg:%v", guild, msg.ChannelID, msg.ID)

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
			Tags:      []string{"@" + msg.Author.Username, "#" + channelName},
		}

		id := DatabaseRecordInsert(rec)
		kept = true

		Info("put %v -> guild:%v chan:%v msg:%v img:%v",
			id, rec.GuildId, rec.ChannelId, rec.MessageId, rec.ImageId)
	}

	return kept, nil
}

func scrapeForward(guild, channelId, channelName, from string) {
	it, err := DiscordMessageForwardIt(channelId, from)
	if err != nil {
		Fatal("unable to create iterator for '%v': %v", channelId, it)
	}

	for true {
		msg, err := it.DiscordItNext()
		if err != nil {
			Fatal("unable to read message for '%v': %v", channelId, err)
		}
		if msg == nil {
			time.Sleep(sleep)
			continue
		}

		if _, err := message(guild, channelName, msg); err != nil {
			Fatal("ERROR: failed to parse '%v': %v", msg.ID, err)
		}

		DatabaseSubUpdateLatest(channelId, msg.ID)
	}
}

func scrapeBackwards(guild, channelId, channelName, from string) {
	it, err := DiscordMessageBackwardsIt(channelId, from)
	if err != nil {
		Fatal("unable to create iterator for '%v': %v", channelId, it)
	}

	count := 0
	for true {
		msg, err := it.DiscordItNext()
		if msg == nil {
			break
		} else if err != nil {
			Fatal("unable to read message for '%v': %v", channelId, err)
			return
		}

		if kept, err := message(guild, channelName, msg); err != nil {
			Warning("failed to parse '%v': %v", msg.ID, err)
		} else if kept {
			count++
		}

		DatabaseSubUpdateEarliest(channelId, msg.ID)
	}

	if from != it.Pos {
		Info("backfilled %v messages for '%v'", count, channelId)
	}
}
