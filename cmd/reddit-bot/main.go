package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	discord "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
)

const (
	envVarPrefix = "REDDITBOT"
)

var (
	discordToken string
	configPath   = "config.yml"
	logLevel     = "info"

	cfg config
)

func envVarStringVar(p *string, name string) {
	key := envVarPrefix + "_" + name
	v := os.Getenv(key)
	if v == "" {
		return
	}
	*p = v
}

func initEnvVars() {
	envVarStringVar(&discordToken, "DISCORD_TOKEN")
	envVarStringVar(&configPath, "CONFIG")
	envVarStringVar(&logLevel, "LOG_LEVEL")
}

func initFlags() {
	pflag.StringVarP(&configPath, "config", "c", configPath, "path to the config file")
	pflag.StringVarP(&logLevel, "log-level", "l", logLevel, "log level [debug, info, warn, error]")
	pflag.Parse()
}

func initConfig() {
	if err := createConfigIfNotExist(); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to create config if not exist")
		return
	}

	var err error
	cfg, err = readAndParseConfig()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to read and parse config")
		return
	}

	if cfg.DiscordToken != "" {
		discordToken = cfg.DiscordToken
	}
}

func initLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	})

	var level zerolog.Level
	switch logLevel {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	default:
		log.Warn().
			Str("level", logLevel).
			Msg("Invalid log level; defaulting to info")
	}

	zerolog.SetGlobalLevel(level)
	log.Info().
		Str("level", logLevel).
		Msg("Log level set")
}

func main() {
	initEnvVars()
	initFlags()
	initLogger()
	initConfig()

	if err := run(); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to run")
	}
}

func run() error {
	tempDir, err := os.MkdirTemp(os.TempDir(), "reddit-bot")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chdir(tempDir); err != nil {
		return err
	}

	log.Info().Msg("Connecting to Discord")
	discordSession, err := discord.New("Bot " + discordToken)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	defer discordSession.Close()

	rb := newRedditBot(cfg.PostProcessingArgs)

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

	log.Info().Msg("Bot is online")

	// Wait for exit signal form OS
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	return nil
}
