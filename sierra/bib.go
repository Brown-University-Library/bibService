package sierra

import (
	"regexp"
	"strings"
	"time"
)

// Bib represents a bibliographic record.
// Notice that Bib records in Sierra don't include "item" data but this
// struct can accomodate them.
type Bib struct {
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
	VarFields       []Field           `json:"varFields,omitempty"`
	Items           []Item            // does not come on the Sierra response
}

func (b Bib) Bib() string {
	return "b" + b.Id
}

func (bib Bib) VernacularValues(specsStr string) []string {
	values := []string{}
	for _, spec := range NewFieldSpecs(specsStr) {
		for _, field := range bib.VarFields {
			vern := bib.VernacularValuesFor(field, spec)
			arrayAppend(&values, vern)
		}
	}
	return values
}

func (bib Bib) VernacularValuesFor(field Field, spec FieldSpec) []string {
	values := []string{}

	// True if the field has subfield with tag 6
	// target would be "880-04"
	vern, target := field.HasVernacular()
	if !vern {
		return values
	}

	tokens := strings.Split(target, "-")
	marcTag := tokens[0]                    // 880
	tag6 := field.MarcTag + "-" + tokens[1] // 700-04

	// Process the fields indicated in target (e.g. 880s)
	for _, vernField := range bib.getFields(marcTag) {
		// if this is the one that corresponds with our target
		// e.g. 700-04
		if vernField.IsVernacularForTag6(tag6) {
			vernValues := vernField.ValuesForForTag6(spec.Subfields)
			safeAppend(&values, strings.Join(vernValues, " "))
		}
	}
	return values
}

// func (bib Bib) VernacularValues(specsStr string) []string {
// 	values := []string{}
// 	f880s := bib.getFields("880")
// 	for _, spec := range NewFieldSpecs(specsStr) {
// 		vern := bib.vernacularValues(f880s, spec)
// 		safeAppend(&values, strings.Join(vern, " "))
// 		// arrayAppend(&values, vern)
// 	}
// 	return values
// }
//
// func (bib Bib) VernacularValue(fieldSpec string) string {
// 	values := bib.VernacularValues(fieldSpec)
// 	if len(values) == 0 {
// 		return ""
// 	}
// 	return strings.Join(values, " ")
// }
//
// func (bib Bib) VernacularValueTrim(fieldSpec string) string {
// 	values := bib.VernacularValues(fieldSpec)
// 	if len(values) == 0 {
// 		return ""
// 	}
// 	return trimPunct(strings.Join(values, " "))
// }
//
// func (bib Bib) VernacularValuesTrim(specsStr string) []string {
// 	values := []string{}
// 	for _, value := range bib.VernacularValues(specsStr) {
// 		trimValue := trimPunct(value)
// 		safeAppend(&values, trimValue)
// 	}
// 	return values
// }
//
// func (bib Bib) vernacularValues(f880s []Field, spec FieldSpec) []string {
// 	values := []string{}
// 	for _, f880 := range f880s {
// 		vern := f880.VernacularValues(spec)
// 		arrayAppend(&values, vern)
// 		// if len(spec.Subfields) == 0 {
// 		// 	vern := f880.VernacularValue(spec)
// 		// 	safeAppend(&values, vern)
// 		// } else {
// 		// 	for _, subField := range spec.Subfields {
// 		// 		subSpec, _ := NewFieldSpec(spec.MarcTag + subField)
// 		// 		vern := f880.VernacularValue(subSpec)
// 		// 		safeAppend(&values, vern)
// 		// 	}
// 		// }
// 	}
// 	return values
// }

