// provides fast matching algorithms
// TODO: aho-corasic on substring matching
package events

import (
	"github.com/ActiveState/log"
	"regexp"
	"strings"
)

// MultiRegexpMatch allows matching a string against multiple regular
// expressions along with substrings for a fast fail-early matching.
type MultiRegexpMatcher struct {
	substrings       map[string]string         // substring to name
	regexps          map[string]*regexp.Regexp // name to regexp
	substringsRegexp *regexp.Regexp            // substring regex combined
}

func NewMultiRegexpMatcher() *MultiRegexpMatcher {
	return &MultiRegexpMatcher{
		make(map[string]string),
		make(map[string]*regexp.Regexp),
		nil}
}

func (m *MultiRegexpMatcher) MustAdd(name string, substring string, re string) {
	if oldName, ok := m.substrings[substring]; ok {
		log.Fatalf(
			"substring %s already added under %s; being added again by %s",
			substring, oldName, name)
	}
	if _, ok := m.regexps[name]; ok {
		log.Fatal("already in regexps")
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
		// fail return early so we don't have to waste time on futile regex
		// matching (below)
		return "", nil
	}

	if name, ok := m.substrings[substring]; ok {
		if re, ok := m.regexps[name]; ok {
			// TODO: if this regex fails, should we try the next
			// matching substring?
			return name, re.FindStringSubmatch(text)
		}
	}
	panic("not reachable")
}
