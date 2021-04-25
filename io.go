package main

import (
	"encoding/json"
	"github.com/diamondburned/arikawa/discord"
	"io/ioutil"
	"os"
	"time"
)

type botConfigStruct struct {
	Token string `json:"token"`
	DefaultPrefix string `json:"prefix"`
	AutosaveSpeed int `json:"autosave"`
	Overlords []discord.UserID `json:"overlords"`
}

func createConfig(path string) bool {
	if fileExists(path) {
		data, err := ioutil.ReadFile(path) ; Panic(err)
		s := botConfigStruct{}
		err = json.Unmarshal(data, &s) ; Panic(err)
		data, err = json.MarshalIndent(s,"","\t") ; Panic(err)
		err = ioutil.WriteFile(path, data, 0764) ; Panic(err)
		return true
	} else {
		data, err := json.MarshalIndent(botConfigStruct{},"","\t") ; Panic(err)
		err = ioutil.WriteFile(path, data, 0764) ; Panic(err)
		return false
	}
}

func readConfig(path string) botConfigStruct {
	data, err := ioutil.ReadFile(path) ; Panic(err)
	s := botConfigStruct{}
	err = json.Unmarshal(data, &s) ; Panic(err)
	return s
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func writeGuildData(path string) {
	data, err := json.MarshalIndent(gs,"","\t") ; Panic(err)
	err = ioutil.WriteFile(path, data, 0764) ; Panic(err)
}

func readGuildData(path string) {
	data, err := ioutil.ReadFile(path) ; Panic(err)
	if string(data) == "" {
		err = json.Unmarshal([]byte("{}"), &gs) ; Panic(err)
	} else {
		err = json.Unmarshal(data, &gs) ; Panic(err)
	}
}

func autosaveLoop(delay int, path string) {
	for {
		time.Sleep(time.Second * time.Duration(delay))
		writeGuildData(path)
	}
}