// fieldSpec is something in the form "nnna" where "nnn" is the tag of the
// field and "a" represents the subfields. For example: "100ac" means
// field "100" subfields "a" and "c". Multiple fields can be indicated
// separated by colons, for example: "100ac:210f"
func (bib Bib) MarcValues(fieldSpec string) []string {
	values := []string{}
	// f880s := bib.getFields("880")

	for _, spec := range NewFieldSpecs(fieldSpec) {
		fields := bib.getFields(spec.MarcTag)
		if len(fields) == 0 {
			// vernacular := bib.vernacularValues(f880s, spec)
			// arrayAppend(&values, vernacular)
			continue
		}

		if len(spec.Subfields) == 0 {
			// Get the value directly
			for _, field := range fields {
				safeAppend(&values, field.Content)
			}
			continue
		}

		// Process the subfields
		for _, field := range fields {
			subValues := field.getSubfieldsValues(spec.Subfields)
			if len(spec.Subfields) == 1 {
				// single subfields specified (060a)
				// append each individual value
				for _, subValue := range subValues {
					safeAppend(&values, subValue)
				}
			} else {
				// multi-subfields specified (e.g. 060abc)
				// concatenate the values and then append them
				strVal := strings.Join(subValues, " ")
				safeAppend(&values, strVal)
			}
		}

		for _, field := range fields {
			vernacular := bib.VernacularValuesFor(field, spec)
			arrayAppend(&values, vernacular)
		}

		// vernacular := bib.vernacularValues(f880s, spec)
		// arrayAppend(&values, vernacular)
	}
	return values
}

// func (bib Bib) MarcValuesOld(fieldSpec string) []string {
// 	values := []string{}
// 	f880s := bib.getFields("880")
//
// 	for _, spec := range NewFieldSpecs(fieldSpec) {
// 		fields := bib.getFields(spec.MarcTag)
// 		if len(fields) == 0 {
// 			vernacular := bib.vernacularValues(f880s, spec)
// 			arrayAppend(&values, vernacular)
// 			continue
// 		}
//
// 		if len(spec.Subfields) == 0 {
// 			// Get the value directly
// 			for _, field := range fields {
// 				safeAppend(&values, field.Content)
// 			}
// 			continue
// 		}
//
// 		// Process the subfields
// 		for _, field := range fields {
// 			subValues := field.getSubfieldsValues(spec.Subfields)
// 			if len(spec.Subfields) == 1 {
// 				// single subfields specified (060a)
// 				// append each individual value
// 				for _, subValue := range subValues {
// 					safeAppend(&values, subValue)
// 				}
// 			} else {
// 				// multi-subfields specified (e.g. 060abc)
// 				// concatenate the values and then append them
// 				strVal := strings.Join(subValues, " ")
// 				safeAppend(&values, strVal)
// 			}
// 		}
//
// 		vernacular := bib.vernacularValues(f880s, spec)
// 		arrayAppend(&values, vernacular)
// 	}
// 	return values
// }

func (bib Bib) MarcValuesTrim(fieldSpec string) []string {
	values := []string{}
	for _, value := range bib.MarcValues(fieldSpec) {
		trimValue := trimPunct(value)
		safeAppend(&values, trimValue)
	}
	return values
}

func (bib Bib) MarcValue(fieldSpec string) string {
	values := bib.MarcValues(fieldSpec)
	if len(values) == 0 {
		return ""
	}
	return strings.Join(values, " ")
}

func (bib Bib) MarcValueTrim(fieldSpec string) string {
	values := bib.MarcValues(fieldSpec)
	if len(values) == 0 {
		return ""
	}
	return trimPunct(strings.Join(values, " "))
}

func (bib Bib) getFields(marcTag string) []Field {
	fields := []Field{}
	for _, field := range bib.VarFields {
		if field.MarcTag == marcTag {
			fields = append(fields, field)
		}
	}
	return fields
}

func (bib Bib) UniformTitles(newVersion bool) []UniformTitles {
	var spec string
	if newVersion {
		spec = "240adfgklmnoprs"
	} else {
		spec = "130adfgklmnoprst"
	}
	return NewUniformTitles(bib, spec)
}

func (bib Bib) UniformTitlesDisplay(newVersion bool) string {
	var spec string
	if newVersion {
		spec = "240adfgklmnoprs"
	} else {
		spec = "130adfgklmnoprst"
	}
	titles, _ := NewUniformTitlesString(bib, spec)
	return titles
}

func (bib Bib) Isbn() []string {
	return bib.MarcValues("020a:020z")
}

func (bib Bib) TitleDisplay() string {
	titles := bib.MarcValuesTrim("245apbfgkn")
	if len(titles) > 0 {
		return titles[0]
	}
	return ""
}

func (bib Bib) TitleSeries() []string {
	specsStr := "400flnptv:410flnptv:411fklnptv:440ap:490a:800abcdflnpqt:"
	specsStr += "810tflnp:811tfklpsv:830adfklmnoprstv"
	return bib.MarcValuesTrim(specsStr)
}

