// provides fast matching algorithms
// TODO: aho-corasic on substring matching
package main

import (
	"regexp"
	"strings"
)

// MultiRegexpMatch allows matching a string against multiple regular
// expressions *after* doing a simple O(n) substring match using
// aho-corasick. 
type MultiRegexpMatcher struct {
	substrings       map[string]string
	regexps          map[string]*regexp.Regexp
	substringsRegexp *regexp.Regexp
}

func NewMultiRegexpMatcher() *MultiRegexpMatcher {
	return &MultiRegexpMatcher{
		make(map[string]string),
		make(map[string]*regexp.Regexp),
		nil}
}

func (m *MultiRegexpMatcher) MustAdd(name string, substring string, re string) {
	if _, ok := m.substrings[name]; ok {
		panic("already in substrings")
	}
	if _, ok := m.regexps[name]; ok {
		panic("already in regexps")
	}
	m.substrings[substring] = name
	m.regexps[name] = regexp.MustCompile(re)
}

func (m *MultiRegexpMatcher) Build() {
	escaped := make([]string, 0, len(m.substrings))
	for substring, _ := range m.substrings {
		escaped = append(escaped, regexp.QuoteMeta(substring))
	}
	m.substringsRegexp = regexp.MustCompile(strings.Join(escaped, "|"))
}

// Match tries to match the text against one of the substring/regexp
// as efficiently as possible. 
func (m *MultiRegexpMatcher) Match(text string) (string, []string) {
	// TODO: use aho-corasick instead of regexp to match the substrings.
	substring := m.substringsRegexp.FindString(text)
	if substring == "" {
		// return early so we don't have to waste time on futile regex
		// matching (below)
		return "", nil
	}

	if name, ok := m.substrings[substring]; ok {
		if re, ok := m.regexps[name]; ok {
			return name, re.FindStringSubmatch(text)
		}
	}
	panic("not reachable")
}
