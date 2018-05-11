package sierra

import (
	"math"
	"strings"
)

type MarcFields []MarcField

// MarcValues returns an array of MarcField with the values for
// the fields and subfields indicated in `specsStr`. The result
// includes one row for each field where data was found.
//
// `specsStr` is something in the form "nnnabc" where "nnn" is the tag of the
// field and "abc" represents the subfields. For example: "100ac" means
// field "100" subfields "a" and "c". Multiple fields can be indicated
// separated by colons, for example: "100ac:210f".
func (allFields MarcFields) FieldValues(specsStr string) MarcFields {
	values := []MarcField{}
	vernProcessed := []string{}
	specs := NewFieldSpecs(specsStr)
	for _, spec := range specs {

		fields := allFields.getFields(spec.MarcTag)
		if len(spec.Subfields) == 0 {
			// Get the value directly
			for _, field := range fields {
				if field.Content != "" {
					value := MarcField{MarcTag: spec.MarcTag, Content: field.Content}
					values = append(values, value)
				}
			}
			continue
		}

		// Process the subfields
		for _, field := range fields {
			fieldValues := field.Values(spec.Subfields)
			values = append(values, fieldValues)
		}

		// Gather the vernacular values for the fields
		for _, field := range fields {
			vernValues := allFields.vernacularValuesFor(field, spec)
			if len(vernValues) > 0 && !in(vernProcessed, field.MarcTag) {
				vernProcessed = append(vernProcessed, field.MarcTag)
				for _, fieldVernValues := range vernValues {
					values = append(values, fieldVernValues)
				}
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
				vernValues := f880.Values(spec.Subfields)
				values = append(values, vernValues)
			}
		}
	}

	return values
}

// Returns the value as-is. No trimming of spaces or punctuation.
// This is very important for control fields.
func (allFields MarcFields) ControlValue(marcTag string) string {
	values := allFields.ControlValues(marcTag)
	return strings.Join(values, " ")
}

// Returns the values as-is. No trimming of spaces or punctuation.
// This is very important for control fields.
func (allFields MarcFields) ControlValues(marcTag string) []string {
	// TODO should I validate the marcTag is >= "001" && <= "009"
	values := []string{}
	for _, field := range allFields.getFields(marcTag) {
		values = append(values, field.Content)
	}
	return values
}

func (fields MarcFields) ToArray() []string {
	return fields.toArray(true, true)
}

func (fields MarcFields) ToArrayTrim() []string {
	return fields.toArray(true, false)
}

func (fields MarcFields) ToArrayJoin() []string {
	return fields.toArray(false, true)
}

func (fields MarcFields) ToArrayRaw() []string {
	return fields.toArray(false, false)
}

func (fields MarcFields) toArray(trim, join bool) []string {
	array := []string{}
	for _, field := range fields {
		if join {
			// join all the values for the field as a single element in
			// the returning array
			value := field.String()
			if trim {
				value = trimPunct(value)
			}
			safeAppend(&array, value)
		} else {
			// preserve each individual value (regardless of their field)
			// as a single element in the returning arrray
			for _, value := range field.Strings() {
				if trim {
					value = trimPunct(value)
				}
				safeAppend(&array, value)
			}
		}
	}
	return array
}

func (allFields MarcFields) VernacularValues(specsStr string) MarcFields {
	// Notice that we loop through the 880 fields rather than checking if
	// each of the indicated fields have vernacular values because sometimes
	// the actual field does not point to the 880 but the 880 always points
	// to the original field.
	values := MarcFields{}
	f880s := allFields.getFields("880")
	for _, spec := range NewFieldSpecs(specsStr) {
		for _, f880 := range f880s {
			if f880.IsVernacularFor(spec.MarcTag) {
				vernValues := f880.Values(spec.Subfields)
				// Should we test for an "empty" value?
				// (i.e. none of the subfields has a value)
				values = append(values, vernValues)
			}
		}
	}
	return values
}

func (allFields MarcFields) Leader() string {
	for _, field := range allFields {
		if field.FieldTag == "_" {
			return field.Content
		}
	}
	return ""
}

func (allFields MarcFields) hasMarc() bool {
	for _, field := range allFields {
		if field.MarcTag != "" {
			return true
		}
	}
	return false
}

func (allFields MarcFields) getFieldTagContent(fieldTag string) string {
	for _, field := range allFields {
		if field.FieldTag == fieldTag {
			return field.Content
		}
	}
	return ""
}

func (allFields MarcFields) getFields(marcTag string) MarcFields {
	fields := MarcFields{}
	for _, field := range allFields {
		if field.MarcTag == marcTag {
			fields = append(fields, field)
		}
	}
	return fields
}

func (allFields MarcFields) vernacularValuesFor(field MarcField, spec FieldSpec) MarcFields {
	values := MarcFields{}

	// True if the field (say "700") has subfield with tag 6.
	// Target would be "880-04"
	vern, target := field.HasVernacular()
	if !vern {
		return values
	}

	tokens := strings.Split(target, "-") // ["880", "04"]
	if len(tokens) < 2 {
		// bail out, we've got a value that we cannot parse
		return values
	}
	marcTag := tokens[0]  // "880"
	tag6 := field.MarcTag // "700" (we ignore the "-04" since it's not always used in the referenced "880".)

	// Process the fields indicated in target (e.g. 880s)...
	for _, vernField := range allFields.getFields(marcTag) {
		// ...is this the one that corresponds with the tag 6
		// value that we calculated (e.g. 700-04)
		if vernField.IsVernacularFor(tag6) {
			vernValues := vernField.Values(spec.Subfields)
			values = append(values, vernValues)
		}
	}
	return values
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
		} else {
			return 0, false
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
