package reddit

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type PostType string

const (
	// PostTypeImage means an image that is hosted on Reddit.
	PostTypeImage PostType = "image"
	// PostTypeVideoHosted is a video that is hosted on Reddit.
	PostTypeVideoHosted PostType = "hosted:video"
	// PostTypeVideoEmbed is a video that is hosted on an external platform.
	// Normally these videos are embedded via an iFrame.
	PostTypeVideoEmbed PostType = "rich:video"
	PostTypeSelf       PostType = "self"

	PostTypeRedGif   PostType = "redgifs.com"
	PostTypeRedGifV3 PostType = "v3.redgifs.com"
)

// Post is a very simplified variation of a JSON response given from the reddit api.
type Post struct {
	ID                 string
	Title              string
	Text               string
	Subreddit          string
	Author             string
	Permalink          string
	URL                string
	IsImage            bool
	IsVideo            bool
	IsEmbed            bool
	WasRemoved         bool
	Embed              Embed
	PostProcessingArgs []string
}

func (p Post) ShortURL() string {
	return "https://www.reddit.com/" + p.ID
}

type Embed struct {
	HTML string
	Type string
} // html embedded media

type postDTO []struct {
	Data struct {
		Children []struct {
			Data struct {
				Title             string `json:"title"`
				Text              string `json:"selftext"`
				Subreddit         string `json:"subreddit"`
				Author            string `json:"author"`
				Permalink         string `json:"permalink"`
				URL               string `json:"url"`
				PostHint          string `json:"post_hint"`
				IsVideo           bool   `json:"is_video"`
				RemovedByCategory string `json:"removed_by_category"`
				Media             struct {
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

func fetchPost(postID string) (postDTO, error) {
	const apiPostURLf string = "https://www.reddit.com/%s/.json"
	url := fmt.Sprintf(apiPostURLf, postID)

	for i := 3; i > 0; i-- {
		resp, err := defaultClient.Get(url)
		if err != nil {
			return postDTO{}, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}

		dto := postDTO{}
		if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
			return postDTO{}, err
		}
		resp.Body.Close()

		return dto, nil
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
	resp.Body.Close()

	return resp.Request.URL.String(), nil
}

// PostByID fetches the post with the corresponding ID.
// A post ID is normally six characters long.
func PostByID(postID string) (Post, error) {
	dto, err := fetchPost(postID)
	if err != nil {
		return Post{}, err
	}

	if len(dto) == 0 {
		return Post{}, ErrBadResponse
	}

	if len(dto[0].Data.Children) == 0 {
		return Post{}, ErrBadResponse
	}

	data := dto[0].Data.Children[0].Data
	isVideo := data.IsVideo ||
		data.Media.Type == string(PostTypeRedGif) ||
		data.Media.Type == string(PostTypeRedGifV3)
	isEmbed := data.PostHint == string(PostTypeVideoEmbed) // TODO: change this
	isImage := data.PostHint == string(PostTypeImage) ||
		(!isEmbed && !isVideo && data.URL != "")
	wasRemoved := data.RemovedByCategory != "" && data.URL == ""
	return Post{
		ID:         postID,
		Title:      data.Title,
		Text:       data.Text,
		Subreddit:  data.Subreddit,
		Author:     data.Author,
		Permalink:  data.Permalink,
		URL:        data.URL,
		IsImage:    isImage,
		IsVideo:    isVideo,
		IsEmbed:    isEmbed,
		WasRemoved: wasRemoved,
		Embed: Embed{
			HTML: data.Media.Oembed.HTML,
			Type: data.Media.Type,
		},
	}, nil
}
