package embed

import (
	"github.com/haveachin/reddit-bot/regex"
)

type matcher struct {
	s Source
	p regex.Pattern
	urlTmpl string
}

// newMatcher returns a matcher for the provided source
func newMatcher(s Source) (matcher, error) {
	m := matcher{s: s}

	switch s {
	case Youtube:
		m.p = pYoutube
		m.urlTmpl = urlYoutube
	case Gfycat:
		m.p = pGfycat
		m.urlTmpl = urlGfycat
	default:
		return m, ErrorNotImplemented
	}

	return m, nil
}

// fetchID gets the ID of the embedded video
func (mchr matcher) fetchID(s string) (string, error) {
	// match against html and fetch video id
	m, err := mchr.p.FindStringSubmatch(s)
	if err != nil {
		return "", err
	}
	return m.CaptureByName(patternID), nil
}
