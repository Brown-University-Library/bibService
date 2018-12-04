package sierra

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type authResp struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`  // (default 3600 seconds)
	URL         string    `json:"url"`         // non-III value
	ValidFrom   time.Time `json:"valid_from"`  // non-III value
	ValidUntil  time.Time `json:"valid_until"` // non-III value
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

	url := s.URL + "/token"
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
	auth.URL = s.URL
	auth.ValidFrom = time.Now()
	auth.ValidUntil = auth.ValidFrom.Add(duration)
	s.Authorization = auth
	if s.Persistent {
		err = s.saveSession()
	}

	return err
}
