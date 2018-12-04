package sierra

// Sierra API documentation:
//	https://sandbox.iii.com/iii/sierra-api/swagger/index.html
// 	https://techdocs.iii.com/sierraapi/Content/zReference/objects/bibObjectDescription.htm
// 	https://techdocs.iii.com/sierraapi/Content/zAppendix/bibObjectExample.htm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Sierra represents the Sierra API endpoint.
type Sierra struct {
	URL           string
	Persistent    bool
	KeySecret     string
	KeySecret64   string
	Authorization authResp
	Verbose       bool
	SessionFile   string
}

// NewSierra defines a Sierra API endpoint.
func NewSierra(apiURL, keySecret, sessionFile string) Sierra {
	s := Sierra{
		URL:         apiURL,
		KeySecret:   keySecret,
		KeySecret64: base64.StdEncoding.EncodeToString([]byte(keySecret)),
		SessionFile: sessionFile,
		Persistent:  (sessionFile != ""),
		Verbose:     false,
	}

	if s.Persistent {
		s.loadSession()
	}

	return s
}

func (s Sierra) httpGet(url, accessToken string) (string, error) {
	s.log("HTTP GET", url)
	req, err := http.NewRequest("GET", url, nil)
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := ioutil.ReadAll(resp.Body)
		s.log("HTTP ERROR", string(body))
		return string(body), fmt.Errorf("Status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	return string(body), err
}

func (s Sierra) httpPost(url string, headers map[string]string) (string, error) {
	s.log("HTTP POST", url)
	req, err := http.NewRequest("POST", url, nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body), err
}

func (s *Sierra) loadSession() {
	bytes, err := ioutil.ReadFile(s.SessionFile)
	if err != nil {
		s.log("ERROR in loadSession", err.Error())
		return
	}

	var auth authResp
	err = json.Unmarshal(bytes, &auth)
	s.Authorization = auth
}

func (s Sierra) log(msg1, msg2 string) {
	if s.Verbose {
		log.Printf("%s: %s", msg1, msg2)
	}
}

func (s *Sierra) saveSession() error {
	bytes, err := json.Marshal(s.Authorization)
	if err != nil {
		return err
	}
	// http://stackoverflow.com/a/18415935/446681
	var normalAccess os.FileMode = 0644
	err = ioutil.WriteFile(s.SessionFile, bytes, normalAccess)
	return err
}
