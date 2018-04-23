package sierra

type Items struct {
	Total   int    `json:"total"`
	Entries []Item `json:"entries"`
}

func (items Items) ForBib(bib string) []Item {
	bibItems := []Item{}
	for _, item := range items.Entries {
		if item.IsForBib(bib) {
			bibItems = append(bibItems, item)
		}
	}
	return bibItems
}
