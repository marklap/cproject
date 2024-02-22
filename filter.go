package cproject

import "strings"

// Filter describes the behavior of a log file filter.
type Filter interface {
	// Include returns true if the provided line should be included in the results.
	Include(string) bool
}

// MatchAnySubstring is a filter that checks that a line contains at least one of a slice of substrings. The filter
// can be configured to ignore case.
type MatchAnySubstring struct {
	substrings    []string
	caseSensitive bool
}

type matchAnySubstringOpt func(*MatchAnySubstring)

// NewMatchAnySubstring creates a new substring match filter given the provided options.
func NewMatchAnySubstring(opts ...matchAnySubstringOpt) *MatchAnySubstring {
	f := &MatchAnySubstring{
		caseSensitive: true,
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

// WithSubstrings adds the provided slice of strings to the filter; each of which will be checked for a match against
// a line of text - first match stops checking.
func WithSubstrings(substrings []string) func(*MatchAnySubstring) {
	return func(f *MatchAnySubstring) {
		f.substrings = substrings
	}
}

// WithCaseSensitivity sets the case sensitivity flag of the filter.
func WithCaseSensitivity(b bool) func(*MatchAnySubstring) {
	return func(f *MatchAnySubstring) {
		f.caseSensitive = b
	}
}

// Include determines if a line of text should be included in the result set.
func (f *MatchAnySubstring) Include(s string) bool {
	if !f.caseSensitive {
		s = strings.ToLower(s)
	}
	for _, ss := range f.substrings {
		if !f.caseSensitive {
			ss = strings.ToLower(ss)
		}
		if strings.Contains(s, ss) {
			return true
		}
	}
	return false
}
