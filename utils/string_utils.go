package utils

import (
	"regexp"
	"strings"
)

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func RemoveEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		str = strings.TrimSpace(str)
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func WildCardToRegexp(pattern string) (bool, string) {
	var result strings.Builder
	isWildCard := false
	result.WriteString("^")
	for i, literal := range strings.Split(pattern, "*") {

		if i > 0 {
			isWildCard = true
			result.WriteString(".*")
		}

		result.WriteString(regexp.QuoteMeta(literal))
	}
	result.WriteString("$")
	return isWildCard, result.String()
}

func MatchAtLeastOne(expectedValues []string, value string) bool {
	for _, expectedValue := range expectedValues {
		isWildCard, pattern := WildCardToRegexp(expectedValue)
		if isWildCard {
			if matched, _ := regexp.MatchString(pattern, value); matched {
				return true
			}
		} else if value == expectedValue {
			return true
		}
	}
	return false
}
