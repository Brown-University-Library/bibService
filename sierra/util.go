package sierra

import (
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

func valuesToArray(values [][]string, trim bool, join bool) []string {
	array := []string{}
	for _, fieldValues := range values {
		if join {
			// join all the values for the field as a single element in
			// the returning array
			value := strings.Join(fieldValues, " ")
			if trim {
				value = trimPunct(value)
			}
			safeAppend(&array, value)
		} else {
			// preserve each individual value (regardless of their field)
			// as a single element in the returning arrray
			for _, value := range fieldValues {
				if trim {
					value = trimPunct(value)
				}
				safeAppend(&array, value)
			}
		}
	}
	return array
}

func valuesToString(values [][]string, trim bool) string {
	rowValues := []string{}
	for _, fieldValues := range values {
		safeAppend(&rowValues, strings.Join(fieldValues, " "))
	}
	if trim {
		return trimPunct(strings.Join(rowValues, " "))
	}
	return strings.Join(rowValues, " ")
}
