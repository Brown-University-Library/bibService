package sierra

import (
	"encoding/json"
	"strings"
)

// Checkouts represents the result from the Sierra API /v5/patrons/{id}/checkouts endpoint
type Checkouts struct {
	Total   int             `json:"total"`
	Entries []CheckoutEntry `json:"entries"`
}

// CheckoutEntry represents an individual entry in the Sierra API /v5/patrons/{id}/checkouts endpoint
type CheckoutEntry struct {
	ID          string `json:"id"`
	Patron      string `json:"patron"`
	ItemURL     string `json:"item"`
	DueDate     string `json:"dueDate"`
	NumRenewals int    `json:"numberOfRenewals"`
	OutDate     string `json:"outDate"`
}

// ItemID returns the item ID for the given checkout entry.
func (e CheckoutEntry) ItemID() string {
	tokens := strings.Split(e.ItemURL, "/")
	if len(tokens) == 0 {
		return ""
	}
	return tokens[len(tokens)-1]
}

// IsBorrowDirect returns true if the item is a Borrow Direct item.
// Borrow Direct Items have an ItemURL in the form "http://.../nnnnnn@xxxx" where xxxx seems
// to be value "ncip" but I suspect it could be other values too.
func (e CheckoutEntry) IsBorrowDirect() bool {
	return strings.Contains(e.ItemID(), "@")
}

// Checkouts returns the checkout information for the given patron ID.
func (s *Sierra) Checkouts(patronID string) (Checkouts, error) {
	err := s.authenticate()
	if err != nil {
		return Checkouts{}, err
	}

	url := s.URL + "/patrons/" + patronID + "/checkouts"
	body, err := s.httpGet(url, s.Authorization.AccessToken)
	if err != nil {
		return Checkouts{}, err
	}

	var checkouts Checkouts
	err = json.Unmarshal([]byte(body), &checkouts)
	return checkouts, err
}
