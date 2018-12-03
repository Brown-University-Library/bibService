package sierra

import (
	"encoding/json"
	"strings"
)

type Checkouts struct {
	Total   int             `json:"total"`
	Entries []CheckoutEntry `json:"entries"`
}

type CheckoutEntry struct {
	ID          string `json:"id"`
	Patron      string `json:"patron"`
	ItemURL     string `json:"item"`
	DueDate     string `json:"dueDate"`
	NumRenewals int    `json:"numberOfRenewals"`
	OutDate     string `json:"outDate"`
}

func (e CheckoutEntry) ItemID() string {
	tokens := strings.Split(e.ItemURL, "/")
	if len(tokens) == 0 {
		return ""
	}
	return tokens[len(tokens)-1]
}

func (s *Sierra) Checkouts(patronId string) (Checkouts, error) {
	err := s.authenticate()
	if err != nil {
		return Checkouts{}, err
	}

	url := s.ApiUrl + "/patrons/" + patronId + "/checkouts"
	body, err := s.httpGet(url, s.Authorization.AccessToken)
	if err != nil {
		return Checkouts{}, err
	}

	var checkouts Checkouts
	err = json.Unmarshal([]byte(body), &checkouts)
	return checkouts, err
}
