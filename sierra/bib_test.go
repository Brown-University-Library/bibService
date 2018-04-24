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

	field := Field{MarcTag: "041"}
	field.Subfields = []map[string]string{lang1, lang2, lang3}
	fields := []Field{field}
	bib := Bib{VarFields: fields}
	values := bib.Languages()
	if !in(values, "English") || !in(values, "Spanish") || !in(values, "French") {
		t.Errorf("Expected languages not found: %#v", values)
	}
}

func TestGetSubfieldValues(t *testing.T) {
	lang1 := map[string]string{"content": "eng", "tag": "a"}
	lang2 := map[string]string{"content": "fre", "tag": "a"}
	lang3 := map[string]string{"content": "spa", "tag": "a"}

	field := Field{MarcTag: "041"}
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

	field := Field{MarcTag: "650"}
	field.Subfields = []map[string]string{z1, z2}
	fields := []Field{field}
	bib := Bib{VarFields: fields}
	facets := bib.RegionFacet()
	if !in(facets, "usa") || !in(facets, "ri (usa)") {
		t.Errorf("Failed to detect parent region: %#v", facets)
	}
}

func TestRegionFacet(t *testing.T) {
	z1 := map[string]string{"content": "usa", "tag": "z"}
	z2 := map[string]string{"content": "ri", "tag": "z"}
	z3 := map[string]string{"content": "zz", "tag": "z"}
	field := Field{MarcTag: "650"}
	field.Subfields = []map[string]string{z1, z2, z3}
	fields := []Field{field}
	bib := Bib{VarFields: fields}
	facets := bib.RegionFacet()
	if !in(facets, "usa") || !in(facets, "ri") || !in(facets, "zz") {
		t.Errorf("Incorrectly handled regions: %#v", facets)
	}
}

func TestUniformTitleTwoValues(t *testing.T) {
	// real sample https://search.library.brown.edu/catalog/b8060083
	f130a := map[string]string{"content": "Neues Licht.", "tag": "a"}
	f130l := map[string]string{"content": "English", "tag": "l"}
	f130 := Field{MarcTag: "130"}
	f130.Subfields = []map[string]string{f130a, f130l}
	fields := []Field{f130}
	bib := Bib{VarFields: fields}
	titles := bib.UniformTitles(false)
	if len(titles) != 1 {
		t.Errorf("Invalid number of titles found (field 130): %d, %v", len(titles), titles)
	} else {
		if titles[0].Title[0].Display != "Neues Licht." {
			t.Errorf("Subfield a not found: %v", titles)
		}

		if titles[0].Title[1].Display != "English." {
			t.Errorf("Subfield l not found: %v", titles)
		}
	}
}

func TestUniformTitleVernacular(t *testing.T) {
	// real sample: https://search.library.brown.edu/catalog/b8060012
	// title in english
	f240a := map[string]string{"content": "title in english", "tag": "a"}
	f240 := Field{MarcTag: "240"}
	f240.Subfields = []map[string]string{f240a}

	// title in language
	f8806 := map[string]string{"content": "240-02/$1", "tag": "6"}
	f880a := map[string]string{"content": "titulo en español", "tag": "a"}
	f880 := Field{MarcTag: "880"}
	f880.Subfields = []map[string]string{f8806, f880a}

	fields := []Field{f240, f880}
	bib := Bib{VarFields: fields}
	titles := bib.UniformTitles(true)
	if len(titles) != 2 {
		t.Errorf("Invalid number of titles found (field 240): %d, %v", len(titles), titles)
	}

	if titles[0].Title[0].Display != "title in english." {
		t.Errorf("Title in english not found: %v", titles)
	}

	if titles[1].Title[0].Display != "titulo en español." {
		t.Errorf("Title in spanish not found: %v", titles)
	}
}

func TestUniformTitleVernacularMany(t *testing.T) {
	// real sample: https://search.library.brown.edu/catalog/b8060012
	// title in english
	f240a := map[string]string{"content": "title in english", "tag": "a"}
	f240l := map[string]string{"content": "English", "tag": "l"}
	f240 := Field{MarcTag: "240"}
	f240.Subfields = []map[string]string{f240a, f240l}

	// title in language
	f8806 := map[string]string{"content": "240-00", "tag": "6"}
	f880a := map[string]string{"content": "titulo en español", "tag": "a"}
	f880l := map[string]string{"content": "Spanish", "tag": "l"}
	f880 := Field{MarcTag: "880"}
	f880.Subfields = []map[string]string{f8806, f880a, f880l}

	fields := []Field{f240, f880}
	bib := Bib{VarFields: fields}
	titles := bib.UniformTitles(true)
	if len(titles) != 2 {
		t.Errorf("Invalid number of titles found (field 240): %d, %v", len(titles), titles)
	}

	if len(titles[0].Title) != 2 {
		t.Errorf("Invalid number of values in first title: %v", titles)
	} else {
		t1 := titles[0].Title[0]
		if t1.Display != "title in english." || t1.Query != "title in english." {
			t.Errorf("Invalid values in first title (1/2): %#v", t1)
		}
		t2 := titles[0].Title[1]
		// TODO add "." 																			here->|
		if t2.Display != "English." || t2.Query != "title in english. English." {
			t.Errorf("Invalid values in first title (2/2): %v", t2)
		}
	}

	if len(titles[1].Title) != 2 {
		t.Errorf("Invalid values in second title: %v", titles)
	} else {
		t1 := titles[1].Title[0]
		if t1.Display != "titulo en español." || t1.Query != "titulo en español." {
			t.Errorf("Invalid values in second title (1/2): %#v", t1)
		}
		t2 := titles[1].Title[1]
		if t2.Display != "Spanish." || t2.Query != "titulo en español. Spanish." {
			t.Errorf("Invalid values in second title (2/2): %v", t2)
		}
	}
}
