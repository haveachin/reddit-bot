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
	PostTypeRedGif      PostType = "redgifs.com"
)

// Post is a very simplified variation of a JSON response given from the reddit api
type Post struct {
	ID        string
	Title     string
	Text      string
	Subreddit string
	Author    string
	Permalink string
	ImageURL  string
	IsImage   bool
	IsVideo   bool
	IsEmbed   bool
	Embed     Embed
}

func (p Post) URL() string {
	return "https://www.reddit.com/" + p.ID
}

type Embed struct {
	HTML string
	Type string
} // html embedded media

type postJSON []struct {
	Data struct {
		Children []struct {
			Data struct {
				Title     string `json:"title"`
				Text      string `json:"selftext"`
				Subreddit string `json:"subreddit"`
				Author    string `json:"author"`
				Permalink string `json:"permalink"`
				URL       string `json:"url"`
				PostHint  string `json:"post_hint"`
				IsVideo   bool   `json:"is_video"`
				Media     struct {
					Type   string `json:"type"`
					Oembed struct {
						HTML string `json:"html"`
					} `json:"oembed"`
				} `json:"media"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type LinkType string

const (
	ShareLinkType    = LinkType("s")
	CommentsLinkType = LinkType("comments")
)

func fetchPost(postID string) (*http.Response, error) {
	const apiPostURLf string = "https://www.reddit.com/%s/.json"
	url := fmt.Sprintf(apiPostURLf, postID)

	for i := 3; i > 0; i-- {
		resp, err := defaultClient.Get(url)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			continue
		}

		return resp, nil
	}
	return nil, ErrBadResponse
}

func ResolvePostURLFromShareID(subreddit, shareID string) (string, error) {
	const shareURLf string = "https://www.reddit.com/r/%s/s/%s"
	url := fmt.Sprintf(shareURLf, subreddit, shareID)
	resp, err := defaultClient.Get(url)
	if err != nil {
		return "", err
	}

	return resp.Request.URL.String(), nil
}

// PostByID fetches the post with the corresponding ID
// A post ID is normally six characters long
func PostByID(postID string) (Post, error) {
	resp, err := fetchPost(postID)
	if err != nil {
		return Post{}, err
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
	isVideo := data.IsVideo || data.Media.Type == string(PostTypeRedGif)
	isEmbed := data.PostHint == string(PostTypeVideoEmbed) // TODO: change this
	isImage := data.PostHint == string(PostTypeImage)
	isImage = isImage || (!isEmbed && !isVideo && data.URL != "")
	return Post{
		ID:        postID,
		Title:     data.Title,
		Text:      data.Text,
		Subreddit: data.Subreddit,
		Author:    data.Author,
		Permalink: data.Permalink,
		ImageURL:  data.URL,
		IsImage:   isImage,
		IsVideo:   isVideo,
		IsEmbed:   isEmbed,
		Embed: Embed{
			HTML: data.Media.Oembed.HTML,
			Type: data.Media.Type,
		},
	}, nil
}
