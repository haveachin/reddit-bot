package main

import (
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/bwmarrin/discordgo"
	discord "github.com/bwmarrin/discordgo"
)

var redditPostPattern *regexp.Regexp
var discordToken string

func init() {
	redditPostPattern = regexp.MustCompile(`https:\/\/www.reddit.com\/r\/(.+)\/comments\/(.+?)\/.+`)
	discordToken = os.Getenv("DISCORD_TOKEN")
}

func main() {
	discordSession, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatal(err)
	}
	defer discordSession.Close()
	discordSession.AddHandler(onMessage)

	if err := discordSession.Open(); err != nil {
		log.Fatal("error opening connection,", err)
	}

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func onMessage(s *discord.Session, m *discord.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	matches := redditPostPattern.FindStringSubmatch(m.Content)
	if len(matches) < 3 {
		return
	}

	redditPost, err := getPostData(matches[2])
	if err != nil {
		s.MessageReactionAdd(m.ChannelID, m.ID, "⚠️")
		log.Println(err)
		return
	}

	s.ChannelMessageDelete(m.ChannelID, m.ID)
	s.ChannelMessageSendEmbed(m.ChannelID, &discord.MessageEmbed{
		Title: redditPost.title,
		Color: 16728833,
		URL:   m.Content,
		Author: &discord.MessageEmbedAuthor{
			Name:    m.Author.Username,
			IconURL: m.Author.AvatarURL(""),
		},
		Image: &discord.MessageEmbedImage{
			URL: redditPost.imageURL,
		},
		Description: redditPost.subreddit,
	})
}
