package main

import (
	"fmt"
	"github.com/haveachin/reddit-bot/regex"
)

// prepare regex
const patternYT = `(?s).*https:\/\/(?:www\.)youtube\.com\/embed\/(?P<%s>.+?)[\?\\\/].*`
const id = "id"
var p = regex.MustCompile(pattern, id)

func generateYouTubeURL(s string) (string, error) {
	// match against html and fetch video id
	m, err := p.FindStringSubmatch(s)
	if err != nil {
		return "", err
	}
	vid := m.CaptureByName(id)

	// generate url that discord can display
	return fmt.Sprintf(`https://www.youtube.com/watch?v=%s`, vid), nil
}