package main

import (
	"context"
	"fmt"
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
	"github.com/diamondburned/arikawa/state"
	"os"
	"os/signal"
	"syscall"
)

var (
	s         *state.State
	botConfig botConfigStruct
	gs        = make(map[discord.GuildID]*guild)
	me        *discord.User
)

func main() {
	killsignal := make(chan os.Signal, 1)
	signal.Notify(killsignal, syscall.SIGINT, syscall.SIGQUIT)

	configFile := "bot.json"
	dataFile := "data.json"

	if createConfig(configFile) {
		botConfig = readConfig(configFile)
	} else {
		fmt.Println("Please configure the bot (" + configFile + ")")
		os.Exit(0)
	}

	if fileExists(dataFile) {
		readGuildData(dataFile)
	}

	var err error
	s = state.NewWithIntents("Bot "+botConfig.Token, gateway.IntentGuildMessages|gateway.IntentGuilds)

	s.AddHandler(messageCreateEvent)
	s.AddHandler(messageEditEvent)
	s.AddHandler(guildCreateEvent)
	s.AddHandler(guildRemoveEvent)
	s.AddHandler(channelDeleteEvent)

	err = s.Open(context.Background())
	Panic(err)

	me, err = s.Me()
	Panic(err)

	go autosaveLoop(botConfig.AutosaveSpeed, dataFile)

	<-killsignal
	gus, err := s.Guilds()

	if err == nil {
		gt := make(map[discord.GuildID]*guild)
		for _, g := range gus {
			gt[g.ID] = gs[g.ID]
		}
		gs = gt
	}

	err = s.Close()
	Panic(err)
	writeGuildData(dataFile)
}
