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
	Id              string              `json:"id"`
	UpdatedDateTime string              `json:"updatedDate,omitempty"`
	CreatedDate     string              `json:"createdDate,omitempty"`
	DeletedDate     string              `json:"deletedDate,omitempty"`
	Deleted         bool                `json:"deleted,omitempty"`
	Suppressed      bool                `json:"suppressed,omitempty"`
	Available       bool                `json:"available,omitempty"`
	Lang            map[string]string   `json:"lang,omitempty"`
	Title           string              `json:"title,omitempty"`
	Author          string              `json:"author,omitempty"`
	MaterialType    map[string]string   `json:"materialType,omitempty"`
	BibLevel        map[string]string   `json:"bibLevel,omitempty"`
	PublishYear     int                 `json:"publishYear,omitempty"`
	CatalogDate     string              `json:"catalogDate,omitempty"`
	Country         map[string]string   `json:"country,omitempty"`
	NormTitle       string              `json:"normTitle,omitempty"`
	NormAuthor      string              `json:"normAuthor,omitempty"`
	Locations       []map[string]string `json:"locations,omitempty"`
	VarFields       MarcFields          `json:"varFields,omitempty"`
	Items           []Item              // does not come on the Sierra response
	hasMarc         string              // does not come on the Sierra response
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
	fields := bib.VarFields.FieldValues("700aqbcd:710abcd:711aqbcde:810abc:811aqdce")
	authors := fields.ToArrayJoin()
	return dedupArray(authors)
}

func (bib Bib) AuthorsT() []string {
	if !bib.HasMarc() {
		value := trimPunct(bib.VarFields.getFieldTagContent("a"))
		return []string{value}
	}
	fields := bib.VarFields.FieldValues("100abcdq:110abcd:111abcdeq")
	authors := fields.ToArray()
	return dedupArray(authors)
}

func (bib Bib) AuthorsAddlDisplay() []string {
	fields := bib.VarFields.FieldValues("700abcd:710ab:711ab")
	authors := fields.ToArray()
	return dedupArray(authors)
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

	fields := bib.VarFields.FieldValues(specStr)
	return fields.ToArray()
}

func (bib Bib) AuthorDisplay() string {
	if !bib.HasMarc() {
		return trimPunct(bib.VarFields.getFieldTagContent("a"))
	}

	fields := bib.VarFields.FieldValues("100abcdq:110abcd:111abcd")
	authors := fields.ToArray()
	if len(authors) > 0 {
		return authors[0]
	}
	return ""
}

func (bib Bib) AuthorVernacularDisplay() string {
	fields := bib.VarFields.VernacularValues("100abcdq:110abcd:111abcd")
	authors := fields.ToArray()
	return strings.Join(authors, " ")
}

func (bib Bib) AbstractDisplay() string {
	values := bib.VarFields.FieldValues("520a")
	if len(values) > 0 {
		return values[0].String()
	}
	return ""
}

/*
 * Uniform Related Works
 */
func (bib Bib) UniformRelatedWorks() string {
	works := []UniformRelatedWorks{}

	for _, work := range bib.relatedWorksForField("730", "", "adfgklmnoprst") {
		works = append(works, work)
	}

	// TODO use "abcdqu"
	for _, work := range bib.relatedWorksForField("700", "abcd", "fklmnoprstv") {
		works = append(works, work)
	}

	// TODO use "abcdgnu"
	for _, work := range bib.relatedWorksForField("710", "abcdg", "fklmorstv") {
		works = append(works, work)
	}

	// TODO use "abcdgnu"
	for _, work := range bib.relatedWorksForField("711", "abcdg", "fklmorstv") {
		works = append(works, work)
	}

	str, _ := toJSON(works)
	return str
}

