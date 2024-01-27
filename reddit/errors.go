package reddit

import "errors"

var (
	// ErrBadResponse indicates a bad or unexpected response from the webapi.
	ErrBadResponse = errors.New("bad response")
)
