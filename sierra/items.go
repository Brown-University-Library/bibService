package sierra

type Items struct {
	Total   int    `json:"total"`
	Entries []Item `json:"entries"`
}

// Returns the items that belong to the specified BIB. Takes into
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
