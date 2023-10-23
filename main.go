package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	discord "github.com/bwmarrin/discordgo"
	"github.com/haveachin/reddit-bot/regex"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	pattern              string = `(?s)(?P<%s>.*)https:\/\/(?:www.)?reddit.com\/r\/(?P<%s>.+)\/(?:comments|s)\/(?P<%s>.+?)\/?[^\s\n]*\s?(?P<%s>.*)`
	captureNamePrefixMsg string = "prefix"
	captureNameSubreddit string = "subreddit"
	captureNamePostID    string = "postID"
	captureNameSuffixMsg string = "suffix"

	envVarPrefix = "REDDITBOT_"
)

var (
	redditPostPattern regex.Pattern
	discordToken      string
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

func init() {
	initEnvVars()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	redditPostPattern = regex.MustCompile(
		pattern,
		captureNamePrefixMsg,
		captureNameSubreddit,
		captureNamePostID,
		captureNameSuffixMsg,
	)

	if err := os.Mkdir("logs", 0644); err != nil {
		if os.IsNotExist(err) {
			log.Error().Err(err).Msg("Could not create logs folder")
		}
	}

	discordToken = fmt.Sprintf("Bot %s", discordToken)
}

func main() {
	log.Info().Msg("Connecting to Discord")
	discordSession, err := discord.New(discordToken)
	if err != nil {
		log.Error().Err(err).Msg("Could not create session")
		return
	}
	defer discordSession.Close()

	discordSession.AddHandler(onRedditLinkMessage)

	if err := discordSession.Open(); err != nil {
		log.Error().Err(err).Msg("Could not connect to discord")
		return
	}

	for _, g := range discordSession.State.Guilds {
		g, err := discordSession.Guild(g.ID)
		if err != nil {
			log.Error().Err(err)
			continue
		}
		log.Print(g.Name)
	}

	status := fmt.Sprintf("Reddit for %d Servers", len(discordSession.State.Guilds))
	discordSession.UpdateWatchStatus(0, status)

	log.Info().Msg("Bot is online")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
