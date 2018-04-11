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
