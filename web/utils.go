package web

import (
	"bytes"
	"encoding/json"
	"strings"
)

func prettyJSON(jsonBytes []byte) string {
	var buffer bytes.Buffer
	err := json.Indent(&buffer, jsonBytes, "", "\t")
	if err != nil {
		return string(jsonBytes)
	}
	return string(buffer.Bytes())
}

// Extracts the BIB from a URL Path. Assumes the BIB is the last segment
// of the path. For example: /whatever/whatever/bib
func bibFromPath(path string) string {
	tokens := strings.Split(path, "/")
	if len(tokens) == 0 {
		return ""
	}
	return tokens[len(tokens)-1]
}

func toJSON(data interface{}, pretty bool) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	if pretty {
		return prettyJSON(bytes), nil
	}
	return string(bytes), nil
}
