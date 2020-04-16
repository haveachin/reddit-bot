package regex

import "fmt"

// Match represents a (sub)match as defined by the go regexp package.
// It stores a pointer to the underlying regex pattern as well as the capture groups' contents
type Match struct {
	Pattern *Pattern
	groups  []string
}

// Capture returns an individual capture group of the given Match m in the
// order of their declaration in the pattern of m.Pattern.
// Beware that m.Capture(0) returns the entire match, so m.Capture(1) will
// return the first capture group of the match and so forth
func (m *Match) Capture(index int) string {
	return m.groups[index]
}

// CaptureByName will return the capture group of the given Match m
// with the identifier name as specified in the pattern m.Pattern.
// This method will panic if name is not a valid name of any capture group of m.Pattern
func (m *Match) CaptureByName(name string) string {
	names := m.Pattern.regex.SubexpNames()
	for i, group := range m.groups {
		if names[i] == name {
			return group
		}
	}
	panic(fmt.Errorf("regex: no capture group with the provided alias name '%s' could be found", name))
}
