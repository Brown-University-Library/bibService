package josiah

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

func ToJSON(data interface{}) string {
	bytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error converting to JSON: %s\r\n", err)
		return ""
	}
	return string(bytes)
}

func DbUtcNow() string {
	t := time.Now().UTC()
	s := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	return s
}

func Slugify(value string) string {
	slug := strings.Trim(value, " ")
	slug = strings.ToLower(slug)
	var chars []rune
	for _, c := range slug {
		isAlpha := c >= 'a' && c <= 'z'
		isDigit := c >= '0' && c <= '9'
		if isAlpha || isDigit {
			chars = append(chars, c)
		} else {
			chars = append(chars, '-')
		}
	}
	slug = string(chars)

	// remove double dashes
	for strings.Index(slug, "--") > -1 {
		slug = strings.Replace(slug, "--", "-", -1)
	}

	if len(slug) == 0 || slug == "-" {
		return ""
	}

	// make sure we don't end with a dash
	if slug[len(slug)-1] == '-' {
		return slug[0 : len(slug)-1]
	}
	return slug
}