func (bib Bib) TitleVernacularDisplay() string {
	titles := bib.VernacularValues("245apbfgkn")
	if len(titles) > 0 {
		return trimPunct(strings.Join(titles, " "))
	}
	return ""
}

func (bib Bib) PublishedVernacularDisplay() string {
	values := bib.VernacularValues("260a")
	if len(values) > 0 {
		return strings.Join(values, " ")
	}
	return ""
}

func (bib Bib) IsDissertaion() bool {
	subs := []string{"a", "c"}
	for _, field := range bib.getFields("502") {
		for _, value := range field.getSubfieldsValues(subs) {
			if strings.Contains(strings.ToLower(value), "brown univ") {
				return true
			}
		}
	}
	return false
}

func (bib Bib) Issn() []string {
	return bib.MarcValues("022a:022l:022y:773x:774x:776x")
}

func (bib Bib) PublicationYear() (int, bool) {
	rangeStart := 500
	rangeEnd := time.Now().Year()
	tolerance := 15

	f008 := bib.MarcValue("008")
	year, ok := pubYear008(f008, tolerance)
	if !ok {
		year, ok = bib.pubYear260()
	}

	if ok && year >= rangeStart && year <= rangeEnd {
		return year, true
	}
	return 0, false
}

func (bib Bib) pubYear260() (int, bool) {
	f260c := bib.MarcValue("260c")
	re := regexp.MustCompile("(\\d{4})")
	year := re.FindString(f260c)
	return toIntTry(year)
}

func (bib Bib) OclcNum() []string {
	// RegEx based on Traject's marc21.rb
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21_semantics.rb
	re := regexp.MustCompile("\\s*(ocm|ocn|on|\\(OCoLC\\))(\\d+)")
	values := []string{}
	for _, value := range bib.MarcValues("001:035a:035z") {
		if strings.HasPrefix(value, "ssj") {
			// TODO: Ask Jeanette about these values
			// eg. b4643178, b4643180
		} else {
			num := strings.TrimSpace(re.ReplaceAllString(value, "$2"))
			safeAppend(&values, num)
		}
	}
	return values
}

func (bib Bib) UpdatedDate() string {
	if len(bib.UpdatedDateTime) < 10 {
		return bib.UpdatedDateTime
	}
	// Drop the time value
	return bib.UpdatedDateTime[0:10]
}

func (bib Bib) IsOnline() bool {
	for _, item := range bib.Items {
		if strings.HasPrefix(item.Location["code"], "es") {
			return true
		}
	}

	for _, value := range bib.MarcValues("338a") {
		if value == "online resource" {
			return true
		}
	}

	// It seems that field 998 does not come in the API and
	// therfore this code does nothing for now.
	for _, value := range bib.MarcValues("998a") {
		if value == "es001" {
			return true
		}
	}
	return false
}

func (bib Bib) Format() string {
	// TODO: Do we need the Traject logic for this or is this value enough?
	return formatName(bib.MaterialType["value"])
}

func (bib Bib) Languages() []string {
	values := []string{}
	f008 := bib.MarcValue("008")
	f008_lang := ""
	if len(f008) > 38 {
		f008_lang = languageName(f008[35:38])
		safeAppend(&values, f008_lang)
	}

	for _, value := range bib.MarcValues("041a:041d:041e:041j") {
		language := languageName(value)
		safeAppend(&values, language)
	}
	return values
}

func (bib Bib) RegionFacet() []string {
	// Stolen from Traject's marc_geo_facet
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21_semantics.rb
	values := []string{}
	for _, value := range bib.MarcValues("043a") {
		code := trimPunct(value)
		code = strings.TrimRight(code, "-")
		name := regionName(code)
		safeAppend(&values, name)
	}

	aFieldSpec := "651a:691a"
	for _, value := range bib.MarcValues(aFieldSpec) {
		trimVal := trimPunct(value)
		safeAppend(&values, trimVal)
	}

	for _, zvalue := range bib.RegionFacetZFields() {
		safeAppend(&values, zvalue)
	}
	return values
}

