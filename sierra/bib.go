package sierra

import (
	"encoding/json"
	"fmt"
	"log"
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

func (b Bib) log(show bool, msg string) {
	if show {
		log.Printf(fmt.Sprintf("%s", msg))
	}
}

func (b Bib) Bib() string {
	return "b" + b.Id
}

func (bib Bib) VernacularValuesByField(specsStr string, join bool) [][]string {
	values := [][]string{}
	for _, spec := range NewFieldSpecs(specsStr) {
		for _, field := range bib.VarFields {
			if field.MarcTag == spec.MarcTag {
				for _, vernValues := range bib.VernacularValuesFor(field, spec, join) {
					values = append(values, vernValues)
				}
			}
		}
	}
	return values
}

func (bib Bib) VernacularValuesFor(field Field, spec FieldSpec, join bool) [][]string {
	values := [][]string{}

	// True if the field has subfield with tag 6
	// target would be "880-04"
	vern, target := field.HasVernacular()
	if !vern {
		return values
	}

	tokens := strings.Split(target, "-")    // ["880", "04"]
	marcTag := tokens[0]                    // "880"
	tag6 := field.MarcTag + "-" + tokens[1] // "700-04"

	// Process the fields indicated in target (e.g. 880s)
	for _, vernField := range bib.getFields(marcTag) {
		// if this is the one that corresponds with our target
		// e.g. 700-04
		if vernField.IsVernacularFor(tag6) {
			vernValues := vernField.Values(spec.Subfields, join)
			values = append(values, vernValues)
		}
	}
	return values
}

func (bib Bib) MarcValue(specsStr string, trim bool) string {
	values := bib.MarcValuesByField(specsStr, true)
	return valuesToString(values, trim)
}

func (bib Bib) MarcValuesByField(specsStr string, join bool) [][]string {
	values := [][]string{}
	vernProcessed := []string{}
	specs := NewFieldSpecs(specsStr)
	for _, spec := range specs {

		fields := bib.getFields(spec.MarcTag)
		if len(spec.Subfields) == 0 {
			// Get the value directly
			for _, field := range fields {
				if field.Content != "" {
					values = append(values, []string{field.Content})
				}
			}
			continue
		}

		// Process the subfields
		for _, field := range fields {
			subValues := field.Values(spec.Subfields, join)
			fieldValues := []string{}
			for _, subValue := range subValues {
				safeAppend(&fieldValues, subValue)
			}
			if len(fieldValues) > 0 {
				values = append(values, fieldValues)
			}
		}

		// Gather the vernacular values for the fields
		for _, field := range fields {
			fieldValues := []string{}
			for _, vernValues := range bib.VernacularValuesFor(field, spec, join) {
				for _, vernValue := range vernValues {
					safeAppend(&fieldValues, vernValue)
				}
			}
			if len(fieldValues) > 0 {
				vernProcessed = append(vernProcessed, field.MarcTag)
				values = append(values, fieldValues)
			}
		}
	}

	// Process the 880 fields again this time to gather vernacular
	// values for fields in the spec that have no values in the
	// record (e.g. we might have an 880 for field 505, but no 505
	// value in the record, or an 880 for field 490a but no 409a
	// on the record)
	f880s := bib.getFields("880")
	for _, spec := range specs {
		for _, f880 := range f880s {
			if f880.IsVernacularFor(spec.MarcTag) && !in(vernProcessed, spec.MarcTag) {
				fieldValues := []string{}
				for _, vernValue := range f880.Values(spec.Subfields, join) {
					safeAppend(&fieldValues, vernValue)
				}
				if len(fieldValues) > 0 {
					values = append(values, fieldValues)
				}
			}
		}
	}

	return values
}

// fieldSpec is something in the form "nnna" where "nnn" is the tag of the
// field and "a" represents the subfields. For example: "100ac" means
// field "100" subfields "a" and "c". Multiple fields can be indicated
// separated by colons, for example: "100ac:210f"
func (bib Bib) MarcValues(fieldSpec string) []string {
	values := []string{}
	for _, valuesForField := range bib.MarcValuesByField(fieldSpec, true) {
		valuesStr := strings.Join(valuesForField, " ")
		values = append(values, valuesStr)
	}
	return values
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

func valuesToArray(values [][]string, trim bool, join bool) []string {
	array := []string{}
	for _, fieldValues := range values {
		if join {
			// join all the values for the field as a single element in
			// the returning array
			value := strings.Join(fieldValues, " ")
			if trim {
				value = trimPunct(value)
			}
			safeAppend(&array, value)
		} else {
			// preserve each individual value (regardless of their field)
			// as a single element in the returning arrray
			for _, value := range fieldValues {
				if trim {
					value = trimPunct(value)
				}
				safeAppend(&array, value)
			}
		}
	}
	return array
}

func valuesToString(values [][]string, trim bool) string {
	rowValues := []string{}
	for _, fieldValues := range values {
		safeAppend(&rowValues, strings.Join(fieldValues, " "))
	}
	if trim {
		return trimPunct(strings.Join(rowValues, " "))
	}
	return strings.Join(rowValues, " ")
}

/*
 * Author functions
 */
func (bib Bib) AuthorsAddlT() []string {
	authors := bib.MarcValuesByField("700aqbcd:710abcd:711aqbcde:810abc:811aqdce", true)
	return valuesToArray(authors, false, true)
}

func (bib Bib) AuthorsT() []string {
	authors := bib.MarcValuesByField("100abcdq:110abcd:111abcdeq", true)
	return valuesToArray(authors, true, true)
}

func (bib Bib) AuthorsAddlDisplay() []string {
	authors := bib.MarcValuesByField("700abcd:710ab:711ab", true)
	return valuesToArray(authors, true, true)
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

	values := bib.MarcValuesByField(specStr, true)
	return valuesToArray(values, true, true)
}

func (bib Bib) AuthorDisplay() string {
	authors := bib.MarcValues("100abcdq:110abcd:111abcd")
	if len(authors) > 0 {
		return trimPunct(authors[0])
	}
	return ""
}

func (bib Bib) AuthorVernacularDisplay() string {
	vernAuthors := bib.VernacularValuesByField("100abcdq:110abcd:111abcd", false)
	return valuesToString(vernAuthors, true)
}

func (bib Bib) AbstractDisplay() string {
	values := bib.MarcValues("520a")
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

/*
 * Title functions
 */
func (bib Bib) UniformTitles(newVersion bool) []UniformTitles {
	var spec string
	if newVersion {
		spec = "240adfgklmnoprs"
	} else {
		spec = "130adfgklmnoprst"
	}

	titlesArray := []UniformTitles{}
	for _, valuesForField := range bib.MarcValuesByField(spec, false) {
		titles := UniformTitles{}
		query := ""
		for _, value := range valuesForField {
			// TODO: revisit this. Is it OK to periods to vernacular values?
			display := addPeriod(value)
			if query == "" {
				query = display
			} else {
				query = query + " " + display
			}
			title := UniformTitle{Display: display, Query: query}
			titles.Title = append(titles.Title, title)
		}
		if len(titles.Title) > 0 {
			titlesArray = append(titlesArray, titles)
		}
	}
	return titlesArray
}

func (bib Bib) UniformTitlesDisplay(newVersion bool) string {
	titles := bib.UniformTitles(newVersion)
	bytes, err := json.Marshal(titles)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func (bib Bib) TitleDisplay() string {
	values := bib.MarcValuesByField("245apbfgkn", true)
	titles := valuesToArray(values, true, true)
	if len(titles) > 0 {
		return titles[0]
	}
	return ""
}

func (bib Bib) TitleT() []string {
	specsStr := "100tflnp:110tflnp:111tfklpsv:130adfklmnoprst:210ab:222ab:"
	specsStr += "240adfklmnoprs:242abnp:246abnp:247abnp:505t:"
	specsStr += "700fklmtnoprsv:710fklmorstv:711fklpt:730adfklmnoprstv:740ap"
	values := bib.MarcValuesByField(specsStr, true)
	titles := valuesToArray(values, true, false)
	return titles
}

func (bib Bib) TitleSeries() []string {
	specsStr := "400flnptv:410flnptv:411fklnptv:440ap:490a:800abcdflnpqt:"
	specsStr += "810tflnp:811tfklpsv:830adfklmnoprstv"
	values := bib.MarcValuesByField(specsStr, true)
	series := valuesToArray(values, true, false)
	return series
}

func (bib Bib) TitleVernacularDisplay() string {
	vernTitles := bib.VernacularValuesByField("245apbfgkn", false)
	return valuesToString(vernTitles, true)
}

/*
 * Others
 */
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
	values := bib.MarcValuesByField("050ab:090ab:091ab:092ab:096ab:099ab", true)
	return valuesToArray(values, true, true)
}

func (bib Bib) TopicFacet() []string {
	values := bib.MarcValuesByField("650a:690a", true)
	return valuesToArray(values, true, false)
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
	values := bib.MarcValuesByField(spec, true)
	return valuesToArray(values, true, true)
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

func (bib Bib) Isbn() []string {
	values := bib.MarcValuesByField("020a:020z", true)
	return valuesToArray(values, true, true)
}

func (bib Bib) PublishedDisplay() []string {
	// More than one "a" subfield can exists on the same 260
	// therefore we can get a single field with multiple values.
	// Here we break each value on its own.
	values := []string{}
	for _, fieldValues := range bib.MarcValuesByField("260a", true) {
		for _, value := range fieldValues {
			safeAppend(&values, trimPunct(value))
		}
	}
	return values
}

func (bib Bib) PublishedVernacularDisplay() string {
	vernPub := bib.VernacularValuesByField("260a", false)
	return valuesToString(vernPub, false)
}

func (bib Bib) IsDissertaion() bool {
	subs := []string{"a", "c"}
	for _, field := range bib.getFields("502") {
		for _, value := range field.Values(subs, true) {
			if strings.Contains(strings.ToLower(value), "brown univ") {
				return true
			}
		}
	}
	return false
}

func (bib Bib) Issn() []string {
	values := bib.MarcValuesByField("022a:022l:022y:773x:774x:776x", true)
	return valuesToArray(values, true, true)
}

func (bib Bib) PublicationYear() (int, bool) {
	rangeStart := 500
	rangeEnd := time.Now().Year()
	tolerance := 15

	f008 := bib.MarcValue("008", true)
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
	f260c := bib.MarcValue("260c", true)
	re := regexp.MustCompile("(\\d{4})")
	year := re.FindString(f260c)
	return toIntTry(year)
}

func (bib Bib) OclcNum() []string {
	// RegEx based on Traject's marc21.rb
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21_semantics.rb
	re := regexp.MustCompile("\\s*(ocm|ocn|on|\\(OCoLC\\))(\\d+)")
	values := bib.MarcValuesByField("001:035a:035z", false)
	nums := []string{}
	for _, value := range valuesToArray(values, false, false) {
		if strings.HasPrefix(value, "ssj") {
			// TODO: Ask Jeanette about these values
			// eg. b4643178, b4643180
		} else {
			num := strings.TrimSpace(re.ReplaceAllString(value, "$2"))
			safeAppend(&nums, num)
		}
	}
	return nums
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
	f008 := bib.MarcValue("008", false)
	f008_lang := ""
	if len(f008) > 38 {
		f008_lang = languageName(f008[35:38])
		safeAppend(&values, f008_lang)
	}

	for _, valuesByField := range bib.MarcValuesByField("041a:041d:041e:041j", true) {
		for _, value := range valuesByField {
			language := languageName(value)
			safeAppend(&values, language)
		}
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

	// TODO: rewrite this using bib.MarcValuesByField()
	//
	// Notice that we don't use bib.MarcValues() here because
	// bib.MarcValues() returns the data without a relationship
	// to the field where each value was found. In this case
	// we care about what values are found on each instance of
	// the field so that we can concatenate "region (parent region)"
	// values if they are found in a specific field.
	for _, zFieldSpec := range zFieldSpecs {
		spec, _ := NewFieldSpec(zFieldSpec)
		for _, field := range bib.getFields(spec.MarcTag) {
			subValues := field.Values(spec.Subfields, true)
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
