package web

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func prettyJSON(jsonBytes []byte) string {
	var buffer bytes.Buffer
	err := json.Indent(&buffer, jsonBytes, "", "\t")
	if err != nil {
		return string(jsonBytes)
	}
	return string(buffer.Bytes())
}

// func isRaw(values map[string][]string) bool {
// 	for _, value := range values["raw"] {
// 		if value == "true" {
// 			return true
// 		}
// 	}
// 	return false
// }
//
// // Extracts the BIB from a URL Path. Assumes the BIB is the last segment
// // of the path. For example: /whatever/whatever/bib
// func bibFromPath(path string) string {
// 	tokens := strings.Split(path, "/")
// 	if len(tokens) == 0 {
// 		return ""
// 	}
// 	return tokens[len(tokens)-1]
// }

// func bibFromQs(req *http.Request) string {
// 	return qsParam("bib", req)
// }

func qsParam(name string, req *http.Request) string {
	params := req.URL.Query()
	if len(params[name]) > 0 {
		return params[name][0]
	}
	return ""
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
