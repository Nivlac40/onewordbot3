package main

import (
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
)

func guildCreateEvent(e *gateway.GuildCreateEvent) {
	if _, ok := gs[e.Guild.ID]; ok {
		gs[e.Guild.ID].guildID = e.Guild.ID
		for id, c := range gs[e.Guild.ID].Channels {
			c.channelID = id
		}
	} else {
		gs[e.Guild.ID] = &guild{
			Prefix: botConfig.DefaultPrefix,
			guildID: e.Guild.ID,
		}
		gs[e.Guild.ID].Channels = make(map[discord.ChannelID]*channel)
	}
}

func guildRemoveEvent(e *gateway.GuildDeleteEvent) {
	delete(gs, e.ID)
}

func messageCreateEvent(e *gateway.MessageCreateEvent) {
	if e.Author.Bot == true {
		return
	}
	gs[e.GuildID].processMessageEvent(e)
}

func messageEditEvent(e *gateway.MessageUpdateEvent) {
	if e.Author.Bot == true {
		return
	}
	 if gs[e.GuildID].channelRegistered(e.ChannelID) {
		gs[e.GuildID].Channels[e.ChannelID].processEditEvent(e)
	}
}

func channelDeleteEvent(e *gateway.ChannelDeleteEvent) {
	delete(gs[e.GuildID].Channels, e.ID)
}