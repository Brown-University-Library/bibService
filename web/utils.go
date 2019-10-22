package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func prettyJSON(jsonBytes []byte) string {
	var buffer bytes.Buffer
	err := json.Indent(&buffer, jsonBytes, "", "\t")
	if err != nil {
		return string(jsonBytes)
	}
	return string(buffer.Bytes())
}

func rangeFromDays(days int) (string, string) {
	dayDuration := 24 * time.Hour
	toDate := time.Now()
	fromDate := toDate.Add(time.Duration(-days) * dayDuration)
	from := fmt.Sprintf("%s", fromDate.String()[0:10])
	to := fmt.Sprintf("%s", toDate.String()[0:10])
	return from, to
}

func qsParam(name string, req *http.Request) string {
	params := req.URL.Query()
	if len(params[name]) > 0 {
		return params[name][0]
	}
	return ""
}

func qsParamInt(name string, req *http.Request) int {
	str := qsParam(name, req)
	num, _ := strconv.Atoi(str)
	return num
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
