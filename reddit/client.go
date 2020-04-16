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
	const (
		headerKeyUserAgent   string = "User-Agent"
		headerValueUserAgent string = "Haveachins-Reddit-Bot"
	)

	req.Header.Add(headerKeyUserAgent, headerValueUserAgent)
	return http.DefaultTransport.RoundTrip(req)
}
