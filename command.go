package main

import (
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
	"strconv"
	"strings"
	"time"
)

var cooldown = make(map[discord.UserID]time.Time)

func (g *guild) processCommand(e *gateway.MessageCreateEvent, args []string) bool {
	for _, c := range commands {
		for _, trigger := range c.triggers {
			if trigger == strings.ToLower(args[0]) {
				if c.ascended {
					for _, u := range botConfig.Overlords {
						if u == e.Author.ID {
							c.action(e, args, g)
							return true
						}
					}
					return false
				}

				if !c.admin {
					c.action(e, args, g)
					return true
				}

				if len(e.Member.RoleIDs) == 0 {
					return false
				}

				for _, roleID := range e.Member.RoleIDs {
					role, _ := s.Role(e.GuildID, roleID)
					if (role.Permissions&discord.PermissionAdministrator != 0) || (role.Permissions&discord.PermissionManageGuild != 0) {
						c.action(e, args, g)
						return true
					}
				}
			}
		}
	}
	return false
}

func cmddown(id discord.UserID) bool {
	if _, ok := cooldown[id]; !ok {
		cooldown[id] = time.Now()
		return true
	}

	 if time.Now().Sub(cooldown[id]) < time.Millisecond * 800 {
	 	cooldown[id] = time.Now()
	 	return false
	 }

	 cooldown[id] = time.Now()
	 return true
}

type commandAction func(e *gateway.MessageCreateEvent, c []string, g *guild)

type command struct {
	triggers []string
	admin bool
	ascended bool
	action commandAction
}

var commands = []command {{
	triggers: []string{"prefix", "p"},
	admin: true,
	ascended: false,
	action: func(e *gateway.MessageCreateEvent, c []string, g *guild) {
		if len(c) < 2 {
			s.SendText(e.ChannelID, "Please provide an argument")
		} else if len(c[1]) >= 4 {
			s.SendText(e.ChannelID, "That prefix is too long")
		} else {
			g.Prefix = c[1]
			s.SendText(e.ChannelID, "The prefix was set to "+g.Prefix)
		}
	},
}, {
	triggers: []string{"register", "reg"},
	admin: true,
	ascended: false,
	action: func(e *gateway.MessageCreateEvent, c []string, g *guild) {
		if !g.channelRegistered(e.ChannelID) {
			if len(g.Channels) > 1 {
				s.SendText(e.ChannelID, "Channel Limit Reached")
			} else {
				g.regChannel(e.ChannelID)
				s.SendText(e.ChannelID, "Channel Registered")
			}
		} else if g.channelRegistered(e.ChannelID) {
			g.delChannel(e.ChannelID)
			s.SendText(e.ChannelID, "Channel Unregistered")
		}
	},
}, {
	triggers: []string{"ping"},
	admin: false,
	ascended: false,
	action: func(e *gateway.MessageCreateEvent, c []string, g *guild) {
		now := time.Now()
		msg, _ := s.SendText(e.ChannelID, "Pong")
		s.EditMessage(e.ChannelID, msg.ID, "Pong " + strconv.FormatInt(msg.Timestamp.Time().Sub(now).Milliseconds(), 10) + "ms", nil, false)
	},
}, {
	triggers: []string{"pong"},
	admin: false,
	ascended: false,
	action: func(e *gateway.MessageCreateEvent, c []string, g *guild) {
		s.SendText(e.ChannelID, "Its \"ping\" dumbass")
	},
}, {
	triggers: []string{"status"},
	admin: false,
	ascended: true,
	action: func(e *gateway.MessageCreateEvent, c []string, g *guild) {
		s.Gateway.UpdateStatus(gateway.UpdateStatusData{
			Since: 0,
			Activities: []discord.Activity{{
				Name: strings.Join(c[1:], " "),
				URL: "",
				Type: 0,
			}},
			Status: gateway.DoNotDisturbStatus,
			AFK: false,
		})

		s.SendText(e.ChannelID, "Status Updated")
	},
}, {
	triggers: []string{"output", "out"},
	admin: true,
	ascended: false,
	action: func(e *gateway.MessageCreateEvent, c []string, g *guild) {
		if len(c) < 2 {
			s.SendText(e.ChannelID, "Please provide a channel ID")
			return
		}

		rawID, err := strconv.ParseUint(c[1], 10, 64)
		if err != nil {
			s.SendText(e.ChannelID, "Invalid Number")
		}

		id := discord.ChannelID(rawID)

		if g.channelRegistered(id) {
			if e.ChannelID == g.Channels[id].OutputChannel {
				g.Channels[id].OutputChannel = 0
				s.SendText(e.ChannelID, "Output Cleared")
			} else {
				g.Channels[id].OutputChannel = e.ChannelID
				s.SendText(e.ChannelID, "Output Set")
			}
		} else {
			s.SendText(e.ChannelID, "Invalid Channel")
		}
	},
}, {
	triggers: []string{"clean"},
	admin: true,
	ascended: false,
	action: func(e *gateway.MessageCreateEvent, c []string, g *guild) {
		for id, _ := range g.Channels {
			_, err := s.Channel(id)
			if err != nil {
				g.delChannel(id)
			}
		}
		s.SendText(e.ChannelID, "Cleared deleted channels")
	},
}, {
	triggers: []string{"config"},
	admin:    true,
	ascended: false,
	action: func(e *gateway.MessageCreateEvent, c []string, g *guild) {
		if !g.channelRegistered(e.ChannelID) {
			s.SendText(e.ChannelID, "This channel is not registered")
			return
		}

		if len(c) == 1 {
			s.SendText(e.ChannelID,"```json\n" + g.Channels[e.ChannelID].getJson() + "```")
		} else if len(c) > 1 {
			j := strings.Join(c[1:], " ")
			j = strings.TrimPrefix(j , "```json")
			j = strings.TrimPrefix(j , "```")
			j = strings.TrimSuffix(j , "```")

			err := g.Channels[e.ChannelID].writeJson(j)

			if err != nil {
				s.SendText(e.ChannelID, err.Error())
			} else {
				s.SendText(e.ChannelID, "Successfully Applied Config")
			}
		}
	},
}, {
	triggers: []string{"setlog", "setlogchannel"},
	admin: true,
	ascended: false,
	action: func(e *gateway.MessageCreateEvent, c []string, g *guild) {
		if g.LogChannel.IsNull() {
			g.LogChannel = 0
		}

		if g.LogChannel == e.ChannelID {
			g.LogChannel = 0
			s.SendText(e.ChannelID, "Log channel cleared")
		} else {
			g.LogChannel = e.ChannelID
			s.SendText(e.ChannelID, "Log channel set")
		}
	},
}}
