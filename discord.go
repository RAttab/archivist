package main

import (
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
		Fatal("unable to connect to discord: %v", err)
	}
}

func DiscordClose() {
	if err := discord.Close(); err != nil {
		Warning("unable to close discord connection: %v", err)
	}
}

const (
	forward   = iota
	backwards = iota
)

type MessageSorter struct {
	msgs      []*discordgo.Message
	direction int
}

func (sorter MessageSorter) Len() int {
	return len(sorter.msgs)
}

func (sorter MessageSorter) Less(lhs, rhs int) bool {
	if sorter.direction == forward {
		return sorter.msgs[lhs].ID < sorter.msgs[rhs].ID
	} else {
		return sorter.msgs[rhs].ID < sorter.msgs[lhs].ID
	}
}

func (sorter MessageSorter) Swap(lhs, rhs int) {
	tmp := sorter.msgs[lhs]
	sorter.msgs[lhs] = sorter.msgs[rhs]
	sorter.msgs[rhs] = tmp
}

type MessageIt struct {
	Channel string
	Pos     string

	pending   []*discordgo.Message
	index     int
	direction int
}

func (it *MessageIt) DiscordItNext() (*Message, error) {
	if it.index == len(it.pending) {
		var err error

		before := ""
		if it.direction == backwards {
			before = it.Pos
		}

		after := ""
		if it.direction == forward {
			after = it.Pos
		}

		if it.pending, err = discord.ChannelMessages(it.Channel, 100, before, after, ""); err != nil {
			return nil, err
		}
		it.index = 0
		sort.Sort(MessageSorter{it.pending, it.direction})
	}

	if len(it.pending) == 0 {
		return nil, nil
	}

	msg := it.pending[it.index]
	it.index++

	it.Pos = msg.ID
	return (*Message)(msg), nil
}

func DiscordMessageForwardIt(channel string, from string) (*MessageIt, error) {
	return &MessageIt{Channel: channel, Pos: from, direction: forward}, nil
}

func DiscordMessageBackwardsIt(channel string, from string) (*MessageIt, error) {
	return &MessageIt{Channel: channel, Pos: from, direction: backwards}, nil
}

func DiscordTimestamp(id string) (time.Time, error) {
	return discordgo.SnowflakeTimestamp(id)
}

func DiscordChannelName(id string) string {
	channel, err := discord.Channel(id)
	if err != nil {
		Fatal("unable to fetch channel name for '%v': %v", id, err)
	}
	return channel.Name
}
