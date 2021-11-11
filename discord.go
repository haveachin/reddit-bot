package main

import (
	"fmt"
	"html"
	"io"
	"os"
	"time"

	discord "github.com/bwmarrin/discordgo"
	"github.com/haveachin/reddit-bot/embed"
	"github.com/haveachin/reddit-bot/reddit"
	"github.com/rs/zerolog/log"
)

const (
	colorReddit        int    = 16729344
	emojiIDWorkingOnIt string = "🎞️"
	emojiIDErrorReddit string = "⚠️"
	emojiIDErrorFFMPEG string = "😵"
	emojiIDTooBig      string = "\U0001F975"
)

var (
	embedder = embed.NewEmbedder()
)

func onRedditLinkMessage(s *discord.Session, m *discord.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	matches, err := redditPostPattern.FindStringSubmatch(m.Content)
	if err != nil {
		return
	}

	postId := matches.CaptureByName(captureNamePostID)
	logger := log.With().Str("postId", postId).Logger()
	logger.Info().Msg("Fetching post metadata")
	post, err := reddit.PostByID(postId)
	if err != nil {
		logger.Error().Err(err).Msg("Could not fetch post metadata")
		s.ChannelMessageSendReply(m.ChannelID, "Reddit did not respond :(", m.Reference())
		s.MessageReactionAdd(m.ChannelID, m.ID, emojiIDErrorReddit)
		return
	}

	prefixMsg := matches.CaptureByName(captureNamePrefixMsg)
	suffixMsg := matches.CaptureByName(captureNameSuffixMsg)
	permalink := fmt.Sprintf("https://reddit.com%s", post.Permalink)
	title := fmt.Sprintf("r/%s - %s", post.Subreddit, post.Title)
	description := post.Text
	if len(description) > 1000 {
		description = html.UnescapeString(fmt.Sprintf("%.1000s...", post.Text))
	}
	footer := fmt.Sprintf("by u/%s", post.Author)

	msg := &discord.MessageSend{
		Content: prefixMsg + suffixMsg,
		Embed: &discord.MessageEmbed{
			Type: discord.EmbedTypeVideo,
			Author: &discord.MessageEmbedAuthor{
				Name:    m.Author.Username,
				IconURL: m.Author.AvatarURL(""),
			},
			Title:       title,
			Color:       colorReddit,
			URL:         permalink,
			Description: description,
			Footer: &discord.MessageEmbedFooter{
				Text: footer,
			},
		},
	}

	if post.IsVideo {
		s.MessageReactionAdd(m.ChannelID, m.ID, emojiIDWorkingOnIt)
		logger.Info().Msg("Processing post video")
		file, eventLog, err := post.Video.DownloadVideo()
		if err != nil && file == nil {
			logger.Error().Err(err).Msg("ffmpeg error")
			s.ChannelMessageSendReply(m.ChannelID, "Oh, no! Something went wrong while processing your video", m.Reference())
			s.MessageReactionAdd(m.ChannelID, m.ID, emojiIDErrorFFMPEG)
			return
		}
		defer func() {
			file.Close()
			os.Remove(file.Name())
		}()

		if err := saveEventLog(eventLog); err != nil {
			logger.Error().Err(err).Msg("Could not save event log")
		}

		logger.Info().Msg("Embedding video file")
		msg.File = &discord.File{
			Name:   postId + ".mp4",
			Reader: file,
		}
	} else if post.IsImage {
		logger.Info().Msg("Embedding image url")
		msg.Embed.Image = &discord.MessageEmbedImage{
			URL: post.ImageURL,
		}
	} else if post.IsEmbed {
		url, err := embedder.Embed(&post)
		if err == embed.ErrorNotImplemented {
			logger.Warn().Err(err).Msg("embedded website (source) is not yet implemented")
		} else if err != nil {
			logger.Error().Err(err).Msg("something went wrong while analyzing embedded content")
		}
		s.ChannelMessageSend(m.ChannelID, url)
		logger.Info().Msg("Sending embedded YouTube video")
	}

	_, err = s.ChannelMessageSendComplex(m.ChannelID, msg)
	if err != nil {
		logger.Error().Err(err).Msg("Could not send embed")
		s.ChannelMessageSendReply(m.ChannelID, "The video is too big", m.Reference())
		s.MessageReactionAdd(m.ChannelID, m.ID, emojiIDTooBig)
		return
	}

	s.ChannelMessageDelete(m.ChannelID, m.ID)
}

func saveEventLog(eventLog []byte) error {
	file, err := os.Create(fmt.Sprintf("./logs/%d.txt", time.Now().UnixNano()))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, string(eventLog))
	return err
}
