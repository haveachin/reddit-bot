package reddit

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Post is a very simplified variation of a JSON response given from the reddit api
type Post struct {
	// Title is the title of the Post
	Title string
	// Subreddit is the name of the subreddit that this post originated from
	Subreddit string
	// Author is the name of the user that posted this post
	Author string
	// Permalink is the permanent URL for this post
	Permalink string
	// ImageURL is the URL to the image form the post
	ImageURL string
	// VideoURL is the URL to the video from the post
	VideoURL string
	// IsVideo determines if the post is a video or an image
	IsVideo bool
}

type postJSON []struct {
	Data struct {
		Children []struct {
			Data struct {
				Title     string `json:"title"`
				Subreddit string `json:"subreddit"`
				Author    string `json:"author"`
				Permalink string `json:"permalink"`
				URL       string `json:"url"`
				IsVideo   bool   `json:"is_video"`
				Media     struct {
					Video struct {
						URL string `json:"dash_url"`
					} `json:"reddit_video"`
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

	return Post{
		Title:     postJSON[0].Data.Children[0].Data.Title,
		Subreddit: postJSON[0].Data.Children[0].Data.Subreddit,
		Author:    postJSON[0].Data.Children[0].Data.Author,
		Permalink: postJSON[0].Data.Children[0].Data.Permalink,
		ImageURL:  postJSON[0].Data.Children[0].Data.URL,
		VideoURL:  postJSON[0].Data.Children[0].Data.Media.Video.URL,
		IsVideo:   postJSON[0].Data.Children[0].Data.IsVideo,
	}, nil
}
