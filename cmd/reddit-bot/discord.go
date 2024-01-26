package main

import (
	"fmt"
	"html"
	"os"

	discord "github.com/bwmarrin/discordgo"
	"github.com/haveachin/reddit-bot/embed"
	"github.com/haveachin/reddit-bot/reddit"
	"github.com/haveachin/reddit-bot/regex"
	"github.com/rs/zerolog/log"
)

const (
	colorReddit        int    = 16729344
	emojiIDWorkingOnIt string = "🎞️"
	emojiIDErrorReddit string = "⚠️"
	emojiIDErrorFFMPEG string = "😵"
	emojiIDTooBig      string = "\U0001F975"
	emojiIDWasRemoved  string = "❌"

	captureNamePrefixMsg string = "prefix"
	captureNameSubreddit string = "subreddit"
	captureNameLinkType  string = "linkType"
	captureNamePostID    string = "postID"
	captureNameSuffixMsg string = "suffix"
)

type redditBot struct {
	redditPostPattern  regex.Pattern
	embedder           embed.Embedder
	postProcessingArgs []string
}

func newRedditBot(ppas []string) redditBot {
	return redditBot{
		redditPostPattern: regex.MustCompile(
			`(?s)(?P<%s>.*)https:\/\/(?:www.|new.)?reddit.com\/r\/(?P<%s>.+)\/(?P<%s>comments|s)\/(?P<%s>[^\s\n\/]+)\/?[^\s\n]*\s?(?P<%s>.*)`,
			captureNamePrefixMsg,
			captureNameSubreddit,
			captureNameLinkType,
			captureNamePostID,
			captureNameSuffixMsg,
		),
		embedder:           embed.NewEmbedder(),
		postProcessingArgs: ppas,
	}
}

func (rb redditBot) onRedditLinkMessage(s *discord.Session, m *discord.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	matches, err := rb.redditPostPattern.FindStringSubmatch(m.Content)
	if err != nil {
		return
	}
	prefixMsg := matches.CaptureByName(captureNamePrefixMsg)
	suffixMsg := matches.CaptureByName(captureNameSuffixMsg)
	linkType := matches.CaptureByName(captureNameLinkType)

	if reddit.LinkType(linkType) == reddit.ShareLinkType {
		subreddit := matches.CaptureByName(captureNameSubreddit)
		shareID := matches.CaptureByName(captureNamePostID)

		url, err := reddit.ResolvePostURLFromShareID(subreddit, shareID)
		if err != nil {
			log.Error().
				Err(err).
				Str("shareID", shareID).
				Msg("Could not resolve post with share ID")

			_, _ = s.ChannelMessageSendReply(m.ChannelID, "Share link could not be resolved :(", m.Reference())
			_ = s.MessageReactionAdd(m.ChannelID, m.ID, emojiIDErrorReddit)
			return
		}

		matches, err = rb.redditPostPattern.FindStringSubmatch(url)
		if err != nil {
			log.Error().
				Err(err).
				Str("url", url).
				Msg("Failed to re-match URL with the pattern")
			return
		}
	}

	postID := matches.CaptureByName(captureNamePostID)
	logger := log.With().
		Str("postID", postID).
		Logger()

	logger.Info().Msg("Fetching post metadata")
	post, err := reddit.PostByID(postID)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("Could not fetch post metadata")

		_, _ = s.ChannelMessageSendReply(m.ChannelID, "Reddit did not respond :(", m.Reference())
		_ = s.MessageReactionAdd(m.ChannelID, m.ID, emojiIDErrorReddit)
		return
	}

	msg := &discord.MessageSend{
		Content: prefixMsg + " " + suffixMsg,
		Embed: &discord.MessageEmbed{
			Type: discord.EmbedTypeVideo,
			Author: &discord.MessageEmbedAuthor{
				Name:    m.Author.Username,
				IconURL: m.Author.AvatarURL(""),
			},
			Title: "r/" + post.Subreddit + " - " + post.Title,
			Color: colorReddit,
			URL:   "https://reddit.com" + post.Permalink,
			Description: func() string {
				if len(post.Text) > 1000 {
					return html.UnescapeString(fmt.Sprintf("%.1000s...", post.Text))
				}
				return post.Text
			}(),
			Footer: &discord.MessageEmbedFooter{
				Text: "by u/" + post.Author,
			},
		},
	}

	if post.WasRemoved {
		_ = s.MessageReactionAdd(m.ChannelID, m.ID, emojiIDWasRemoved)
		return
	}

	if post.IsVideo {
		s.MessageReactionAdd(m.ChannelID, m.ID, emojiIDWorkingOnIt)
		logger.Info().Msg("Processing post video")
		post.PostProcessingArgs = rb.postProcessingArgs
		file, err := post.DownloadVideo()
		if err != nil && file == nil {
			logger.Error().
				Err(err).
				Msg("ffmpeg error")
			_, _ = s.ChannelMessageSendReply(m.ChannelID, "Oh, no! Something went wrong while processing your video", m.Reference())
			_ = s.MessageReactionAdd(m.ChannelID, m.ID, emojiIDErrorFFMPEG)
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
			URL: post.URL,
		}
	} else if post.IsEmbed {
		url, err := rb.embedder.Embed(&post)
		if err == embed.ErrorNotImplemented {
			logger.Warn().
				Err(err).
				Msg("embedded website (source) is not yet implemented")
		} else if err != nil {
			logger.Error().
				Err(err).
				Msg("something went wrong while analyzing embedded content")
		}
		_, _ = s.ChannelMessageSend(m.ChannelID, url)
		logger.Info().Msg("Sending embedded YouTube video")
	}

	_, err = s.ChannelMessageSendComplex(m.ChannelID, msg)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("Could not send embed")
		_, _ = s.ChannelMessageSendReply(m.ChannelID, "The video is too big", m.Reference())
		_ = s.MessageReactionAdd(m.ChannelID, m.ID, emojiIDTooBig)
		return
	}

	_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
}
