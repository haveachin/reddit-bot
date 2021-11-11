package reddit

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type PostType string

const (
	PostTypeImage       PostType = "image"
	PostTypeVideoHosted PostType = "hosted:video"
	PostTypeVideoEmbed  PostType = "rich:video"
	PostTypeSelf        PostType = "self"
)

// Post is a very simplified variation of a JSON response given from the reddit api
type Post struct {
	Title     string
	Text      string
	Subreddit string
	Author    string
	Permalink string
	ImageURL  string
	IsImage   bool
	IsVideo   bool
	IsEmbed   bool
	Video     Video
	Embed     Embed
}

type Embed struct {
	HTML string
	Type string
} // html embedded media

type Video struct {
	VideoURL string
	AudioURL string
}

type postJSON []struct {
	Data struct {
		Children []struct {
			Data struct {
				Title     string `json:"title"`
				Text      string `json:"selftext"`
				Subreddit string `json:"subreddit"`
				Author    string `json:"author"`
				Permalink string `json:"permalink"`
				ImageURL  string `json:"url"`
				PostHint  string `json:"post_hint"`
				IsVideo   bool   `json:"is_video"`
				Media     struct {
					Type  string `json:"type"`
					Video struct {
						URL string `json:"fallback_url"`
					} `json:"reddit_video"`
					Oembed struct {
						HTML string `json:"html"`
					} `json:"oembed"`
				} `json:"media"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// PostByID fetches the post with the corresponding ID
// A post ID is normally six characters long
func PostByID(postID string) (Post, error) {
	const apiPostURLf string = "https://www.reddit.com/%s/.json"
	url := fmt.Sprintf(apiPostURLf, postID)

	resp, err := defaultClient.Get(url)
	if err != nil {
		return Post{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return Post{}, ErrBadResponse
	}

	postJSON := postJSON{}

	if err := json.NewDecoder(resp.Body).Decode(&postJSON); err != nil {
		return Post{}, err
	}

	if len(postJSON) <= 0 {
		return Post{}, ErrBadResponse
	}

	if len(postJSON[0].Data.Children) <= 0 {
		return Post{}, ErrBadResponse
	}

	data := postJSON[0].Data.Children[0].Data
	isImage := data.PostHint == string(PostTypeImage)
	isEmbed := data.PostHint == string(PostTypeVideoEmbed) // TODO: change this
	isImage = isImage || (!isEmbed && !data.IsVideo && data.ImageURL != "")
	return Post{
		Title:     data.Title,
		Text:      data.Text,
		Subreddit: data.Subreddit,
		Author:    data.Author,
		Permalink: data.Permalink,
		ImageURL:  data.ImageURL,
		IsImage:   isImage,
		IsVideo:   data.IsVideo,
		IsEmbed:   isEmbed,
		Video: Video{
			VideoURL: data.Media.Video.URL,
			AudioURL: data.ImageURL + "/DASH_audio.mp4",
		},
		Embed: Embed{
			HTML: data.Media.Oembed.HTML,
			Type: data.Media.Type,
		},
	}, nil
}
