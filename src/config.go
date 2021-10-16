package main

import (
	"encoding/json"
	"io/ioutil"
)

type config struct {
	Port                    string
	MaxClipboardItemCount   int
	MaxFileSize             string
	MaxDeviceCount          int
	PreserveClipboardOnExit bool
}

func loadConfig(path string) config {
	file, err := ioutil.ReadFile(path)
	config := config{}

	if err != nil {
		return config
	}

	_ = json.Unmarshal([]byte(file), &config)

	return config
}
