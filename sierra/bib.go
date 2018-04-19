package sierra

import (
	"regexp"
	"strings"
)

type BibsResp struct {
	Total   int       `json:"total"`
	Entries []BibResp `json:"entries"`
}

type BibResp struct {
	Id              string            `json:"id"`
	UpdatedDateTime string            `json:"updatedDate,omitempty"`
	CreatedDate     string            `json:"createdDate,omitempty"`
	DeletedDate     string            `json:"deletedDate,omitempty"`
	Deleted         bool              `json:"deleted,omitempty"`
	Suppressed      bool              `json:"suppressed,omitempty"`
	Available       bool              `json:"available,omitempty"`
	Lang            map[string]string `json:"lang,omitempty"`
	Title           string            `json:"title,omitempty"`
	Author          string            `json:"author,omitempty"`
	MaterialType    map[string]string `json:"materialType,omitempty"`
	BibLevel        map[string]string `json:"bibLevel,omitempty"`
	PublishYear     int               `json:"publishYear,omitempty"`
	CatalogDate     string            `json:"catalogDate,omitempty"`
	Country         map[string]string `json:"country,omitempty"`
	NormTitle       string            `json:"normTitle,omitempty"`
	NormAuthor      string            `json:"normAuthor,omitempty"`
	VarFields       []VarFieldResp    `json:"varFields,omitempty"`
	Items           []ItemResp
	// Locations    []map[string]string `json:"locations"`
}

type VarFieldResp struct {
	FieldTag  string              `json:"fieldTag"`
	MarcTag   string              `json:"marcTag"`
	Ind1      string              `json:"ind1"`
	Ind2      string              `json:"ind2"`
	Subfields []map[string]string `json:"subfields"`
	Content   string              `json:"content"`
}

func (b BibsResp) BibsIdStr() string {
	ids := []string{}
	for _, bib := range b.Entries {
		ids = append(ids, bib.Id)
	}
	return strings.Join(ids, ",")
}

func (b BibsResp) BibsIdPages() [][]string {
	ids := []string{}
	for _, bib := range b.Entries {
		ids = append(ids, bib.Id)
	}
	return arrayToPages(ids, 10)
}

// fieldSpec is something in the form "nnna" where "nnn" is the tag of the
// field and "a" represents the subfields. For example: "100ac" means
// field "100" subfields "a" and "c". Multiple fields can be indicated
// separated by colons, for example: "100ac:210f"
func (bib BibResp) MarcValues(fieldSpec string) []string {
	values := []string{}
	for _, spec := range NewFieldSpecs(fieldSpec) {
		fields, found := bib.getFields(spec.MarcTag)
		if !found {
			continue
		}

		if len(spec.Subfields) == 0 {
			for _, field := range fields {
				if field.Content != "" {
					values = append(values, field.Content)
				}
			}
			continue
		}

		for _, field := range fields {
			subValues := field.getSubfieldsValues(spec.Subfields)
			if len(subValues) > 0 {
				values = append(values, strings.Join(subValues, " "))
			}
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
	return strings.Join(values, " ")
}

func (bib BibResp) MarcValueTrim(fieldSpec string) string {
	values := bib.MarcValues(fieldSpec)
	if len(values) == 0 {
		return ""
	}
	return trimPunct(strings.Join(values, " "))
}

func (bib BibResp) getFields(marcTag string) ([]VarFieldResp, bool) {
	fields := []VarFieldResp{}
	for _, field := range bib.VarFields {
		if field.MarcTag == marcTag {
			fields = append(fields, field)
		}
	}
	return fields, len(fields) > 0
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

func (bib BibResp) OclcNum() []string {
	// RegEx based on Traject's marc21.rb
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21_semantics.rb
	re := regexp.MustCompile("\\s*(ocm|ocn|on|\\(OCoLC\\))(\\d+)")
	values := []string{}
	for _, value := range bib.MarcValues("001:035a:035z") {
		num := strings.TrimSpace(re.ReplaceAllString(value, "$2"))
		if num != "" {
			if !in(values, num) {
				values = append(values, num)
			}
		}
	}
	return values
}

func (bib BibResp) UpdatedDate() string {
	if len(bib.UpdatedDateTime) < 10 {
		return bib.UpdatedDateTime
	}
	// Drop the time value
	return bib.UpdatedDateTime[0:10]
}

func (bib BibResp) IsOnline() bool {
	v945l := bib.MarcValue("945l")
	if strings.HasPrefix(v945l, "es") {
		return true
	}

	for _, value := range bib.MarcValues("998a") {
		if value == "es001" {
			return true
		}
	}
	return false
}

func (bib BibResp) Format() string {
	// TODO: Do we need the Traject logic for this or is this value enough?
	return bib.MaterialType["value"]
}

func (bib BibResp) LanguageName() string {
	// TODO: Do we need the Traject logic for this or is this value enough?
	return bib.Lang["name"]
}

func (bib BibResp) AuthorFacet() []string {
	// TODO: add logic for field "710" indicator 2 "9"
	// Make sure we remove "." authors
	values := bib.MarcValuesTrim("100abcd:110ab:111ab:700abcd:710ab:711ab")
	return values
}

func (bib BibResp) LocationCodes() []string {
	values := []string{}
	for _, item := range bib.Items {
		code := item.Location["code"]
		if !in(values, code) {
			values = append(values, code)
		}
	}
	return values
}

func (bib BibResp) BuildingFacets() []string {
	values := []string{}
	for _, item := range bib.Items {
		name := item.BuildingName()
		if !in(values, name) {
			values = append(values, name)
		}
	}
	return values
}

func (bib BibResp) SortableTitle() string {
	// TODO do we need the field k and indicator 2 logic?
	// as in get_sortable_title() in
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21_semantics.rb
	values := bib.MarcValuesTrim("245ab")
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (bib BibResp) CallNumbers() []string {
	values := []string{}
	callNumbers := bib.MarcValuesTrim("050ab:090ab:091ab:092ab:096ab:099ab")
	for _, number := range callNumbers {
		if !in(values, number) {
			values = append(values, number)
		}
	}
	return values
}

func (bib BibResp) Subjects() []string {
	spec := "600abcdefghjklmnopqrstuvxyz"
	spec += ":610abcdefghklmnoprstuvxyz"
	spec += ":611acdefghjklnpqstuvxyz"
	spec += ":630adefghklmnoprstvxyz"
	spec += ":648avxyz:650abcdevxyz:651aevxyz:653a"
	spec += ":654abevyz:655abvxyz:656akvxyz:657avxyz"
	spec += ":658ab:662abcdefgh:690abcdevxyz"
	return bib.MarcValuesTrim(spec)
}
