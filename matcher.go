package genlib

import (
	"fmt"
	"regexp"
	"strings"
)

type Matcher interface {
	fmt.Stringer

	Match(input string) bool
}

// ---

type listMatcher struct {
	matchers []Matcher
}

func (m *listMatcher) Match(input string) bool {
	for _, matcher := range m.matchers {
		if matcher.Match(input) {
			return true
		}
	}
	return false
}

func (m *listMatcher) String() string {
	strs := make([]string, len(m.matchers))
	for i, matcher := range m.matchers {
		strs[i] = matcher.String()
	}
	return strings.Join(strs, ", ")
}

func NewListMatcher(matchers ...Matcher) Matcher {
	return &listMatcher{matchers: matchers}
}

// ---

type stringMatcher struct {
	string     string
	ignoreCase bool
}

func (n *stringMatcher) Match(input string) bool {
	if n.ignoreCase {
		return strings.ToLower(n.string) == strings.ToLower(input)
	}
	return n.string == input
}

func (n *stringMatcher) String() string {
	if n.ignoreCase {
		return fmt.Sprintf(`match "%v" ignore case`, n.string)
	}
	return fmt.Sprintf(`match "%v" case sensitive`, n.string)
}

func NewStringMatcher(name string, ignoreCase bool) Matcher {
	return &stringMatcher{
		string:     name,
		ignoreCase: ignoreCase,
	}
}

// ---

type regexMatcher struct {
	pattern *regexp.Regexp
	regex   string
}

func (r *regexMatcher) Match(input string) bool {
	if r.pattern == nil {
		pattern, err := regexp.Compile(r.regex)
		if err != nil {
			return false
		}
		r.pattern = pattern
	}
	return r.pattern.MatchString(input)
}

func (r *regexMatcher) String() string {
	return fmt.Sprintf(`match by regex "%v"`, r.regex)
}

func NewRegexMatcher(regex string) (Matcher, error) {
	pattern, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	return &regexMatcher{pattern: pattern, regex: regex}, nil
}

var _ Matcher = (*listMatcher)(nil)
var _ Matcher = (*stringMatcher)(nil)
var _ Matcher = (*regexMatcher)(nil)
