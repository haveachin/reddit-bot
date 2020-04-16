package main

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	discord "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	msgStartingBot string = "Loading all resources"
	msgBotIsOnline string = "Bot is now online"
)

const (
	captureNamePrefixMsg string = "prefix"
	captureNameSubreddit string = "subreddit"
	captureNamePostID    string = "postID"
	captureNameSuffixMsg string = "suffix"
	discordBotTokenf     string = "Bot %s"
	discordTokenEnvKey   string = "DISCORD_TOKEN"
)

var (
	redditPostPattern *regexp.Regexp
	discordToken      string
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	redditPostPattern = regexp.MustCompile(
		fmt.Sprintf(
			`(?s)(?P<%s>.*)https:\/\/(?:www.)?reddit.com\/r\/(?P<%s>.+)\/comments\/(?P<%s>.+?)\/[^\s\n]*\s?(?P<%s>.*)`,
			captureNamePrefixMsg,
			captureNameSubreddit,
			captureNamePostID,
			captureNameSuffixMsg,
		),
	)

	discordToken = fmt.Sprintf(discordBotTokenf, os.Getenv(discordTokenEnvKey))
}

func main() {
	discordSession, err := discord.New(discordToken)
	if err != nil {
		log.Err(err)
	}
	defer discordSession.Close()

	discordSession.AddHandler(onRedditLinkMessage)

	if err := discordSession.Open(); err != nil {
		log.Err(err)
	}

	log.Info().Msg(msgBotIsOnline)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
