package sierra

import (
	"regexp"
	"strings"
	"time"
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
			if len(spec.Subfields) == 1 {
				// single subfields specified (060a)
				// append each individual value
				for _, subValue := range subValues {
					if !in(values, subValue) {
						values = append(values, subValue)
					}
				}
			} else {
				// multi-subfields specified (e.g. 060abc)
				// concatenate the values and then append them
				strVal := strings.Join(subValues, " ")
				if strVal != "" && !in(values, strVal) {
					values = append(values, strVal)
				}
			}
		}
	}
	return values
}

func (bib BibResp) MarcValuesTrim(fieldSpec string) []string {
	values := []string{}
	for _, value := range bib.MarcValues(fieldSpec) {
		trimValue := trimPunct(value)
		if !in(values, trimValue) {
			values = append(values, trimValue)
		}
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
			if fieldSub["tag"] == subfield && fieldSub["content"] != "" {
				values = append(values, fieldSub["content"])
			}
		}
	}
	return values
}

// TODO: change MarcValues to omit empty and duplicate values
// since we do that all over the place
func (bib BibResp) Isbn() []string {
	values := []string{}
	for _, value := range bib.MarcValues("020a:020z") {
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func (bib BibResp) Issn() []string {
	values := []string{}
	for _, value := range bib.MarcValues("022a:022l:022y:773x:774x:776x") {
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func (bib BibResp) PublicationYear() (int, bool) {
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

func (bib BibResp) pubYear260() (int, bool) {
	f260c := bib.MarcValue("260c")
	re := regexp.MustCompile("(\\d{4})")
	year := re.FindString(f260c)
	return toIntTry(year)
}

func (bib BibResp) OclcNum() []string {
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
			if num != "" && !in(values, num) {
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
	for _, item := range bib.Items {
		if strings.HasPrefix(item.Location["code"], "es") {
			return true
		}
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

func (bib BibResp) Languages() []string {
	codes := []string{}

	// 008[35-37]:041a:041d:041e:041j
	f008 := bib.MarcValue("008")
	f008_lang := ""
	if len(f008) > 38 {
		f008_lang = f008[35:38]
		if f008_lang != "" {
			codes = append(codes, f008_lang)
		}
	}

	for _, value := range bib.MarcValues("041a:041d:041e:041j") {
		if value != f008_lang {
			codes = append(codes, value)
		}
	}

	values := []string{}
	for _, code := range codes {
		name := languageName(code)
		if name != "" {
			values = append(values, name)
		}
	}
	return values
}

func (bib BibResp) RegionFacet() []string {
	// a_fields_spec = options[:geo_a_fields] || "651a:691a"
	// z_fields_spec = options[:geo_z_fields] || "600:610:611:630:648:650:654:655:656:690:651:691"
	//
	// extractor_043a      = MarcExtractor.new("043a", :separator => nil)
	// extractor_a_fields  = MarcExtractor.new(a_fields_spec, :separator => nil)
	// extractor_z_fields  = MarcExtractor.new(z_fields_spec)

	values := []string{}
	for _, value := range bib.MarcValues("043a") {
		code := trimPunct(value)
		code = strings.TrimRight(code, "-")
		name := regionName(code)
		// if name == "" {
		// 	name = code
		// }
		if name != "" && !in(values, name) {
			values = append(values, name)
		}
	}

	aFieldSpec := "651a:691a"
	for _, value := range bib.MarcValues(aFieldSpec) {
		trimVal := trimPunct(value)
		if !in(values, trimVal) {
			values = append(values, trimVal)
		}
	}

	zFieldSpec := "600z:610z:611z:630z:648z:650z:654z:655z:656z:690z:651z:691z"
	for _, value := range bib.MarcValues(zFieldSpec) {
		trimVal := trimPunct(value)
		if !in(values, trimVal) {
			values = append(values, trimVal)
		}
	}

	return values
}

func (bib BibResp) AuthorFacet() []string {
	fieldSpec := "100abcd:110ab:111ab:700abcd:711ab"

	if f710, found := bib.getFields("710"); found {
		// If there is more than one 710 field this will only check the first one.
		// TODO: handle multi 710 fields
		if f710[0].Ind2 != "9" {
			fieldSpec += "710ab"
		}
	}

	values := bib.MarcValuesTrim(fieldSpec)
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
	// Logic stolen from
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21_semantics.rb
	// TODO do we need the field k logic here?
	titles := bib.MarcValues("245ab")
	if len(titles) == 0 {
		return ""
	}

	sortTitle := titles[0]
	fields, found := bib.getFields("245")
	if found {
		ind2 := toInt(fields[0].Ind2)
		if ind2 > 0 && len(sortTitle) > ind2 {
			// drop the prefix as notes in the second indicator
			sortTitle = sortTitle[ind2:len(sortTitle)]
		}
	}
	return trimPunct(sortTitle)
}

func (bib BibResp) CallNumbers() []string {
	values := []string{}
	callNumbers := bib.MarcValuesTrim("050ab:090ab:091ab:092ab:096ab:099ab")
	for _, number := range callNumbers {
		values = append(values, number)
	}
	return values
}

func (bib BibResp) Subjects() []string {
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
