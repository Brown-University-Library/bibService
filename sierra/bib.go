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
	VarFields       MarcFields        `json:"varFields,omitempty"`
	Items           []Item            // does not come on the Sierra response
	hasMarc         string            // does not come on the Sierra response
}

func (b Bib) log(show bool, msg string) {
	if show {
		log.Printf(fmt.Sprintf("%s", msg))
	}
}

func (b Bib) Bib() string {
	return "b" + b.Id
}

func (bib Bib) HasMarc() bool {
	if bib.hasMarc == "" {
		if bib.VarFields.hasMarc() {
			bib.hasMarc = "Y"
		} else {
			bib.hasMarc = "N"
		}
	}
	return bib.hasMarc == "Y"
}

/*
 * Author functions
 */
func (bib Bib) AuthorsAddlT() []string {
	authors := bib.VarFields.MarcValuesByField("700aqbcd:710abcd:711aqbcde:810abc:811aqdce", true)
	return valuesToArray(authors, false, true)
}

func (bib Bib) AuthorsT() []string {
	if !bib.HasMarc() {
		value := trimPunct(bib.VarFields.getFieldTagContent("a"))
		return []string{value}
	}
	authors := bib.VarFields.MarcValuesByField("100abcdq:110abcd:111abcdeq", true)
	return valuesToArray(authors, true, true)
}

func (bib Bib) AuthorsAddlDisplay() []string {
	authors := bib.VarFields.MarcValuesByField("700abcd:710ab:711ab", true)
	return valuesToArray(authors, true, true)
}

func (bib Bib) AuthorFacet() []string {
	specStr := "100abcd:110ab:111ab:700abcd:711ab"

	f710 := bib.VarFields.getFields("710")
	if len(f710) > 0 {
		// If there is more than one 710 field this will only check the first one.
		// TODO: handle multi 710 fields
		if f710[0].Ind2 != "9" {
			specStr += ":710ab"
		}
	}

	return bib.VarFields.MarcValues(specStr, true)
}

func (bib Bib) AuthorDisplay() string {
	if !bib.HasMarc() {
		return trimPunct(bib.VarFields.getFieldTagContent("a"))
	}

	values := bib.VarFields.MarcValuesByField("100abcdq:110abcd:111abcd", true)
	authors := valuesToArray(values, true, true)
	if len(authors) > 0 {
		return authors[0]
	}
	return ""
}

func (bib Bib) AuthorVernacularDisplay() string {
	vernAuthors := bib.VarFields.VernacularValuesByField("100abcdq:110abcd:111abcd")
	return valuesToString(vernAuthors, true)
}

