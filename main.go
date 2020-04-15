package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	discord "github.com/bwmarrin/discordgo"
)

const (
	captureNamePrefix        string = "prefix"
	captureNameSubreddit     string = "subreddit"
	captureNamePostID        string = "postID"
	captureNameSuffix        string = "suffix"
	redditPostPatternStringf string = `(?s)(?P<%s>.*)https:\/\/(?:www.)?reddit.com\/r\/(?P<%s>.+)\/comments\/(?P<%s>.+?)\/[^\s\n]*\s?(?P<%s>.*)`
	discordTokenEnv          string = "DISCORD_TOKEN"
)

var (
	redditPostPattern *regexp.Regexp
	discordToken      string
)

func init() {
	// prefix - msg before link; subreddit - subreddit of post;
	// post - ID of post; suffix - msg after link;
	redditPostPattern = regexp.MustCompile(
		fmt.Sprintf(
			redditPostPatternStringf,
			captureNamePrefix,
			captureNameSubreddit,
			captureNamePostID,
			captureNameSuffix,
		),
	)

	discordToken = os.Getenv(discordTokenEnv)
}

func main() {
	discordSession, err := discord.New("Bot " + discordToken)
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
		log.Println(err)
		return
	}

	redditPost, err := getPostData(matches.CaptureByName(captureNamePostID))
	if err != nil {
		s.MessageReactionAdd(m.ChannelID, m.ID, "⚠️")
		log.Println(err)
		return
	}

	_, err = s.ChannelMessageSendComplex(m.ChannelID, &discord.MessageSend{
		Content: fmt.Sprintf("%s%s", matches.CaptureByName(captureNamePrefix), matches.CaptureByName(captureNameSuffix)),
		Embed: &discord.MessageEmbed{
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
		},
	})
	if err != nil {
		log.Println(err)
		return
	}

	s.ChannelMessageDelete(m.ChannelID, m.ID)
}
