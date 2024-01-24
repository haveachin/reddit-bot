package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	discord "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	envVarPrefix = "REDDITBOT_"
)

var (
	discordToken string
)

func envVarStringVar(p *string, name, value string) {
	key := envVarPrefix + name
	if v := os.Getenv(key); v != "" {
		*p = v
	} else {
		*p = value
	}
}

func initEnvVars() {
	envVarStringVar(&discordToken, "DISCORD_TOKEN", discordToken)
}

func initLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func main() {
	initEnvVars()
	initLogger()

	if err := run(); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to run")
	}
}

func run() error {
	log.Info().Msg("Connecting to Discord")
	discordSession, err := discord.New("Bot " + discordToken)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	defer discordSession.Close()

	rb := newRedditBot()

	rmReddidLinkMsg := discordSession.AddHandler(rb.onRedditLinkMessage)
	defer rmReddidLinkMsg()

	if err := discordSession.Open(); err != nil {
		return fmt.Errorf("connect to discord: %w", err)
	}

	status := fmt.Sprintf("Reddit for %d Servers", len(discordSession.State.Guilds))
	if err := discordSession.UpdateWatchStatus(0, status); err != nil {
		log.Warn().
			Err(err).
			Msg("Failed to update status")
	}

	log.Info().
		Msg("Bot is online")

	// Wait for exit signal form OS
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	return nil
}
