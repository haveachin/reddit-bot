package reddit

import (
	"net/http"
	"time"
)

var defaultClient http.Client

func init() {
	defaultClient = http.Client{
		Transport: &headerTransport{},
		Timeout:   time.Second * 5,
	}
}

type headerTransport struct{}

func (ht *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "RedditBot (https://github.com/haveachin/reddit-bot)")
	return http.DefaultTransport.RoundTrip(req)
}
