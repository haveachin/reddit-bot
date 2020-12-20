package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const configPath = "config.json"

type config struct {
	DiscordToken string `json:"discordToken"`
}

func loadConfig() (config, error) {
	var cfg config

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		_ = saveConfig(cfg)
		return cfg, nil
	}

	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		return cfg, err
	}

	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func saveConfig(cfg config) error {
	b, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(b); err != nil {
		return err
	}

	return nil
}
