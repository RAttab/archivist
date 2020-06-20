package main

import (
	"log"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
)

var discord *discordgo.Session

type Message discordgo.Message
type Attachment discordgo.MessageAttachment

func DiscordConnect() {
	var err error
	if discord, err = discordgo.New("Bot " + Config.Discord.Token); err != nil {
		log.Fatalf("unable to connect to discord: %v", err)
	}
}

func DiscordClose() {
	if err := discord.Close(); err != nil {
		log.Printf("unable to close discord connection: %v", err)
	}
}

type MessageIt struct {
	Channel string
	Pos     string

	pending []*discordgo.Message
	index   int
}

type MessageSorter []*discordgo.Message

func (sorter MessageSorter) Len() int {
	return len(sorter)
}

func (sorter MessageSorter) Less(lhs, rhs int) bool {
	return sorter[lhs].ID < sorter[rhs].ID
}

func (sorter MessageSorter) Swap(lhs, rhs int) {
	tmp := sorter[lhs]
	sorter[lhs] = sorter[rhs]
	sorter[rhs] = tmp
}

func (it *MessageIt) DiscordIteratorNext() (*Message, error) {
	if it.index == len(it.pending) {
		var err error
		if it.pending, err = discord.ChannelMessages(it.Channel, 100, "", it.Pos, ""); err != nil {
			return nil, err
		}
		it.index = 0
		sort.Sort(MessageSorter(it.pending))
	}

	if len(it.pending) == 0 {
		return nil, nil
	}

	msg := it.pending[it.index]
	it.index++

	it.Pos = msg.ID
	return (*Message)(msg), nil
}

func DiscordMessageIterator(channel string, since string) (*MessageIt, error) {
	return &MessageIt{Channel: channel, Pos: since}, nil
}

func DiscordTimestamp(id string) (time.Time, error) {
	return discordgo.SnowflakeTimestamp(id)
}
