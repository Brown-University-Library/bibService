package sierra

import (
	"strings"
)

type ItemsResp struct {
	Total   int        `json:"total"`
	Entries []ItemResp `json:"entries"`
}

type ItemResp struct {
	Id          string            `json:"id"`
	UpdatedDate string            `json:"updatedDate"`
	CreatedDate string            `json:"createdDate"`
	Deleted     bool              `json:"deleted"`
	BibIds      []string          `json:"bibIds"`
	Location    map[string]string `json:"location"`
	Status      map[string]string `json:"status"`
	Barcode     string            `json:"barcode"`
}

func (i ItemResp) IsForBib(bib string) bool {
	for _, b := range i.BibIds {
		if b == bib {
			return true
		}
	}
	return false
}

func (i ItemResp) BarcodeClean() string {
	return strings.Replace(i.Barcode, " ", "", -1)
}

func (i ItemResp) LocationName() string {
	return i.Location["name"]
}

func (i ItemResp) StatusDisplay() string {
	dueDate := i.Status["duedate"]
	if dueDate != "" {
		// TODO: Account for time zones, format as MM/DD/YYYY
		return "DUE " + dueDate
	}
	return i.Status["display"]
}
