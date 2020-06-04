package sierra

import "encoding/json"

// Items represents a collection of Sierra items.
type Items struct {
	Total   int    `json:"total"`
	Entries []Item `json:"entries"`
}

// ForBib returns the items that belong to the specified BIB. Takes into
// acount that items in the collection could be for more than one
// BIB and also that a single Item could belong to more than one BIB.
func (items Items) ForBib(bib string) []Item {
	bibItems := []Item{}
	for _, item := range items.Entries {
		if item.IsForBib(bib) {
			bibItems = append(bibItems, item)
		}
	}
	return bibItems
}

// Items fetches item information for a comma delimited list of Bib IDs.
func (s *Sierra) Items(bibsList string) (Items, error) {
	err := s.authenticate()
	if err != nil {
		return Items{}, err
	}

	body, err := s.ItemsRaw(bibsList)
	if err != nil {
		return Items{}, err
	}

	var items Items
	err = json.Unmarshal([]byte(body), &items)
	return items, err
}

// ItemsRaw returns the raw item information (a string) for a comma delimited list of Bib IDs.
func (s *Sierra) ItemsRaw(bibsList string) (string, error) {
	err := s.authenticate()
	if err != nil {
		return "", err
	}

	url := s.URL + "/items?bibIds=" + bibsList
	url += "&fields=default,varFields,fixedFields"
	return s.httpGet(url, s.Authorization.AccessToken)
}
