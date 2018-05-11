package marc

import (
	"regexp"
	"strconv"
	"strings"
)

func in(values []string, searchedFor string) bool {
	for _, value := range values {
		if value == searchedFor {
			return true
		}
	}
	return false
}

func safeAppend(values *[]string, value string) {
	if value == "" {
		return
	}
	trimedValue := strings.TrimSpace(value)
	if !in(*values, trimedValue) {
		*values = append(*values, trimedValue)
	}
}

func toInt(str string) int {
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return int(num)
}

func TrimPunct(str string) string {
	if str == "" {
		return str
	}

	// RegEx stolen from Traject's marc21.rb
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21.rb
	//
	// # trailing: comma, slash, semicolon, colon (possibly preceded and followed by whitespace)
	// str = str.sub(/ *[ ,\/;:] *\Z/, '')
	re1 := regexp.MustCompile(" *[ ,\\/;:] *$")
	cleanStr := re1.ReplaceAllString(str, "")

	// # trailing period if it is preceded by at least three letters (possibly preceded and followed by whitespace)
	// str = str.sub(/( *\w\w\w)\. *\Z/, '\1')
	re2 := regexp.MustCompile("( *\\w\\w\\w)\\. *$")
	cleanStr = re2.ReplaceAllString(cleanStr, "$1")

	// # single square bracket characters if they are the start
	// # and/or end chars and there are no internal square brackets.
	// str = str.sub(/\A\[?([^\[\]]+)\]?\Z/, '\1')
	re3 := regexp.MustCompile("^\\[?([^\\[\\]]+)\\]?$")
	cleanStr = re3.ReplaceAllString(cleanStr, "$1")

	return cleanStr
}
