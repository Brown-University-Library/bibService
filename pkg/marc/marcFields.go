package marc

import (
	"strings"
)

// MarcFields represents a MARC record as an array of MarcField.
// All the information of the MARC record is meant to be in MarcFields
// (including the Leader)
type MarcFields []MarcField

// ContentForFieldTag() returns the content for the Field Tag that matches
// the request tag. Notice that the FieldTag is different from the MARC tag.
// FieldTag seems to be a Sierra-specific thing. It is used to represent
// MARC records with minimal data (i.e. records with a minimal skeleton of
// data that is not in the MARC fields)
func (fields MarcFields) ContentForFieldTag(fieldTag string) string {
	for _, field := range fields {
		if field.FieldTag == fieldTag {
			return field.Content
		}
	}
	return ""
}

// ControlValue() returns the value as-is for a given MARC tag.
// This is meant to be used with MARC control fields (001-009)
// that have a single value. If more than one value is found they
// will be joined.
func (fields MarcFields) ControlValue(marcTag string) string {
	values := fields.ControlValues(marcTag)
	return strings.Join(values, " ")
}

// ControlValue() returns an array with the values as-is for a
// given MARC tag. This is meant to be used with MARC control
// fields (001-009).
func (fields MarcFields) ControlValues(marcTag string) []string {
	// TODO should I validate the marcTag is >= "001" && <= "009"
	values := []string{}
	for _, field := range fields.GetFields(marcTag) {
		values = append(values, field.Content)
	}
	return values
}

// MarcValues() returns an array of MarcField with the values for
// the fields and subfields indicated in `specsStr`. The result
// includes one row for each field where data was found.
//
// `specsStr` is something in the form "nnnabc" where "nnn" is the tag of the
// field and "abc" represents the subfields. For example: "100ac" means
// field "100" subfields "a" and "c". Multiple fields can be indicated
// separated by colons, for example: "100ac:210f".
func (fields MarcFields) FieldValues(specsStr string) MarcFields {
	values := []MarcField{}
	vernProcessed := []string{}
	specs := NewFieldSpecs(specsStr)
	for _, spec := range specs {

		fieldsFound := fields.GetFields(spec.MarcTag)
		if len(spec.Subfields) == 0 {
			// Get the value directly
			for _, field := range fieldsFound {
				if field.Content != "" {
					value := MarcField{MarcTag: spec.MarcTag, Content: field.Content}
					values = append(values, value)
				}
			}
			continue
		}

		// Process the subfields
		for _, field := range fieldsFound {
			fieldValues := field.Values(spec.Subfields)
			values = append(values, fieldValues)
		}

		// Gather the vernacular values for the fields
		for _, field := range fieldsFound {
			vernValues := fields.vernacularValuesFor(field, spec)
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
	f880s := fields.GetFields("880")
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

// GetFields() returns an array of MarcField for those that match the
// given MARC tag.
func (fields MarcFields) GetFields(marcTag string) MarcFields {
	fieldsFound := MarcFields{}
	for _, field := range fields {
		if field.MarcTag == marcTag {
			fieldsFound = append(fieldsFound, field)
		}
	}
	return fieldsFound
}

// HasMarc() returns true if there any of fields has a MARC tag
func (fields MarcFields) HasMarc() bool {
	for _, field := range fields {
		if field.MarcTag != "" {
			return true
		}
	}
	return false
}

// Leader() returns the content of the lader field tag.
// This is very Sierra-specific.
func (fields MarcFields) Leader() string {
	for _, field := range fields {
		if field.FieldTag == "_" {
			return field.Content
		}
	}
	return ""
}

// ToArray() returns an string array with the values in the fields.
// Values will be trimmed and subfields will be joined.
func (fields MarcFields) ToArray() []string {
	return fields.toArray(true, true)
}

// ToArray() returns an string array with the values in the fields.
// Values will be trimmed but subfields will not be joined.
func (fields MarcFields) ToArrayTrim() []string {
	return fields.toArray(true, false)
}

// ToArray() returns an string array with the values in the fields.
// Values will not be trimmed but subfields will be joined.
func (fields MarcFields) ToArrayJoin() []string {
	return fields.toArray(false, true)
}

// ToArray() returns an string array with the values in the fields.
// No trimming or joining is performed on the values.
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
				value = TrimPunct(value)
			}
			safeAppend(&array, value)
		} else {
			// preserve each individual value (regardless of their field)
			// as a single element in the returning arrray
			for _, value := range field.Strings() {
				if trim {
					value = TrimPunct(value)
				}
				safeAppend(&array, value)
			}
		}
	}
	return array
}

// VernacularValues() returns an array of MarcFields with the fields that
// are the vernacular representation for the given specs.
func (fields MarcFields) VernacularValues(specsStr string) MarcFields {
	// Notice that we loop through the 880 fields rather than checking if
	// each of the indicated fields have vernacular values because sometimes
	// the actual field does not point to the 880 but the 880 always points
	// to the original field.
	values := MarcFields{}
	f880s := fields.GetFields("880")
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

func (fields MarcFields) vernacularValuesFor(field MarcField, spec FieldSpec) MarcFields {
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
	for _, vernField := range fields.GetFields(marcTag) {
		// ...is this the one that corresponds with the tag 6
		// value that we calculated (e.g. 700-04)
		if vernField.IsVernacularFor(tag6) {
			vernValues := vernField.Values(spec.Subfields)
			values = append(values, vernValues)
		}
	}
	return values
}
