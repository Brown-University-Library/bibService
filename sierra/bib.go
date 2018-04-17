package sierra

import (
	"regexp"
	"strings"
)

type BibsResp struct {
	Total   int       `json:"total"`
	Entries []BibResp `json:"entries"`
}

func (b BibsResp) BibsIdStr() string {
	ids := []string{}
	for _, bib := range b.Entries {
		ids = append(ids, bib.Id)
	}
	return strings.Join(ids, ",")
}

type BibResp struct {
	Id           string            `json:"id"`
	UpdatedDate  string            `json:"updatedDate,omitempty"`
	CreatedDate  string            `json:"createdDate,omitempty"`
	DeletedDate  string            `json:"deletedDate,omitempty"`
	Deleted      bool              `json:"deleted,omitempty"`
	Suppressed   bool              `json:"suppressed,omitempty"`
	Available    bool              `json:"available,omitempty"`
	Lang         map[string]string `json:"lang,omitempty"`
	Title        string            `json:"title,omitempty"`
	Author       string            `json:"author,omitempty"`
	MaterialType map[string]string `json:"materialType,omitempty"`
	BibLevel     map[string]string `json:"bibLevel,omitempty"`
	PublishYear  int               `json:"publishYear,omitempty"`
	CatalogDate  string            `json:"catalogDate,omitempty"`
	Country      map[string]string `json:"country,omitempty"`
	NormTitle    string            `json:"normTitle,omitempty"`
	NormAuthor   string            `json:"normAuthor,omitempty"`
	VarFields    []VarFieldResp    `json:"varFields,omitempty"`
	Items        []ItemResp
	// Locations    []map[string]string `json:"locations"`
}

type FieldSpec struct {
	MarcTag   string
	Subfields []string
}

type VarFieldResp struct {
	FieldTag  string              `json:"fieldTag"`
	MarcTag   string              `json:"marcTag"`
	Ind1      string              `json:"ind1"`
	Ind2      string              `json:"ind2"`
	Subfields []map[string]string `json:"subfields"`
	Content   string              `json:"content"`
}

// fieldSpec is something in the form "nnna" where "nnn" is the tag of the
// field and "a" represents the subfields. For example: "100ac" means
// field "100" subfields "a" and "c". Multiple fields can be indicated
// separated by colons, for example: "100ac:210f"
func (bib BibResp) MarcValues(fieldSpec string) []string {
	values := []string{}
	for _, spec := range NewFieldSpecs(fieldSpec) {
		field, found := bib.getField(spec.MarcTag)
		if !found {
			continue
		}

		if len(spec.Subfields) == 0 {
			if field.Content != "" {
				values = append(values, field.Content)
			}
			continue
		}

		newValues := field.getSubfieldsValues(spec.Subfields)
		for _, value := range newValues {
			values = append(values, value)
		}
	}
	return values
}

func (bib BibResp) MarcValuesTrim(fieldSpec string) []string {
	values := []string{}
	for _, value := range bib.MarcValues(fieldSpec) {
		values = append(values, trimPunct(value))
	}
	return values
}

func (bib BibResp) MarcValue(fieldSpec string) string {
	values := bib.MarcValues(fieldSpec)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (bib BibResp) MarcValueTrim(fieldSpec string) string {
	return trimPunct(bib.MarcValue(fieldSpec))
}

func (bib BibResp) getField(marcTag string) (VarFieldResp, bool) {
	for _, field := range bib.VarFields {
		if field.MarcTag == marcTag {
			return field, true
		}
	}
	return VarFieldResp{}, false
}

func (field VarFieldResp) getSubfieldsValues(subfields []string) []string {
	values := []string{}
	for _, subfield := range subfields {
		for _, fieldSub := range field.Subfields {
			if fieldSub["tag"] == subfield {
				values = append(values, fieldSub["content"])
			}
		}
	}
	return values
}

func NewFieldSpecs(spec string) []FieldSpec {
	fieldSpecs := []FieldSpec{}
	for _, token := range strings.Split(spec, ":") {
		length := len(token)
		if length < 3 {
			// not a valid spec
			continue
		}

		fieldSpec := FieldSpec{
			MarcTag:   token[0:3],
			Subfields: []string{},
		}

		if length > 3 {
			// process the subfields in the spec
			for _, c := range token[3:length] {
				fieldSpec.Subfields = append(fieldSpec.Subfields, string(c))
			}
		}
		fieldSpecs = append(fieldSpecs, fieldSpec)
	}
	return fieldSpecs
}

func trimPunct(str string) string {
	if str == "" {
		return str
	}

	// RegEx stolen from Traject's marc21.rb
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21.rb
	//
	// # trailing: comma, slash, semicolon, colon (possibly preceded and followed by whitespace)
	// str = str.sub(/ *[ ,\/;:] *\Z/, '')
	re1 := regexp.MustCompile(" *[ ,\\/;:] *$")
	cleanStr := re1.ReplaceAllString(str, "")

	// # trailing period if it is preceded by at least three letters (possibly preceded and followed by whitespace)
	// str = str.sub(/( *\w\w\w)\. *\Z/, '\1')
	re2 := regexp.MustCompile("( *\\w\\w\\w)\\. *$")
	cleanStr = re2.ReplaceAllString(cleanStr, "$1")

	// # single square bracket characters if they are the start
	// # and/or end chars and there are no internal square brackets.
	// str = str.sub(/\A\[?([^\[\]]+)\]?\Z/, '\1')
	re3 := regexp.MustCompile("^\\[?([^\\[\\]]+)\\]?$")
	cleanStr = re3.ReplaceAllString(cleanStr, "$1")

	return cleanStr
}
