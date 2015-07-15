package scopes

import (
	"strings"
)

type (
	Given    []string
	Required [][]string
)

// Calls function `matches(j)` for each `j` in `list` until `true` is returned,
// or the list is exhausted. Returns `true` if result `true` was returned from
// `matches(j)` for some `j`, otherwise false.
func some(list []interface{}, matches func(interface{}) bool) bool {
	for _, j := range list {
		if matches(j) {
			return true
		}
	}
	return false
}

// Calls function `matches(j)` for each `j` in `list` until `false` is returned,
// or the list is exhausted. Returns `false` if result `false` was returned from
// `matches(j)` for some `j`, otherwise true.
func every(list []interface{}, matches func(interface{}) bool) bool {
	for _, j := range list {
		if !matches(j) {
			return false
		}
	}
	return true
}

// Returns true if `given` satisfies `required`.
//
// `given` is an array of strings such as:
//
//  given := []string{
//  	"abc:*",
//  	"123:4:56",
//  	"xyz",
//  	"AB:*",
//  }
//
// `required` is an array of arrays of strings such as:
//
//  required := [][]string{
//  	{"abc:def", "AB:CD:EF"},
//  	{"123:4:5"},
//  	{"abc:def", "123:4"},
//  	{"Xxyz"},
//  }
//
// 1) The `*` when specified at the end of a scope in `given` operates as a
// wildcard, matching anything.
//
// 2) Each string of inner []string of `required` must be satisfied for the
// []string to be satisfied.  `required` is satisfied if only one of the outer
// []string of the [][]string is satisfied. In other words, outer array =>
// logical OR, inner array => logical AND.
//
// In the example above, ScopeMatch would return false, since the scopes are not
// satisfied.
func (given *Given) Satisfies(required *Required) bool {
	requiredI := make([]interface{}, len(*required))
	for i, d := range *required {
		requiredI[i] = d
	}
	return some(requiredI, func(scopeSet interface{}) bool {
		scopeSetI := make([]interface{}, len(scopeSet.([]string)))
		for i, d := range scopeSet.([]string) {
			scopeSetI[i] = d
		}
		return every(scopeSetI, func(scope interface{}) bool {
			scopeString := scope.(string)
			givenI := make([]interface{}, len(*given))
			for i, d := range *given {
				givenI[i] = d
			}
			return some(givenI, func(pattern interface{}) bool {
				patternString := pattern.(string)
				if scopeString == patternString {
					return true
				}
				if strings.HasSuffix(patternString, "*") {
					return strings.HasPrefix(scopeString, patternString[0:len(patternString)-1])
				}
				return false
			})
		})
	})
}
