package sierra

import (
	"encoding/json"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func arrayToPages(values []string, pageSize int) [][]string {
	pages := [][]string{}
	total := len(values)
	pageCount := total / pageSize
	if math.Mod(float64(total), float64(pageSize)) != 0 {
		pageCount += 1
	}
	for i := 0; i < pageCount; i++ {
		start := i * pageSize
		end := start + pageSize
		if end > total {
			end = total
		}
		log.Printf("from: %d to %d", start, end)
		page := values[start:end]
		pages = append(pages, page)
	}
	return pages
}

func toIntTry(str string) (int, bool) {
	num, err := strconv.ParseInt(str, 10, 64)
	return int(num), err == nil
}

func toInt(str string) int {
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return int(num)
}

func index(values []string, searchedFor string) int {
	for i := 0; i < len(values); i++ {
		if values[i] == searchedFor {
			return i
		}
	}
	return -1
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

func arrayAppend(values *[]string, newValues []string) {
	for _, newValue := range newValues {
		safeAppend(values, newValue)
	}
}

func addPeriod(value string) string {
	if value == "" || strings.HasSuffix(value, ".") || strings.HasSuffix(value, ")") {
		return value
	}
	return value + "."
}

func trimDot(str string) string {
	if str == "" {
		return str
	}
	if strings.HasSuffix(str, ".") {
		return str[0:(len(str) - 1)]
	}
	return str
}

func trimPunct(str string) string {
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

// This is a hack to try to achieve the same items that Traject is inserting
// based on the MARC data. We might not need this in the future
func dedupArray(original []string) []string {
	dedup := []string{}
	for _, value := range original {
		trimVal := trimDot(value)
		if trimVal == value {
			// the value ("a") and the trimmed ("a") version of the value are the same
			// add it to the array if it is not already there.
			safeAppend(&dedup, value)
		} else {
			// the value ("a.") is different from the trimmed version ("a")
			indexTrim := index(dedup, trimVal)
			if indexTrim >= 0 {
				// if the trimmed version ("a") is already in the array,
				// replace it the not trimmed version ("a.")
				dedup[indexTrim] = value
			} else {
				// add the not trimmed version to the array if it is not already there
				safeAppend(&dedup, value)
			}
		}
	}
	return dedup
}

func toJSON(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func stringToArray(str string) []string {
	array := []string{}
	for _, c := range str {
		array = append(array, string(c))
	}
	return array
}
