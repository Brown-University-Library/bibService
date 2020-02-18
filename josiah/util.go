package josiah

import (
	"encoding/json"
	"fmt"
	"log"
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
