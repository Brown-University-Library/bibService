package sierra

import (
	"strings"
)

type Field struct {
	FieldTag  string              `json:"fieldTag"`
	MarcTag   string              `json:"marcTag"`
	Ind1      string              `json:"ind1"`
	Ind2      string              `json:"ind2"`
	Subfields []map[string]string `json:"subfields"`
	Content   string              `json:"content"`
}

func (f Field) Tags() []string {
	tags := []string{}
	for _, sub := range f.Subfields {
		safeAppend(&tags, sub["tag"])
	}
	return tags
}

func (f Field) HasVernacular() (bool, string) {
	for _, sub := range f.Subfields {
		if sub["tag"] == "6" {
			return true, sub["content"]
		}
	}
	return false, ""
}

func (f Field) IsVernacularForTag6(target string) bool {
	for _, sub := range f.Subfields {
		if sub["tag"] == "6" {
			if strings.HasPrefix(sub["content"], target) {
				return true
			}
		}
	}
	return false
}

func (f Field) ValuesForForTag6(subfields []string) []string {
	values := []string{}
	for _, sub := range f.Subfields {
		if in(subfields, sub["tag"]) {
			safeAppend(&values, sub["content"])
		}
	}
	return values
}

func (f Field) getSubfieldsValues(subfields []string) []string {
	values := []string{}
	// We walk through the subfields in the Field because it is important
	// to preserve the order of the values returned according to the order
	// in which they are listed on the data, not on the spec.
	for _, fieldSub := range f.Subfields {
		for _, specSub := range subfields {
			if fieldSub["tag"] == specSub {
				safeAppend(&values, fieldSub["content"])
			}
		}
	}
	return values
}
