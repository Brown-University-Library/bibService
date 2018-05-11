package marc

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

var reTrailingPunct *regexp.Regexp
var reTrailingPeriod *regexp.Regexp
var reSquareBracket *regexp.Regexp

func init() {
	// RegEx stolen from Traject's marc21.rb
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21.rb
	//
	// # trailing: comma, slash, semicolon, colon (possibly preceded and followed by whitespace)
	// str = str.sub(/ *[ ,\/;:] *\Z/, '')
	reTrailingPunct = regexp.MustCompile(" *[ ,\\/;:] *$")

	// # trailing period if it is preceded by at least three letters (possibly preceded and followed by whitespace)
	// str = str.sub(/( *\w\w\w)\. *\Z/, '\1')
	reTrailingPeriod = regexp.MustCompile("( *\\w\\w\\w)\\. *$")

	// # single square bracket characters if they are the start
	// # and/or end chars and there are no internal square brackets.
	// str = str.sub(/\A\[?([^\[\]]+)\]?\Z/, '\1')
	reSquareBracket = regexp.MustCompile("^\\[?([^\\[\\]]+)\\]?$")
}

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

// Removes punctuation from a string. The algorithm to remove punctuation
// is tailored for common issues in MARC data.
func TrimPunct(str string) string {
	if str == "" {
		return str
	}

	cleanStr := reTrailingPunct.ReplaceAllString(str, "")
	cleanStr = reTrailingPeriod.ReplaceAllString(cleanStr, "$1")
	cleanStr = reSquareBracket.ReplaceAllString(cleanStr, "$1")
	return cleanStr
}

// Calculates the publication year from a value that is from the MARC 008
// in a MARC record.
func PubYear008(f008 string, tolerance int) (int, bool) {
	// Logic stolen from
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21_semantics.rb
	//
	// e.g. "760629c19749999ne tr pss o   0   a0eng  cas   "
	if len(f008) < 11 {
		return 0, false
	}

	dateType := f008[6:7]
	if dateType == "n" {
		// unknown
		return 0, false
	}

	var dateStr1, dateStr2 string
	dateStr1 = f008[7:11]
	if len(f008) >= 15 {
		dateStr2 = f008[11:15]
	} else {
		dateStr2 = dateStr1
	}

	if dateType == "q" {
		// questionable
		date1 := toInt(strings.Replace(dateStr1, "u", "0", -1))
		date2 := toInt(strings.Replace(dateStr2, "u", "9", -1))
		if (date2 > date1) && ((date2 - date1) <= tolerance) {
			return (date2 + date1) / 2, true
		} else {
			return 0, false
		}
	}

	var dateStr string
	if dateType == "p" {
		// use the oldest date
		if dateStr1 <= dateStr2 || toInt(dateStr2) == 0 {
			dateStr = dateStr1
		} else {
			dateStr = dateStr2
		}
	} else if dateType == "r" && toInt(dateStr2) != 0 {
		dateStr = dateStr2 // use the second date
	} else {
		dateStr = dateStr1 // use the first date
	}

	uCount := strings.Count(dateStr, "u")
	// should we replace with "9" if we pick dateStr2 ?
	date := toInt(strings.Replace(dateStr, "u", "0", -1))
	if uCount > 0 && date != 0 {
		delta := int(math.Pow10(uCount))
		if delta <= tolerance {
			return date + (delta / 2), true
		}
	} else if date != 0 {
		return date, true
	}

	return 0, false
}
