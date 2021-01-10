package regex

import (
	"fmt"
	"regexp"
	"strings"
)

// Pattern represents a regex-pattern similar to regexp.Regexp
// and essentially works in the same way.
type Pattern struct {
	regex *regexp.Regexp
}

// Regex returns the underlying *regexp.Regexp of p.
// The return value should NOT be modified as it changes in p respectively.
func (p *Pattern) Regex() *regexp.Regexp {
	return p.regex
}

// MustCompile returns a new Pattern p with an underlying *regexp.Regexp that
// is created as specified by regexp.MustCompile(..).
// This method allows for appending the names of the capture groups after pattern.
// To refer to a capture group name in the pattern use %s as specified by e.g. fmt.Sprintf(..)
func MustCompile(pattern string, names ...string) Pattern {
	s := pattern
	for _, v := range names {
		s = strings.Replace(s, "%s", v, 1)
	}

	re := regexp.MustCompile(s)
	return Pattern{re}
}

// FindStringSubmatch works like *regexp.Regexp.FindStringSubmatch(...)
// but returns an error if s can't be matched against the pattern p
func (p *Pattern) FindStringSubmatch(s string) (Match, error) {
	if !p.regex.MatchString(s) {
		return Match{}, fmt.Errorf("regex: no matches found in provided string s: %s", s)
	}

	return Match{p, p.regex.FindStringSubmatch(s)}, nil
}
