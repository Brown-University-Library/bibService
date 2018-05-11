package sierra

import (
	"bibService/marc"
	"strings"
)

type Item struct {
	Id          string            `json:"id"`
	UpdatedDate string            `json:"updatedDate"`
	CreatedDate string            `json:"createdDate"`
	Deleted     bool              `json:"deleted"`
	BibIds      []string          `json:"bibIds"`
	Location    map[string]string `json:"location"`
	Status      map[string]string `json:"status"`
	Barcode     string            `json:"barcode"`
	Fields      []marc.MarcField  `json:"varFields"`
}

func (i Item) IsForBib(bib string) bool {
	for _, b := range i.BibIds {
		if b == bib {
			return true
		}
	}
	return false
}

func (i Item) BarcodeClean() string {
	return strings.Replace(i.Barcode, " ", "", -1)
}

func (i Item) LocationName() string {
	return i.Location["name"]
}

func (i Item) BuildingName() string {
	return buildingName(i.Location["code"])
}

func (i Item) StatusDisplay() string {
	dueDate := i.Status["duedate"]
	if dueDate != "" {
		// TODO: Account for time zones, format as MM/DD/YYYY
		return "DUE " + dueDate
	}
	return i.Status["display"]
}

func (i Item) BookplateCodes() []string {
	values := []string{}
	for _, field := range i.Fields {
		if field.FieldTag == "f" {
			safeAppend(&values, field.Content)
		}
	}
	return values
}
