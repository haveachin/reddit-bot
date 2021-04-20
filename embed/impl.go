package embed

import (
	"fmt"
	"github.com/haveachin/reddit-bot/reddit"
)

type embedderImpl struct {
}

func NewEmbedder() Embedder {
	return embedderImpl{}
}

func (e embedderImpl) Embed(p *reddit.Post) (string, error) {
	// determine source and get fitting matcher
	s := Source(p.Embed.Type)
	m, err := newMatcher(s)
	if err != nil {
		return "", err
	}

	// get id of video
	id, err := m.fetchID(p.Embed.HTML)
	if err != nil {
		return "", err
	}

	// return the link to be posted to discord
	return fmt.Sprintf(m.urlTmpl, id), nil
}