func (bib Bib) RegionFacetZFields() []string {
	values := []string{}

	zFieldSpecs := []string{
		"600z", "610z", "611z", "630z", "648z", "650z",
		"654z", "655z", "656z", "690z", "651z", "691z",
	}

	// Notice that we don't use bib.MarcValues() here because
	// bib.MarcValues() returns the data without a relationship
	// to the field where each value was found. In this case
	// we care about what values are found on each instance of
	// the field so that we can concatenate "region (parent region)"
	// values if they are found in a specific field.
	for _, zFieldSpec := range zFieldSpecs {
		spec, _ := NewFieldSpec(zFieldSpec)
		for _, field := range bib.getFields(spec.MarcTag) {
			subValues := field.getSubfieldsValues(spec.Subfields)
			if len(subValues) == 2 {
				// Asumme the first one is the parent region of the second one
				// e.g. v0 := "USA", v1 := "Rhode Island (USA)"
				parentRegion := trimPunct(subValues[0])
				region := trimPunct(subValues[1]) + " (" + parentRegion + ")"
				safeAppend(&values, parentRegion)
				safeAppend(&values, region)
			} else {
				arrayAppend(&values, subValues)
			}
		}

	}
	return values
}

func (bib Bib) AuthorFacet() []string {
	specStr := "100abcd:110ab:111ab:700abcd:711ab"

	f710 := bib.getFields("710")
	if len(f710) > 0 {
		// If there is more than one 710 field this will only check the first one.
		// TODO: handle multi 710 fields
		if f710[0].Ind2 != "9" {
			specStr += ":710ab"
		}
	}

	values := bib.MarcValuesTrim(specStr)
	vernValues := bib.VernacularValues(specStr)
	arrayAppend(&values, vernValues)
	return values
}

func (bib Bib) AuthorDisplay() string {
	authors := bib.MarcValues("100abcdq:110abcd:111abcd")
	if len(authors) > 0 {
		return trimPunct(authors[0])
	}
	return ""
}

func (bib Bib) AbstractDisplay() string {
	values := bib.MarcValues("520a")
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func (bib Bib) AuthorVernacularDisplay() string {
	values := bib.VernacularValues("100abcdq:110abcd:111abcd")
	return strings.Join(values, " ")
	// return bib.VernacularValueTrim("100abcdq:110abcd:111abcd")
}

func (bib Bib) LocationCodes() []string {
	values := []string{}
	for _, item := range bib.Items {
		safeAppend(&values, item.Location["code"])
	}
	return values
}

func (bib Bib) BuildingFacets() []string {
	values := []string{}
	for _, item := range bib.Items {
		name := item.BuildingName()
		safeAppend(&values, name)
	}
	return values
}

func (bib Bib) SortableTitle() string {
	// Logic stolen from
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21_semantics.rb
	// TODO do we need the field k logic here?
	titles := bib.MarcValues("245ab")
	if len(titles) == 0 {
		return ""
	}

	sortTitle := titles[0]
	fields := bib.getFields("245")
	if len(fields) > 0 {
		ind2 := toInt(fields[0].Ind2)
		if ind2 > 0 && len(sortTitle) > ind2 {
			// drop the prefix as notes in the second indicator
			sortTitle = sortTitle[ind2:len(sortTitle)]
		}
	}
	return trimPunct(sortTitle)
}

func (bib Bib) CallNumbers() []string {
	return bib.MarcValuesTrim("050ab:090ab:091ab:092ab:096ab:099ab")
}

func (bib Bib) Subjects() []string {
	spec := "600a:600abcdefghjklmnopqrstuvxyz:"
	spec += "610a:610abcdefghklmnoprstuvxyz:"
	spec += "611a:611acdefghjklnpqstuvxyz:"
	spec += "630a:630adefghklmnoprstvxyz:"
	spec += "648a:648avxyz:"
	spec += "650a:650abcdezxvy:"
	spec += "651a:651aexzvy:"
	spec += "653a:654abevyz:"
	spec += "654a:655abvxyz:"
	spec += "655a:656akvxyz:"
	spec += "656a:657avxyz:"
	spec += "657a:658ab:"
	spec += "658a:662abcdefgh:"
	spec += "690a:690abcdevxyz"
	return bib.MarcValuesTrim(spec)
}

func (bib Bib) BookplateCodes() []string {
	values := []string{}
	for _, item := range bib.Items {
		arrayAppend(&values, item.BookplateCodes())
	}
	// TODO: Do we need this? It seems that the data has been
	// consolidate in the item records.
	// arrayAppend(&values, MarcValues("935a"))
	return values
}
