package main

import (
	"fmt"
	"regexp"
)

// Match represents a (sub)match as defined by the go regexp package.
// It stores a pointer to the underlying regex pattern as well as the capture groups' contents
type Match struct {
	Regex  *regexp.Regexp
	groups []string
}

// FindStringSubmatch works like *regexp.Regexp.FindStringSubmatch(...)
// but returns an error if s can't be matched against the pattern re
func FindStringSubmatch(re *regexp.Regexp, s string) (Match, error) {
	if !re.MatchString(s) {
		return Match{}, fmt.Errorf("no matches found in provided string s: %s", s)
	}

	return Match{re, re.FindStringSubmatch(s)}, nil
}

// Capture returns an individual capture group of the given Match m in the
// order of their declaration in the pattern of m.Regex.
// Beware that m.Capture(0) returns the entire match, so m.Capture(1) will
// return the first capture group of the match and so forth
func (m Match) Capture(index int) string {
	return m.groups[index]
}

// CaptureByName will return the capture group of the given Match m
// with the identifier name as specified in the pattern m.Regex.
// This method will panic if name is not a valid name of any capture group of m.Regex
func (m Match) CaptureByName(name string) string {
	names := m.Regex.SubexpNames()
	for i, group := range m.groups {
		if names[i] == name {
			return group
		}
	}
	panic(fmt.Sprintf("no capture group with the provided alias name '%s' could be found", name))
}
