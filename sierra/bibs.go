package sierra

import (
	"strings"
)

type Bibs struct {
	Total   int   `json:"total"`
	Entries []Bib `json:"entries"`
}

func (b Bibs) BibsIdStr() string {
	ids := []string{}
	for _, bib := range b.Entries {
		ids = append(ids, bib.Id)
	}
	return strings.Join(ids, ",")
}

func (b Bibs) BibsIdPages() [][]string {
	ids := []string{}
	for _, bib := range b.Entries {
		ids = append(ids, bib.Id)
	}
	return arrayToPages(ids, 10)
}
