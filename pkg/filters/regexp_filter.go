package filters

import (
	"fmt"
	"regexp"
	"strings"
)

// This ILineFilter implementation is testing the line against a regular expression
type RegexpFilter struct {
	regexp regexp.Regexp
	negate bool
}

func NewRegexpFilter(expression string, negate bool) (filter RegexpFilter, failure error) {

	if strings.HasPrefix(expression, "/") {
		if strings.HasSuffix(expression, "/i") {
			// we need a case insensitive regexp!
			expression = expression[1 : len(expression)-2]
			expression = fmt.Sprintf("(?i)%v", expression)
		} else {
			failure = fmt.Errorf("invalid /.../ wrapping! To create a case-insensitive Regexp you need to wrap into '/.../i' but you provided: '%v'", expression)
			return
		}
	}
	regex, err := regexp.Compile(expression)
	if err != nil {
		failure = fmt.Errorf("the value '%v' does not seem to be a valid pattern, error: %v", expression, err)
		return
	}

	filter.regexp = *regex
	filter.negate = negate

	return
}

func (filter RegexpFilter) EvaluateLine(line string) (isMatch bool) {
	isMatch = filter.regexp.Match([]byte(line))
	if filter.negate {
		isMatch = !isMatch
	}
	return
}