func (bib Bib) AbstractDisplay() string {
	values := bib.VarFields.MarcValues("520a", true)
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
	for _, valuesForField := range bib.VarFields.MarcValuesByField(spec, false) {
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
	if !bib.HasMarc() {
		return trimPunct(bib.VarFields.getFieldTagContent("t"))
	}
	values := bib.VarFields.MarcValuesByField("245apbfgkn", true)
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
	values := bib.VarFields.MarcValuesByField(specsStr, true)
	titles := valuesToArray(values, true, false)
	return titles
}

func (bib Bib) TitleSeries() []string {
	specsStr := "400flnptv:410flnptv:411fklnptv:440ap:490a:800abcdflnpqt:"
	specsStr += "810tflnp:811tfklpsv:830adfklmnoprstv"
	values := bib.VarFields.MarcValuesByField(specsStr, true)
	series := valuesToArray(values, true, false)
	return series
}

func (bib Bib) TitleVernacularDisplay() string {
	vernTitles := bib.VarFields.VernacularValuesByField("245apbfgkn")
	return valuesToString(vernTitles, true)
}

func (bib Bib) SortableTitle() string {
	if !bib.HasMarc() {
		return trimPunct(bib.VarFields.getFieldTagContent("t"))
	}

	// Logic stolen from
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21_semantics.rb
	// TODO do we need the field k logic here?
	titles := bib.VarFields.MarcValues("245ab", true)
	if len(titles) == 0 {
		return ""
	}

	sortTitle := titles[0]
	fields := bib.VarFields.getFields("245")
	if len(fields) > 0 {
		ind2 := toInt(fields[0].Ind2)
		if ind2 > 0 && len(sortTitle) > ind2 {
			// drop the prefix as noted in the second indicator
			sortTitle = sortTitle[ind2:len(sortTitle)]
		}
	}
	return trimPunct(sortTitle)
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

func (bib Bib) CallNumbers() []string {
	values := bib.VarFields.MarcValuesByField("050ab:090ab:091ab:092ab:096ab:099ab", true)
	return valuesToArray(values, true, true)
}

func (bib Bib) TopicFacet() []string {
	values := bib.VarFields.MarcValuesByField("650a:690a", true)
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
	subjects := bib.VarFields.MarcValuesByField(spec, true)
	return valuesToArray(subjects, true, false)
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
	values := bib.VarFields.MarcValuesByField("020a:020z", true)
	return valuesToArray(values, true, true)
}

func (bib Bib) PublishedDisplay() []string {
	if !bib.HasMarc() {
		value := trimPunct(bib.VarFields.getFieldTagContent("p"))
		return []string{value}
	}

	// More than one "a" subfield can exists on the same 260
	// therefore we can get a single field with multiple values.
	// Here we break each value on its own.
	values := []string{}
	for _, fieldValues := range bib.VarFields.MarcValuesByField("260a", true) {
		for _, value := range fieldValues {
			safeAppend(&values, trimPunct(value))
		}
	}
	return values
}

func (bib Bib) PublishedVernacularDisplay() string {
	vernPub := bib.VarFields.VernacularValuesByField("260a")
	return valuesToString(vernPub, false)
}

func (bib Bib) IsDissertaion() bool {
	subs := []string{"a", "c"}
	for _, field := range bib.VarFields.getFields("502") {
		for _, value := range field.Values(subs, true) {
			if strings.Contains(strings.ToLower(value), "brown univ") {
				return true
			}
		}
	}
	return false
}

func (bib Bib) Issn() []string {
	values := bib.VarFields.MarcValuesByField("022a:022l:022y:773x:774x:776x", true)
	return valuesToArray(values, true, true)
}

func (bib Bib) PublicationYear() (int, bool) {
	rangeStart := 500
	rangeEnd := time.Now().Year()
	tolerance := 15

	f008 := bib.VarFields.MarcValue("008", true)
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
	f260c := bib.VarFields.MarcValue("260c", true)
	re := regexp.MustCompile("(\\d{4})")
	year := re.FindString(f260c)
	return toIntTry(year)
}

func (bib Bib) OclcNum() []string {
	// RegEx based on Traject's marc21.rb
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21_semantics.rb
	re := regexp.MustCompile("\\s*(ocm|ocn|on|\\(OCoLC\\))(\\d+)")
	values := bib.VarFields.MarcValuesByField("001:035a:035z", true)
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

	for _, value := range bib.VarFields.MarcValues("338a", true) {
		if value == "online resource" {
			return true
		}
	}

	// Using this instead of the 998 logic below
	for _, value := range bib.VarFields.MarcValues("300a", true) {
		if strings.Contains(value, "online resource") {
			return true
		}
	}

	// It seems that field 998 does not come in the API and
	// therfore this code does nothing for now.
	for _, value := range bib.VarFields.MarcValues("998a", true) {
		if value == "es001" {
			return true
		}
	}
	return false
}

func (bib Bib) Format() string {
	return formatName(bib.FormatCode())
}

func (bib Bib) FormatCode() string {
	// Logic from bul_format.rb
	if bib.IsDissertation() {
		return "BTD"
	}

	leader := bib.VarFields.Leader()
	code := formatCode(leader)
	if code == "VM" {
		for _, value := range bib.VarFields.MarcValues("007", false) {
			if strings.Contains(value, "v") || strings.Contains(value, "m") {
				return "BV" // video
			}
		}
	}
	return code
}

func (bib Bib) IsDissertation() bool {
	for _, value := range bib.VarFields.MarcValues("502ac", false) {
		if strings.HasPrefix(strings.ToLower(value), "brown univ") {
			return true
		}
	}
	return false
}

func (bib Bib) Languages() []string {
	values := []string{}
	f008 := bib.VarFields.MarcValue("008", false)
	f008_lang := ""
	if len(f008) > 38 {
		f008_lang = languageName(f008[35:38])
		safeAppend(&values, f008_lang)
	}

	for _, valuesByField := range bib.VarFields.MarcValuesByField("041a:041d:041e:041j", true) {
		for _, value := range valuesByField {
			langs := languageNames(value)
			arrayAppend(&values, langs)
		}
	}
	return values
}

func (bib Bib) UrlDisplay(specStr string) []string {
	values := bib.VarFields.MarcValuesByField(specStr, false)
	return valuesToArray(values, false, false)
}

func (bib Bib) RegionFacet() []string {
	// Stolen from Traject's marc_geo_facet
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21_semantics.rb
	values := []string{}
	for _, value := range bib.VarFields.MarcValues("043a", true) {
		code := trimPunct(value)
		code = strings.TrimRight(code, "-")
		name := regionName(code)
		safeAppend(&values, name)
	}

	for _, value := range bib.VarFields.MarcValues("651a:691a", true) {
		trimVal := trimPunct(value)
		safeAppend(&values, trimVal)
	}

	arrayAppend(&values, bib.RegionFacetZFields())
	return values
}

func (bib Bib) RegionFacetZFields() []string {
	values := []string{}
	zFieldSpecs := "600z:610z:611z:630z:648z:650z:"
	zFieldSpecs += "654z:655z:656z:690z:651z:691z"
	for _, fieldValues := range bib.VarFields.MarcValuesByField(zFieldSpecs, true) {
		if len(fieldValues) == 2 {
			// Asumme the first one is the parent region of the second one
			// e.g. v[0] := "USA", v[1] := "Rhode Island (USA)"
			parentRegion := trimPunct(fieldValues[0])
			region := trimPunct(fieldValues[1]) + " (" + parentRegion + ")"
			safeAppend(&values, parentRegion)
			safeAppend(&values, region)
		} else {
			arrayAppend(&values, fieldValues)
		}
	}
	return values
}
