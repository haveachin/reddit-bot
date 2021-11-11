package embed

import "github.com/haveachin/reddit-bot/reddit"

type Embedder interface {
	Embed(*reddit.Post) (string, error)
}
