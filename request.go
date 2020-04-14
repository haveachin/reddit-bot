package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const apiURL string = "https://www.reddit.com/%s/.json"

type RedditPostJSON []struct {
	Data struct {
		Children []struct {
			Data struct {
				Title     string `json:"title"`
				Subreddit string `json:"subreddit_name_prefixed"`
				Author    string `json:"author"`
				Permalink string `json:"permalink"`
				URL       string `json:"url"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type redditPost struct {
	title     string
	subreddit string
	author    string
	permalink string
	imageURL  string
}

func getPostData(postID string) (*redditPost, error) {
	redditPostJSON := RedditPostJSON{}

	url := fmt.Sprintf(apiURL, postID)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-agent", "BonoBanani")

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if err := json.NewDecoder(resp.Body).Decode(&redditPostJSON); err != nil {
		return nil, err
	}

	if len(redditPostJSON) <= 0 {
		return nil, errors.New("bad response")
	}

	if len(redditPostJSON[0].Data.Children) <= 0 {
		return nil, errors.New("bad response")
	}

	return &redditPost{
		title:     redditPostJSON[0].Data.Children[0].Data.Title,
		subreddit: redditPostJSON[0].Data.Children[0].Data.Subreddit,
		author:    redditPostJSON[0].Data.Children[0].Data.Author,
		permalink: redditPostJSON[0].Data.Children[0].Data.Permalink,
		imageURL:  redditPostJSON[0].Data.Children[0].Data.URL,
	}, nil
}
