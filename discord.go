package main

import (
	"fmt"
	"html"
	"os"

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

	linkType := matches.CaptureByName(captureNameLinkType)
	logger := log.With().Str("linkType", linkType).Logger()
	if reddit.LinkType(linkType) == reddit.ShareLinkType {
		subreddit := matches.CaptureByName(captureNameSubreddit)
		shareID := matches.CaptureByName(captureNamePostID)
		logger = logger.With().Str("shareID", shareID).Logger()
		url, err := reddit.ResolvePostURLFromShareID(subreddit, shareID)
		if err != nil {
			logger.Error().Err(err).Msg("Could not fetch post metadata")
			s.ChannelMessageSendReply(m.ChannelID, "Share link could not be resolved :(", m.Reference())
			s.MessageReactionAdd(m.ChannelID, m.ID, emojiIDErrorReddit)
			return
		}

		prefixMsg := matches.CaptureByName(captureNamePrefixMsg)
		suffixMsg := matches.CaptureByName(captureNameSuffixMsg)
		content := fmt.Sprintf("%s%s %s", prefixMsg, url, suffixMsg)
		matches, err = redditPostPattern.FindStringSubmatch(content)
		if err != nil {
			return
		}
	}

	postID := matches.CaptureByName(captureNamePostID)
	logger = logger.With().Str("postID", postID).Logger()
	logger.Info().Msg("Fetching post metadata")
	post, err := reddit.PostByID(postID)
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
		file, err := post.DownloadVideo()
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

		logger.Info().Msg("Embedding video file")
		msg.File = &discord.File{
			Name:   postID + ".mp4",
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
