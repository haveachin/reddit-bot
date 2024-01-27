package reddit

import (
	"net/http"
	"time"
)

const cientTimeout = time.Second * 10

var defaultClient = http.Client{
	Transport: &headerTransport{},
	Timeout:   cientTimeout,
}

type headerTransport struct{}

func (ht *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "RedditBot (https://github.com/haveachin/reddit-bot)")
	return http.DefaultTransport.RoundTrip(req)
}
