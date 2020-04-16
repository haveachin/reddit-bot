package main

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	discord "github.com/bwmarrin/discordgo"
	"github.com/haveachin/reddit-bot/reddit"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	msgStartingBot string = "Loading all resources"
	msgBotIsOnline string = "Bot is now online"
)

const (
	captureNamePrefixMsg  string = "prefix"
	captureNameSubreddit  string = "subreddit"
	captureNamePostID     string = "postID"
	captureNameSuffixMsg  string = "suffix"
	discordTokenBotPrefix string = "Bot "
	discordTokenEnvKey    string = "DISCORD_TOKEN"
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

	const discordBotTokenPrefix string = "Bot"
	discordToken = fmt.Sprintf("%s %s", discordBotTokenPrefix, os.Getenv(discordTokenEnvKey))
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

func onRedditLinkMessage(s *discord.Session, m *discord.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	matches, err := FindStringSubmatch(redditPostPattern, m.Content)
	if err != nil {
		return
	}

	redditPost, err := reddit.PostByID(matches.CaptureByName(captureNamePostID))
	if err != nil {
		const couldNotFetchDataEmojiID string = "⚠️"
		s.MessageReactionAdd(m.ChannelID, m.ID, couldNotFetchDataEmojiID)
		log.Err(err)
		return
	}

	_, err = s.ChannelMessageSendComplex(m.ChannelID, &discord.MessageSend{
		Content: fmt.Sprintf("%s%s", matches.CaptureByName(captureNamePrefixMsg), matches.CaptureByName(captureNameSuffixMsg)),
		Embed: &discord.MessageEmbed{
			Title: redditPost.Title,
			Color: 16728833,
			URL:   fmt.Sprintf("https://reddit.com%s", redditPost.Permalink),
			Author: &discord.MessageEmbedAuthor{
				Name:    m.Author.Username,
				IconURL: m.Author.AvatarURL(""),
			},
			Image: &discord.MessageEmbedImage{
				URL: redditPost.ImageURL,
			},
			Description: fmt.Sprintf("%s by u/%s", redditPost.Subreddit, redditPost.Author),
		},
	})
	if err != nil {
		log.Err(err)
		return
	}

	s.ChannelMessageDelete(m.ChannelID, m.ID)
}
