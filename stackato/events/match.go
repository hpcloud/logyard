// provides fast matching algorithms
package main

import (
	"regexp"
	"strings"
)

// MultiRegexpMatch allows matching a string against multiple regular
// expressions *after* doing a simple O(n) substring match using
// aho-corasick. TODO: actually use aho-corasick.
type MultiRegexpMatcher struct {
	substrings map[string]string
	regexps    map[string]*regexp.Regexp
}

func NewMultiRegexpMatcher() *MultiRegexpMatcher {
	return &MultiRegexpMatcher{make(map[string]string), make(map[string]*regexp.Regexp)}
}

func (m *MultiRegexpMatcher) MustAdd(name string, substring string, re string) {
	if _, ok := m.substrings[name]; ok {
		panic("already in substrings")
	}
	if _, ok := m.regexps[name]; ok {
		panic("already in regexps")
	}
	m.substrings[name] = substring
	m.regexps[name] = regexp.MustCompile(re)
}

// Match tries to match the text against one of the substring/regexp
// as efficiently as possible. 
func (m *MultiRegexpMatcher) Match(text string) (string, []string) {
	// TODO: use aho-corasick instead of looping
	for name, substring := range m.substrings {
		if strings.Contains(text, substring) {
			if re, ok := m.regexps[name]; ok {
				return name, re.FindStringSubmatch(text)
			}
			panic("not in regexps")
		}
	}
	return "", nil
}
