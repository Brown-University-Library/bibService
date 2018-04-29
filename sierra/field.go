package sierra

import (
	"strings"
)

type MarcField struct {
	FieldTag  string              `json:"fieldTag"`
	MarcTag   string              `json:"marcTag"`
	Ind1      string              `json:"ind1"`
	Ind2      string              `json:"ind2"`
	Subfields []map[string]string `json:"subfields"`
	Content   string              `json:"content"`
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
	if join {
		return f.valuesJoin(tagsWanted)
	}
	return f.valuesNoJoin(tagsWanted)
}

func (f MarcField) valuesNoJoin(subfields []string) []string {
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

// Gets the values in a Field and outputs the tags requested.
// The logic to group the output is a bit complex because it combines
// the values for different tags into a single value. For example,
// if we want tags "abc" from a field with the following information:
//
//    tag   content
//    ---   -------
//    a      A1
//    b      B1
//    a      A2
//    a      A3
//    c      C3
//
// it will output:
//
//      "A1 B1"        // combined two tags
//      "A2"           // single tag
//      "A3 C3"        // combined two tags
//
func (f MarcField) valuesJoin(tagsWanted []string) []string {
	output := []string{}
	processedTags := []string{}
	batchValues := []string{}
	for _, subfield := range f.Subfields {
		tag := subfield["tag"]
		content := subfield["content"]
		tagAlreadyProcessed := in(processedTags, tag)
		if tagAlreadyProcessed {
			// output whatever we've gathered so far...
			if len(batchValues) > 0 {
				output = append(output, strings.Join(batchValues, " "))
			}

			// start a new batch...
			processedTags = []string{}
			batchValues = []string{}
		}

		if in(tagsWanted, tag) && content != "" {
			// add value to the batch
			batchValues = append(batchValues, content)
		}
		processedTags = append(processedTags, tag)
	}

	if len(batchValues) > 0 {
		// output the last batch
		output = append(output, strings.Join(batchValues, " "))
	}
	return output
}
