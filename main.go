package main

import (
	"fmt"
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
	// prefix - msg before link; subreddit - subreddit of post;
	// post - ID of post; suffix - msg after link;
	redditPostPattern = regexp.MustCompile(`(?s)(?P<prefix>.*)https:\/\/(?:www.)?reddit.com
		\/r\/(?P<subreddit>.+)\/comments\/(?P<post>.+?)\/[^\s\n]*\s?(?P<suffix>.*)`)
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

	matches, err := FindStringSubmatch(redditPostPattern, m.Content)
	if err != nil {
		return
	}

	redditPost, err := getPostData(matches.CaptureByName("post"))
	if err != nil {
		s.MessageReactionAdd(m.ChannelID, m.ID, "⚠️")
		log.Println(err)
		return
	}

	s.ChannelMessageDelete(m.ChannelID, m.ID)
	s.ChannelMessageSendEmbed(m.ChannelID, &discord.MessageEmbed{
		Title: redditPost.title,
		Color: 16728833,
		URL:   fmt.Sprintf("https://reddit.com%s", redditPost.permalink),
		Author: &discord.MessageEmbedAuthor{
			Name:    m.Author.Username,
			IconURL: m.Author.AvatarURL(""),
		},
		Image: &discord.MessageEmbedImage{
			URL: redditPost.imageURL,
		},
		Description: fmt.Sprintf("%s by u/%s", redditPost.subreddit, redditPost.author),
	})
}
