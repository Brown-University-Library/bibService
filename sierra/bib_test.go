package sierra

import (
	"regexp"
	"testing"
)

func TestFieldSpec(t *testing.T) {
	// Simple spec: MARC field and no subfields
	specs := NewFieldSpecs("100")
	if len(specs) != 1 {
		t.Errorf("Invalid number of specs detected")
	}

	if specs[0].MarcTag != "100" || len(specs[0].Subfields) > 0 {
		t.Errorf("Invalid spec detected: %v", specs[0])
	}

	// Spec with a MARC field and subfields
	specs = NewFieldSpecs("200ac")
	if len(specs) != 1 {
		t.Errorf("Invalid number of specs detected")
	}

	if specs[0].MarcTag != "200" || len(specs[0].Subfields) != 2 ||
		specs[0].Subfields[0] != "a" || specs[0].Subfields[1] != "c" {
		t.Errorf("Invalid spec detected: %v", specs[0])
	}

	// Multi-field spec
	specs = NewFieldSpecs("300ac:456:567x")
	if len(specs) != 3 {
		t.Errorf("Invalid number of specs detected")
	}
}

func TestTrimPunct(t *testing.T) {
	if trimPunct("one hundred/ ") != "one hundred" {
		t.Errorf("Failed to remove trailing slash")
	}

	if trimPunct("one.") != "one" {
		t.Errorf("Failed to remove trailing period")
	}

	if trimPunct("ct.") != "ct." {
		t.Errorf("Removed trailing period")
	}

	if trimPunct("[hello") != "hello" {
		t.Errorf("Failed to remove square bracket")
	}

	if trimPunct("[hello]") != "hello" {
		t.Errorf("Failed to remove square brackets")
	}

	if trimPunct("[hello [world]]") != "[hello [world]]" {
		t.Errorf("Removed square brackets")
	}
}

func TestLanguage(t *testing.T) {
	lang1 := map[string]string{"content": "eng", "tag": "a"}
	lang2 := map[string]string{"content": "fre", "tag": "a"}
	lang3 := map[string]string{"content": "spa", "tag": "a"}

	field := VarFieldResp{MarcTag: "041"}
	field.Subfields = []map[string]string{lang1, lang2, lang3}
	fieldData := []VarFieldResp{field}
	bib := BibResp{VarFields: fieldData}
	values := bib.Languages()
	if !in(values, "English") || !in(values, "Spanish") || !in(values, "French") {
		t.Errorf("Expected languages not found: %#v", values)
	}
}

func TestGetSubfieldValues(t *testing.T) {
	lang1 := map[string]string{"content": "eng", "tag": "a"}
	lang2 := map[string]string{"content": "fre", "tag": "a"}
	lang3 := map[string]string{"content": "spa", "tag": "a"}

	field := VarFieldResp{MarcTag: "041"}
	field.Subfields = []map[string]string{lang1, lang2, lang3}

	subfields := []string{"a"}
	values := field.getSubfieldsValues(subfields)
	if len(values) != 3 {
		t.Errorf("Incorrect number of values found: %#v", values)
	}

	if !in(values, "eng") || !in(values, "fre") || !in(values, "spa") {
		t.Errorf("Expected value not found: %#v", values)
	}
}

func TestOclcNum(t *testing.T) {
	re := regexp.MustCompile("\\s*(ocn|\\(OCoLC\\))(\\d+)")

	test1 := "ocn987070476"
	value1 := re.ReplaceAllString(test1, "$2")
	if value1 != "987070476" {
		t.Errorf("Failed to detect ocn prefix: %s", value1)
	}

	test2 := " (OCoLC)987070476"
	value2 := re.ReplaceAllString(test2, "$2")
	if value2 != "987070476" {
		t.Errorf("Failed to detect (OCoLC) prefix: %s", value2)
	}
}

func TestRegionFacetWithParent(t *testing.T) {
	z1 := map[string]string{"content": "usa", "tag": "z"}
	z2 := map[string]string{"content": "ri", "tag": "z"}

	field := VarFieldResp{MarcTag: "650"}
	field.Subfields = []map[string]string{z1, z2}
	fieldData := []VarFieldResp{field}
	bib := BibResp{VarFields: fieldData}
	facets := bib.RegionFacet()
	if !in(facets, "usa") || !in(facets, "ri (usa)") {
		t.Errorf("Failed to detect parent region: %#v", facets)
	}
}

func TestRegionFacet(t *testing.T) {
	z1 := map[string]string{"content": "usa", "tag": "z"}
	z2 := map[string]string{"content": "ri", "tag": "z"}
	z3 := map[string]string{"content": "zz", "tag": "z"}
	field := VarFieldResp{MarcTag: "650"}
	field.Subfields = []map[string]string{z1, z2, z3}
	fieldData := []VarFieldResp{field}
	bib := BibResp{VarFields: fieldData}
	facets := bib.RegionFacet()
	if !in(facets, "usa") || !in(facets, "ri") || !in(facets, "zz") {
		t.Errorf("Incorrectly handled regions: %#v", facets)
	}
}
