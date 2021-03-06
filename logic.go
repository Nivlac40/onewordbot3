package main

import (
	"encoding/json"
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
	"strings"
	"time"
)

type guild struct {
	guildID             discord.GuildID
	Prefix              string                         `json:"prefix"`
	Channels            map[discord.ChannelID]*channel `json:"channels"`
	UpgradedUntil       time.Duration                  `json:"upgraded_until"`
	BlacklistedWords    []string                       `json:"blacklisted_words"`
	BlacklistedAccounts []discord.UserID               `json:"blacklisted_accounts"`
}

var lock = false

func (g *guild) processMessageEvent(e *gateway.MessageCreateEvent) {
	if strings.HasPrefix(e.Content, g.Prefix) && cmddown(e.Author.ID) {
		decodedCmd := decodeCommand(e.Content, g.Prefix)
		if g.processCommand(e, decodedCmd) {
			return
		}
	}

	if g.channelRegistered(e.ChannelID) {
		g.Channels[e.ChannelID].processMessage(e, g)
	}

	if e.Content == me.Mention() {
		s.SendMessage(e.ChannelID, "My prefix here is "+g.Prefix)
		return
	}
}

func decodeCommand(inp, prefix string) []string {
	inp = strings.TrimPrefix(inp, prefix)
	return strings.Split(inp, " ")
}

func (c *channel) processMessage(e *gateway.MessageCreateEvent, g *guild) {
	if e.Content != c.EndTrigger {
		if (len(e.Attachments) != 0) || (len(e.Stickers) != 0) || (e.ReferencedMessage != nil) || (len(e.Embeds) != 0) {
			s.DeleteMessage(e.ChannelID, e.ID, "")
			return
		}

		lastmsg := c.getLastValidMessage()

		for _, account := range g.BlacklistedAccounts {
			if account == e.Author.ID {
				goto invalid
			}
		}

		if lastmsg == nil {
			if c.isLegal(e.Content, g.BlacklistedWords) {
				goto valid
			} else {
				goto invalid
			}
		} else {
			if e.Content == lastmsg.Content && !c.AllowIdentical {
				goto invalid
			}

			if e.Author.ID == lastmsg.Author.ID && !c.AllowSameAuthor {
				goto invalid
			}

			if c.isLegal(e.Content, g.BlacklistedWords) {
				goto valid
			} else {
				goto invalid
			}
		}

	invalid:
		s.DeleteMessage(e.ChannelID, e.ID, "")
		return

	valid:
		c.store1 = append(c.store1, e.ID)
		return

	} else if len(c.store1) != 0 {
		if lock {
			s.DeleteMessage(e.ChannelID, e.ID, "")
			return
		}
		lock = true
		str1 := ""
		for _, ID := range c.store1 {
			msg, err := s.Message(e.ChannelID, ID)
			if err == nil && c.isLegal(msg.Content, g.BlacklistedWords) {
				str1 += c.Separator + msg.Content
			}
		}
		msg, err := s.SendMessage(e.ChannelID, str1)

		if len(str1) != 0 && c.OutputChannel != 0 && !c.OutputChannel.IsNull() {
			s.SendMessage(c.OutputChannel, "", discord.Embed{
				Title:       "",
				Type:        "",
				Description: str1,
				URL:         "",
				Timestamp:   discord.Timestamp{},
				Color:       0x990000,
				Footer:      nil,
				Image:       nil,
				Thumbnail:   nil,
				Video:       nil,
				Provider:    nil,
				Author: &discord.EmbedAuthor{
					Name:      "One Word Bot",
					URL:       "",
					Icon:      me.AvatarURL(),
					ProxyIcon: "",
				},
				Fields: nil,
			})
		}

		if err == nil && c.PinSentences {
			pins, _ := s.PinnedMessages(msg.ChannelID)
			if len(pins) == 50 {
				lastpin := pins[len(pins)-1]
				s.UnpinMessage(lastpin.ChannelID, lastpin.ID, "")
				s.PinMessage(msg.ChannelID, msg.ID, "")
			} else {
				s.PinMessage(msg.ChannelID, msg.ID, "")
			}
		}
	}
	c.store1 = nil
	lock = false
}

func (c *channel) processEditEvent(e *gateway.MessageUpdateEvent) {
	if !c.isLegal(e.Content, gs[e.GuildID].BlacklistedWords) {
		s.React(e.ChannelID, e.ID, "???")
	} else {
		s.Unreact(e.ChannelID, e.ID, "???")
	}
}

func (c *channel) isLegal(msg string, bl []string) bool {
	wordcount := len(strings.Split(msg, " "))

	if (wordcount < c.MinimumWords) || (wordcount > c.MaximumWords) {
		return false
	}

	if (len(msg) < c.MinimumLength) || (len(msg) > c.MaximumLength) {
		return false
	}

	if strings.ContainsAny(msg, c.DisallowedCharacters) {
		return false
	}

	for _, word := range bl {
		if strings.Contains(strings.ToLower(msg), strings.ToLower(word)) {
			return false
		}
	}

	return true
}

type channel struct {
	guildID              discord.GuildID
	channelID            discord.ChannelID
	store1               []discord.MessageID
	AllowIdentical       bool              `json:"allow_identical_words"`
	AllowSameAuthor      bool              `json:"allow_same_author"`
	MinimumWords         int               `json:"minimum_words"`
	MaximumWords         int               `json:"maximum_words"`
	MinimumLength        int               `json:"minimum_length"`
	MaximumLength        int               `json:"maximum_length"`
	EndTrigger           string            `json:"end_trigger"`
	DisallowedCharacters string            `json:"disallowed_characters"`
	PinSentences         bool              `json:"pin_sentences"`
	OutputChannel        discord.ChannelID `json:"output_channel"`
	Separator            string            `json:"separator"`
}

func (c *channel) getLastValidMessage() *discord.Message {
	for i, _ := range c.store1 {
		msg, err := s.Message(c.channelID, c.store1[len(c.store1)-i-1])
		if err == nil {
			return msg
		}
	}
	return nil
}

func (g *guild) regChannel(id discord.ChannelID) {
	if _, ok := g.Channels[id]; !ok {
		g.Channels[id] = &channel{
			channelID:            id,
			AllowIdentical:       true,
			AllowSameAuthor:      false,
			MinimumWords:         1,
			MaximumWords:         1,
			MinimumLength:        1,
			MaximumLength:        14,
			EndTrigger:           ".",
			DisallowedCharacters: ":<>@`\n/_",
			PinSentences:         true,
			OutputChannel:        0,
			Separator:            " ",
		}
	}
}

func (g *guild) delChannel(id discord.ChannelID) {
	delete(g.Channels, id)
}

func (g *guild) getChannel(id discord.ChannelID) *channel {
	if _, ok := g.Channels[id]; ok {
		return g.Channels[id]
	} else {
		return nil
	}
}

func (g *guild) channelRegistered(id discord.ChannelID) bool {
	if _, ok := g.Channels[id]; ok {
		return true
	} else {
		return false
	}
}

func (c *channel) getJson() string {
	raw, err := json.MarshalIndent(c, "", "\t")
	Panic(err)
	return string(raw)
}

func (c *channel) writeJson(j string) error {
	err := json.Unmarshal([]byte(j), c)
	return err
}
