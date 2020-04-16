package regex

import "errors"

var (
	// ErrNoSuchCaptureGroup should be thrown when a Match doesn't contain a Pattern with the given capture gorup name.
	// It can be additionally supplemented with the faulty name, when used with a fmt.Errorf call
	ErrNoSuchCaptureGroup error = errors.New("no capture group with the provided alias name '%s' could be found")
	// ErrNoMatch should be thrown when the Pattern doesn't match to the provided string.
	// It can be additionally supplemented with the faulty string, when used with a fmt.Errorf call
	ErrNoMatch error = errors.New("no matches found in provided string s: %s")
)
