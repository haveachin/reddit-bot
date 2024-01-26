package main

import (
	"errors"
	"os"

	"github.com/haveachin/reddit-bot/configs"
	"gopkg.in/yaml.v3"
)

type config struct {
	DiscordToken       string   `yaml:"discordToken"`
	PostProcessingArgs []string `yaml:"postProcessingArgs"`
}

func createConfigIfNotExist() error {
	info, err := os.Stat(configPath)
	if errors.Is(err, os.ErrNotExist) {
		return createDefaultConfigFile()
	}

	if info.IsDir() {
		return errors.New("is a directory")
	}

	return nil
}

func createDefaultConfigFile() error {
	bb := configs.DefaultConfig
	return os.WriteFile(configPath, bb, 0664)
}

func readAndParseConfig() (config, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return config{}, err
	}
	defer f.Close()

	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, err
	}
	return cfg, err
}
