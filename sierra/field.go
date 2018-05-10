package sierra

import (
	"strings"
)

type MarcValues []MarcValue

type MarcValue struct {
	MarcTag  string // e.g. 700
	Subfield string // e.g. a
	Value    string
}

type MarcField struct {
	FieldTag  string              `json:"fieldTag"`
	MarcTag   string              `json:"marcTag"`
	Ind1      string              `json:"ind1"`
	Ind2      string              `json:"ind2"`
	Subfields []map[string]string `json:"subfields"`
	Content   string              `json:"content"`
}

func (f MarcField) Add(subField string, value string) {
	data := map[string]string{subField: value}
	f.Subfields = append(f.Subfields, data)
}

func (f MarcField) Set(subFields []map[string]string) {
	f.Subfields = subFields
}

func (f MarcField) Tags() []string {
	tags := []string{}
	for _, sub := range f.Subfields {
		safeAppend(&tags, sub["tag"])
	}
	return tags
}

// Returns true if the field indicates that there are vernacular
// values associated with it. It also returns the MARC field
// where the vernacular values are.
func (f MarcField) HasVernacular() (bool, string) {
	for _, sub := range f.Subfields {
		if sub["tag"] == "6" {
			return true, sub["content"]
		}
	}
	return false, ""
}

// Returns true if the field contains vernacular values for another
// (target) field.
//
// A field is considered to have vernacular values for another field if
// it has a subfield with tag "6" where the content is for the target.
// The Target tyically comes in the form "NNN-nn" where "NNN" is the MARC
// field and "nn" is a sequence value (e.g. "700-02")
func (f MarcField) IsVernacularFor(target string) bool {
	for _, sub := range f.Subfields {
		if sub["tag"] == "6" {
			if strings.HasPrefix(sub["content"], target) {
				return true
			}
		}
	}
	return false
}

func (f MarcField) Values(tagsWanted []string, join bool) []string {
	values := f.valuesNoJoin(tagsWanted)
	if len(tagsWanted) > 1 && join {
		return []string{strings.Join(values, " ")}
	}
	return values
}

func (f MarcField) ValuesNew(subsWanted []string) []MarcValue {
	values := []MarcValue{}
	// We walk through the subfields in the Field because it is important
	// to preserve the order of the values returned according to the order
	// in which they are listed on the data, not on the spec.
	for _, fieldSub := range f.Subfields {
		for _, sub := range subsWanted {
			if fieldSub["tag"] == sub {
				content := fieldSub["content"]
				if content != "" {
					value := MarcValue{MarcTag: f.MarcTag, Subfield: sub, Value: content}
					values = append(values, value)
				}
			}
		}
	}
	return values
}

func (f MarcField) valuesNoJoin(subfields []string) []string {
	values := []string{}
	// We walk through the subfields in the Field because it is important
	// to preserve the order of the values returned according to the order
	// in which they are listed on the data, not on the spec.
	for _, fieldSub := range f.Subfields {
		for _, specSub := range subfields {
			if fieldSub["tag"] == specSub {
				content := fieldSub["content"]
				if content != "" {
					values = append(values, content)
				}
			}
		}
	}
	return values
}
