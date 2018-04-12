package sierra

// Sierra API documentation:
//	https://sandbox.iii.com/iii/sierra-api/swagger/index.html
// 	https://techdocs.iii.com/sierraapi/Content/zReference/objects/bibObjectDescription.htm
// 	https://techdocs.iii.com/sierraapi/Content/zAppendix/bibObjectExample.htm

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type authResp struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`  // (default 3600 seconds)
	Url         string    `json:"url"`         // non-III value
	ValidFrom   time.Time `json:"valid_from"`  // non-III value
	ValidUntil  time.Time `json:"valid_until"` // non-III value
}

type marcFileResp struct {
	File        string `json:"file"`
	InputCount  int    `json:"inputRecords"`
	OutputCount int    `json:"outputRecords"`
	ErrorCount  int    `json:"errors"`
}

type Sierra struct {
	ApiUrl        string
	Persistent    bool
	KeySecret     string
	KeySecret64   string
	Authorization authResp
	Verbose       bool
	SessionFile   string
}

func NewSierra(apiUrl, keySecret, sessionFile string) Sierra {
	s := Sierra{
		ApiUrl:      apiUrl,
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

func (s *Sierra) Search(value string) (string, error) {
	err := s.authenticate()
	if err != nil {
		return "", err
	}

	url := s.ApiUrl + "/bibs"
	url += "?deleted=false&suppressed=false&fields=title,author,publishYear,updatedDate"
	body, err := s.httpGet(url, s.Authorization.AccessToken)
	return body, err
}

func (s *Sierra) Get(params map[string]string) (BibsResp, error) {
	body, err := s.GetRaw(params)
	if err != nil {
		return BibsResp{}, err
	}

	var bibs BibsResp
	err = json.Unmarshal([]byte(body), &bibs)
	return bibs, err
}

func (s *Sierra) GetRaw(params map[string]string) (string, error) {
	err := s.authenticate()
	if err != nil {
		return "", err
	}

	// fixedFields,
	fields := "fields=default,available,orders,normTitle,normAuthor,locations,varFields"
	url := s.ApiUrl + "/bibs?"
	for key, value := range params {
		url += key + "=" + value + "&"
	}
	url += fields
	return s.httpGet(url, s.Authorization.AccessToken)
}

func (s *Sierra) BibsUpdatedSince(date string) (BibsResp, error) {
	err := s.authenticate()
	if err != nil {
		return BibsResp{}, err
	}

	url := s.ApiUrl + "/bibs?updatedDate=" + date
	body, err := s.httpGet(url, s.Authorization.AccessToken)
	if err != nil {
		return BibsResp{}, err
	}

	var bibs BibsResp
	err = json.Unmarshal([]byte(body), &bibs)
	// Should this return an array of IDs instead?
	return bibs, err
}

func (s *Sierra) Items(bibRange string) (ItemsResp, error) {
	err := s.authenticate()
	if err != nil {
		return ItemsResp{}, err
	}

	body, err := s.ItemsRaw(bibRange)
	if err != nil {
		return ItemsResp{}, err
	}

	var items ItemsResp
	err = json.Unmarshal([]byte(body), &items)
	return items, err
}

func (s *Sierra) ItemsRaw(bibRange string) (string, error) {
	err := s.authenticate()
	if err != nil {
		return "", err
	}

	url := s.ApiUrl + "/items?bibIds=" + bibRange
	return s.httpGet(url, s.Authorization.AccessToken)
}

// idRange can be a single ID or a comma delimited list of IDs.
// Becareful because it seems that Sierra's backend chokes when
// the list is to long (e.g. it fails with 50 IDs)
func (s *Sierra) Marc(idRange string) (string, error) {
	err := s.authenticate()
	if err != nil {
		return "", err
	}

	mapping := "" // TODO
	url := s.ApiUrl + "/bibs/marc?id=" + idRange
	if mapping != "" {
		url += "&mapping=" + mapping
	}

	body, err := s.httpGet(url, s.Authorization.AccessToken)
	if err != nil {
		return "", err
	}

	var marcFile marcFileResp
	err = json.Unmarshal([]byte(body), &marcFile)
	if err != nil {
		return "", err
	}

	data, err := s.httpGet(marcFile.File, s.Authorization.AccessToken)
	return data, err
}

func (s *Sierra) Deleted(dateRange string) (string, error) {
	err := s.authenticate()
	if err != nil {
		return "", err
	}

	url := s.ApiUrl
	if dateRange == "" {
		url += "/bibs?deleted=true"
	} else {
		// TODO: validate dateRange is in the form a,b
		url += fmt.Sprintf("/bibs?deletedDate=[%s]", dateRange)
	}
	body, err := s.httpGet(url, s.Authorization.AccessToken)
	return body, err
}

func (s *Sierra) loadSession() {
	bytes, err := ioutil.ReadFile(s.SessionFile)
	if err != nil {
		return
	}

	var auth authResp
	err = json.Unmarshal(bytes, &auth)
	s.Authorization = auth
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

func (s Sierra) isAuthenticated() bool {
	if s.Authorization.AccessToken == "" {
		return false
	}
	validSession := time.Now().Before(s.Authorization.ValidUntil)
	return validSession
}

func (s *Sierra) authenticate() error {
	if s.isAuthenticated() {
		return nil
	}

	url := s.ApiUrl + "/token"
	headers := map[string]string{
		"Authorization": "Basic " + s.KeySecret64,
		"Content-Type":  "text/plain",
	}
	body, err := s.httpPost(url, headers)
	if err != nil {
		return err
	}

	var auth authResp
	err = json.Unmarshal([]byte(body), &auth)
	if err != nil {
		return err
	}

	if auth.AccessToken == "" {
		errorMsg := fmt.Sprintf("No authentication token was returned %s", body)
		return errors.New(errorMsg)
	}

	duration := time.Duration(auth.ExpiresIn) * time.Second
	auth.Url = s.ApiUrl
	auth.ValidFrom = time.Now()
	auth.ValidUntil = auth.ValidFrom.Add(duration)
	s.Authorization = auth
	if s.Persistent {
		err = s.saveSession()
	}

	return err
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
	body, err := ioutil.ReadAll(resp.Body)
	// s.log("body", string(body))
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

func (s Sierra) log(msg1, msg2 string) {
	if s.Verbose {
		log.Printf("%s: %s", msg1, msg2)
	}
}
