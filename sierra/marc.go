package sierra

import (
	"math"
	"strings"
)

type MarcFields []MarcField

// TODO: should this return MarcFields?
func (allFields MarcFields) getFields(marcTag string) []MarcField {
	fields := []MarcField{}
	for _, field := range allFields {
		if field.MarcTag == marcTag {
			fields = append(fields, field)
		}
	}
	return fields
}

func (allFields MarcFields) vernacularValuesFor(field MarcField, spec FieldSpec, join bool) [][]string {
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

	// Process the fields indicated in target (e.g. 880s)...
	for _, vernField := range allFields.getFields(marcTag) {
		// ...is this the one that corresponds with the tag 6
		// value that we calculated (e.g. 700-04)
		if vernField.IsVernacularFor(tag6) {
			vernValues := vernField.Values(spec.Subfields, join)
			values = append(values, vernValues)
		}
	}
	return values
}

func (allFields MarcFields) VernacularValuesByField(specsStr string) [][]string {
	values := [][]string{}
	for _, spec := range NewFieldSpecs(specsStr) {
		for _, field := range allFields {
			if field.MarcTag == spec.MarcTag {
				for _, vernValues := range allFields.vernacularValuesFor(field, spec, false) {
					values = append(values, vernValues)
				}
			}
		}
	}
	return values
}

// `specsStr` is something in the form "nnna" where "nnnbc" is the tag of the
// field and "abc" represents the subfields. For example: "100ac" means
// field "100" subfields "a" and "c". Multiple fields can be indicated
// separated by colons, for example: "100ac:210f".
// When `join` is true the different values from the subfields are
// joined as a string for each field.
func (allFields MarcFields) MarcValuesByField(specsStr string, join bool) [][]string {
	values := [][]string{}
	vernProcessed := []string{}
	specs := NewFieldSpecs(specsStr)
	for _, spec := range specs {

		fields := allFields.getFields(spec.MarcTag)
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
			for _, vernValues := range allFields.vernacularValuesFor(field, spec, join) {
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
	f880s := allFields.getFields("880")
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

func (allFields MarcFields) MarcValue(specsStr string, join bool, trim bool) string {
	values := allFields.MarcValuesByField(specsStr, join)
	return valuesToString(values, trim)
}

func (allFields MarcFields) MarcValues(specStr string, join bool) []string {
	values := allFields.MarcValuesByField(specStr, join)
	return valuesToArray(values, join, false)
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

func pubYear008(f008 string, tolerance int) (int, bool) {
	// Logic stolen from
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21_semantics.rb
	//
	// e.g. "760629c19749999ne tr pss o   0   a0eng  cas   "
	if len(f008) < 11 {
		return 0, false
	}

	dateType := f008[6:7]
	if dateType == "n" {
		// unknown
		return 0, false
	}

	var dateStr1, dateStr2 string
	dateStr1 = f008[7:11]
	if len(f008) >= 15 {
		dateStr2 = f008[11:15]
	} else {
		dateStr2 = dateStr1
	}

	if dateType == "q" {
		// questionable
		date1 := toInt(strings.Replace(dateStr1, "u", "0", -1))
		date2 := toInt(strings.Replace(dateStr2, "u", "9", -1))
		if (date2 > date1) && ((date2 - date1) <= tolerance) {
			return (date2 + date1) / 2, true
		}
	}

	var dateStr string
	if dateType == "p" {
		// use the oldest date
		if dateStr1 <= dateStr2 || toInt(dateStr2) == 0 {
			dateStr = dateStr1
		} else {
			dateStr = dateStr2
		}
	} else if dateType == "r" && toInt(dateStr2) != 0 {
		dateStr = dateStr2 // use the second date
	} else {
		dateStr = dateStr1 // use the first date
	}

	uCount := strings.Count(dateStr, "u")
	// should we replace with "9" if we pick dateStr2 ?
	date := toInt(strings.Replace(dateStr, "u", "0", -1))
	if uCount > 0 && date != 0 {
		delta := int(math.Pow10(uCount))
		if delta <= tolerance {
			return date + (delta / 2), true
		}
	} else if date != 0 {
		return date, true
	}

	return 0, false
}
