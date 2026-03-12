package client

import "regexp"

// Result holds named capture groups from a regex match, or arbitrary
// key/value data returned by a custom MatcherFunc.
type Result map[string]string

// Matcher matches an incoming Message and returns a Result if it matches,
// or nil if it does not.
type Matcher interface {
	Match(Message) Result
}

// MatcherFunc wraps a plain function as a Matcher.
type MatcherFunc func(Message) Result

func (fn MatcherFunc) Match(msg Message) Result {
	return fn(msg)
}

// RegexMatcher matches messages against one or more regular expressions.
// Named capture groups are available in the Result.
type RegexMatcher struct {
	expressions []*regexp.Regexp
}

// NewRegexMatcher returns a RegexMatcher compiled from the provided expressions.
// Panics if any expression is invalid.
func NewRegexMatcher(expressions ...string) RegexMatcher {
	r := RegexMatcher{}
	for _, e := range expressions {
		r.expressions = append(r.expressions, regexp.MustCompile(e))
	}
	return r
}

func (r RegexMatcher) Match(msg Message) Result {
	res := make(Result)
	for _, expr := range r.expressions {
		match := expr.FindStringSubmatch(msg.Payload)
		if len(match) == 0 {
			continue
		}
		for i, name := range expr.SubexpNames() {
			res[name] = match[i]
		}
		return res
	}
	return nil
}
