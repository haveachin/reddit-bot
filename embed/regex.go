package embed

import "github.com/haveachin/reddit-bot/regex"

const patternID = "id"

// regex pattern of the sources
const(
	patternYoutube = `(?s).*https:\/\/(?:www\.)youtube\.com\/embed\/(?P<%s>.+?)[\?\\\/].*`
)

// compiled patterns
var (
	pYoutube = regex.MustCompile(patternYoutube, patternID)
)