func (bib Bib) relatedWorksForField(marcTag, authorSubs, titleSubs string) []UniformRelatedWorks {
	works := []UniformRelatedWorks{}
	authorSubfields := stringToArray(authorSubs)
	titleSubfields := stringToArray(titleSubs)
	spec := marcTag + authorSubs + titleSubs
	for _, field := range bib.VarFields.FieldValues(spec) {
		authors := field.Values(authorSubfields)
		author := trimDot(trimPunct(authors.String()))
		titles := field.Values(titleSubfields)
		if len(titles.Subfields) > 0 {
			// TODO: Ask Jeanette why we don't process empty titles
			// (but we process empty authors)
			work := UniformRelatedWorks{Author: author}
			query := ""
			for i, titleStr := range titles.Strings() {
				titleDisplay := addPeriod(titleStr)
				if i == 0 {
					query = titleDisplay
				} else {
					query += " " + titleDisplay
				}
				title := UniformTitle{
					Display: titleDisplay,
					Query:   query,
				}
				work.Titles = append(work.Titles, title)
			}
			works = append(works, work)
		}
	}
	return works
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
	for _, field := range bib.VarFields.FieldValues(spec) {
		titles := UniformTitles{}
		query := ""
		for _, sub := range field.Subfields {
			// TODO: revisit this. Is it OK to periods to vernacular values?
			value := sub["content"]
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
	fields := bib.VarFields.FieldValues("245apbfgkn")
	titles := fields.ToArray()
	if len(titles) > 0 {
		return titles[0]
	}
	return ""
}

func (bib Bib) TitleT() []string {
	specsStr := "100tflnp:110tflnp:111tfklpsv:130adfklmnoprst:210ab:222ab:"
	specsStr += "240adfklmnoprs:242abnp:246abnp:247abnp:"
	specsStr += "700fklmtnoprsv:710fklmorstv:711fklpt:730adfklmnoprstv:740ap"
	fields := bib.VarFields.FieldValues(specsStr)
	titles := fields.ToArray()

	// We handle table of content titles (505t) separate because we want
	// each of them as a single item rather then combined.
	// Traject does this via the "joinable?" flag that automatically
	// treats single subfield specs as not joinable.
	for _, f505t := range bib.VarFields.FieldValues("505t") {
		arrayAppend(&titles, f505t.StringsTrim())
	}
	return titles
}

func (bib Bib) TitleSeries() []string {
	specsStr := "400flnptv:410flnptv:411fklnptv:440ap:800abcdflnpqt:"
	specsStr += "810tflnp:811tfklpsv:830adfklmnoprstv"
	fields := bib.VarFields.FieldValues(specsStr)
	titles := fields.ToArray()
	// Special treatment for 490 because we want each "a" value
	// on its own.
	for _, f490a := range bib.VarFields.FieldValues("490a") {
		arrayAppend(&titles, f490a.StringsTrim())
	}
	return titles
}

func (bib Bib) TitleVernacularDisplay() string {
	fields := bib.VarFields.VernacularValues("245apbfgkn")
	titles := fields.ToArray()
	return strings.Join(titles, " ")
}

func (bib Bib) SortableTitle() string {
	if !bib.HasMarc() {
		return strings.TrimSpace(trimPunct(bib.VarFields.getFieldTagContent("t")))
	}

	// Logic stolen from
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21_semantics.rb
	// TODO do we need the field k logic here?

	// Use MarcValuesByField because we don't want to trim the value prematurely.
	// We trim at the end _after_ processing ind2.
	titles := bib.VarFields.FieldValues("245ab")
	if len(titles) == 0 || titles[0].String() == "" {
		return ""
	}

	sortTitle := titles[0].String()
	fields := bib.VarFields.getFields("245")
	if len(fields) > 0 {
		ind2 := toInt(fields[0].Ind2)
		if ind2 > 0 && len(sortTitle) > ind2 {
			// drop the prefix as noted in the second indicator
			sortTitle = sortTitle[ind2:len(sortTitle)]
		}
	}
	return strings.TrimSpace(trimPunct(sortTitle))
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
	fields := bib.VarFields.FieldValues("050ab:090ab:091ab:092ab:096ab:099ab")
	return fields.ToArray()
}

func (bib Bib) TopicFacet() []string {
	fields := bib.VarFields.FieldValues("650a:690a")
	return fields.ToArrayTrim()
}

func (bib Bib) Subjects() []string {
	spec := "600abcdefghjklmnopqrstuvxyz:"
	spec += "610abcdefghklmnoprstuvxyz:"
	spec += "611acdefghjklnpqstuvxyz:"
	spec += "630adefghklmnoprstvxyz:"
	spec += "648avxyz:"
	spec += "650abcdezxvy:"
	spec += "651aexzvy:"
	spec += "654abevyz:"
	spec += "655abvxyz:"
	spec += "656akvxyz:"
	spec += "657avxyz:"
	spec += "658ab:"
	spec += "662abcdefgh:"
	spec += "690abcdevxyz"
	fields := bib.VarFields.FieldValues(spec)
	subjects := fields.ToArray()

	specA := "600a:610a:611a:630a:648a:650a:651a:653a:"
	specA += "654a:655a:656a:657a:658a:690a"
	for _, field := range bib.VarFields.FieldValues(specA) {
		arrayAppend(&subjects, field.StringsTrim())
	}
	return subjects
}

func (bib Bib) BookplateCodes() []string {
	// The items inside the bib record have information about the text and URL for
	// bookplates in their own 856uz. We sort of had this informastion before at
	// the bib level in 996uz but it was not linked to each item.
	// Example: b6177452.
	values := []string{}
	for _, item := range bib.Items {
		arrayAppend(&values, item.BookplateCodes())
	}
	fields := bib.VarFields.FieldValues("935a")
	arrayAppend(&values, fields.ToArrayTrim())
	return values
}

func (bib Bib) Isbn() []string {
	fields := bib.VarFields.FieldValues("020a:020z")
	return fields.ToArray()
}

func (bib Bib) PublishedDisplay() []string {
	if !bib.HasMarc() {
		value := trimPunct(bib.VarFields.getFieldTagContent("p"))
		return []string{value}
	}

	// More than one "a" subfield can exists on the same 260
	// therefore we can get a single field with multiple values.
	// Here we break each value on its own.
	fields := bib.VarFields.FieldValues("260a")
	return fields.ToArrayTrim()
}

func (bib Bib) PublishedVernacularDisplay() string {
	fields := bib.VarFields.VernacularValues("260a")
	published := fields.ToArrayJoin()
	return strings.Join(published, " ")
}

func (bib Bib) IsDissertaion() bool {
	for _, field := range bib.VarFields.FieldValues("502ac") {
		value := field.String()
		if strings.Contains(strings.ToLower(value), "brown univ") {
			return true
		}
	}
	return false
	//
	// subs := []string{"a", "c"}
	// for _, field := range bib.VarFields.getFields("502") {
	// 	for _, value := range field.Values(subs) {
	// 		if strings.Contains(strings.ToLower(value), "brown univ") {
	// 			return true
	// 		}
	// 	}
	// }
	// return false
}

func (bib Bib) PhysicalDisplay() []string {
	displays := []string{}
	for _, field := range bib.VarFields.FieldValues("300abcefg:530abcd") {
		display := addPeriod(field.String())
		displays = append(displays, display)
	}
	return displays
}

func (bib Bib) Issn() []string {
	fields := bib.VarFields.FieldValues("022a:022l:022y:773x:774x:776x")
	return fields.ToArray()
}

func (bib Bib) PublicationYear() (int, bool) {
	rangeStart := 500
	rangeEnd := time.Now().Year()
	tolerance := 15

	f008 := bib.VarFields.ControlValue("008")
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
	fields := bib.VarFields.FieldValues("260c")
	for _, value := range fields.ToArrayTrim() {
		re := regexp.MustCompile("(\\d{4})")
		year := re.FindString(value)
		return toIntTry(year)
	}
	return 0, false
}

func (bib Bib) OclcNum() []string {
	nums := []string{}

	f001 := bib.VarFields.ControlValue("001")
	safeAppend(&nums, cleanOclcNum(f001))

	fields := bib.VarFields.FieldValues("035a:035z")
	for _, value := range fields.ToArrayRaw() {
		safeAppend(&nums, cleanOclcNum(value))
	}
	return nums
}

func cleanOclcNum(value string) string {
	// RegEx based on Traject's marc21.rb
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21_semantics.rb
	re := regexp.MustCompile("\\s*(ocm|ocn|on|\\(OCoLC\\))(\\d+)")
	if strings.HasPrefix(value, "ssj") {
		// TODO: Ask Jeanette about these values
		// eg. b4643178, b4643180
		return ""
	}
	return strings.TrimSpace(re.ReplaceAllString(value, "$2"))
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

	f338s := bib.VarFields.FieldValues("338a")
	for _, value := range f338s.ToArrayTrim() {
		if value == "online resource" {
			return true
		}
	}

	// Using this instead of the 998 logic below
	f300s := bib.VarFields.FieldValues("300a")
	for _, value := range f300s.ToArrayTrim() {
		if strings.Contains(value, "online resource") {
			return true
		}
	}

	// We don't get the MARC 998a field, but the Location has the equivalent value
	for _, location := range bib.Locations {
		if location["code"] == "es001" {
			return true
		}
	}

	// // It seems that field 998 does not come in the API and
	// // therfore this code does nothing for now.
	// for _, value := range bib.VarFields.MarcValues("998a", true) {
	// 	if value == "es001" {
	// 		return true
	// 	}
	// }
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
		for _, value := range bib.VarFields.ControlValues("007") {
			if strings.Contains(value, "v") || strings.Contains(value, "m") {
				return "BV" // video
			}
		}
	}
	return code
}

func (bib Bib) IsDissertation() bool {
	fields := bib.VarFields.FieldValues("502ac")
	for _, value := range fields.ToArrayRaw() {
		if strings.Contains(strings.ToLower(value), "brown univ") {
			return true
		}
	}
	return false
}

func (bib Bib) Languages() []string {
	values := []string{}
	f008 := bib.VarFields.ControlValue("008")
	f008_lang := ""
	if len(f008) > 38 {
		f008_lang = languageName(f008[35:38])
		safeAppend(&values, f008_lang)
	}

	for _, valuesByField := range bib.VarFields.FieldValues("041a:041d:041e:041j") {
		for _, value := range valuesByField.Strings() {
			langs := languageNames(value)
			arrayAppend(&values, langs)
		}
	}
	return values
}

func (bib Bib) UrlDisplay(specStr string) []string {
	fields := bib.VarFields.FieldValues(specStr)
	return fields.ToArrayRaw()
}

func (bib Bib) RegionFacet() []string {
	// Stolen from Traject's marc_geo_facet
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21_semantics.rb
	values := []string{}
	f043s := bib.VarFields.FieldValues("043a")
	for _, value := range f043s.ToArrayTrim() {
		code := trimPunct(value)
		code = strings.TrimRight(code, "-")
		name := regionName(code)
		safeAppend(&values, name)
	}

	f6XXs := bib.VarFields.FieldValues("651a:691a")
	for _, value := range f6XXs.ToArrayTrim() {
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
	for _, fieldValues := range bib.VarFields.FieldValues(zFieldSpecs) {
		regions := fieldValues.Strings()
		if len(regions) == 2 {
			// Asumme the first one is the parent region of the second one
			// e.g. v[0] := "USA", v[1] := "Rhode Island (USA)"
			parentRegion := trimPunct(regions[0])
			region := trimPunct(regions[1]) + " (" + parentRegion + ")"
			safeAppend(&values, parentRegion)
			safeAppend(&values, region)
		} else {
			arrayAppend(&values, regions)
		}
	}
	return values
}
