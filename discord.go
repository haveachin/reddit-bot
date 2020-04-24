package main

import (
	"fmt"

	discord "github.com/bwmarrin/discordgo"
	"github.com/haveachin/reddit-bot/reddit"
	"github.com/rs/zerolog/log"
)

const (
	colorReddit  int    = 16728833
	emojiIDError string = "⚠️"
)

func onRedditLinkMessage(s *discord.Session, m *discord.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	matches, err := redditPostPattern.FindStringSubmatch(m.Content)
	if err != nil {
		return
	}

	post, err := reddit.PostByID(matches.CaptureByName(captureNamePostID))
	if err != nil {
		s.MessageReactionAdd(m.ChannelID, m.ID, emojiIDError)
		log.Err(err)
		return
	}

	prefixMsg := matches.CaptureByName(captureNamePrefixMsg)
	suffixMsg := matches.CaptureByName(captureNameSuffixMsg)
	permalink := fmt.Sprintf("https://reddit.com%s", post.Permalink)
	description := fmt.Sprintf("%s by u/%s", post.Subreddit, post.Author)

	messageSend := &discord.MessageSend{
		Content: prefixMsg + suffixMsg,
		Embed: &discord.MessageEmbed{
			Title: post.Title,
			Color: colorReddit,
			URL:   permalink,
			Author: &discord.MessageEmbedAuthor{
				Name:    m.Author.Username,
				IconURL: m.Author.AvatarURL(""),
			},
			Description: description,
		},
	}

	if post.IsVideo {
		messageSend.Embed.Video = &discord.MessageEmbedVideo{
			URL: post.VideoURL,
		}
		messageSend.Embed.Footer = &discord.MessageEmbedFooter{
			Text: "Watch the video on Reddit",
		}
	} else {
		messageSend.Embed.Image = &discord.MessageEmbedImage{
			URL: post.ImageURL,
		}
	}

	_, err = s.ChannelMessageSendComplex(m.ChannelID, messageSend)
	if err != nil {
		log.Err(err)
		return
	}

	s.ChannelMessageDelete(m.ChannelID, m.ID)
}
