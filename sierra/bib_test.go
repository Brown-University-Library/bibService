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

func TestValuesWithVernacular(t *testing.T) {
	// Two author values
	s1 := map[string]string{"tag": "6", "content": "880-04"}
	a1 := map[string]string{"tag": "a", "content": "aaa"}
	b1 := map[string]string{"tag": "b", "content": "bbb"}
	f700_1 := Field{MarcTag: "700"}
	f700_1.Subfields = []map[string]string{s1, a1, b1}

	s2 := map[string]string{"tag": "6", "content": "880-05"}
	a2 := map[string]string{"tag": "a", "content": "ccc"}
	f700_2 := Field{MarcTag: "700"}
	f700_2.Subfields = []map[string]string{s2, a2}

	// and their vernacular values
	s3 := map[string]string{"tag": "6", "content": "700-04/$1"}
	a3 := map[string]string{"tag": "a", "content": "AAA"}
	b3 := map[string]string{"tag": "b", "content": "BBB"}
	f880_1 := Field{MarcTag: "880"}
	f880_1.Subfields = []map[string]string{s3, a3, b3}

	s4 := map[string]string{"tag": "6", "content": "700-05/$1"}
	a4 := map[string]string{"tag": "a", "content": "CCC"}
	z4 := map[string]string{"tag": "z", "content": "ZZZ"} // should not be picked up
	f880_2 := Field{MarcTag: "880"}
	f880_2.Subfields = []map[string]string{s4, a4, z4}

	// A document with all the values
	fields := []Field{f700_1, f700_2, f880_1, f880_2}
	bib := Bib{VarFields: fields}

	// Make sure fetching the 700 picks up the associated 880 fields
	values := bib.MarcValues("700ab")
	if !in(values, "aaa bbb") || !in(values, "ccc") {
		t.Errorf("700 field values not found")
	}

	if !in(values, "AAA BBB") || !in(values, "CCC") {
		t.Errorf("880 field values not found")
	}

	if len(values) != 4 {
		t.Errorf("Unexpected values found")
	}
}

func TestTwoFields(t *testing.T) {
	ta1 := map[string]string{"tag": "a", "content": "a1"}
	field1 := Field{MarcTag: "100"}
	field1.Subfields = []map[string]string{ta1}

	ta2 := map[string]string{"tag": "a", "content": "a2"}
	field2 := Field{MarcTag: "100"}
	field2.Subfields = []map[string]string{ta2}

	fields := []Field{field1, field2}
	bib := Bib{VarFields: fields}

	// Two fields should result in two results
	// (even if they are the same MARC field).
	values := bib.MarcValuesByField("100abc")
	if len(values) != 2 {
		t.Errorf("Unexpected number of results: %d", len(values))
	}

	if !in(values[0], "a1") || !in(values[1], "a2") {
		t.Errorf("Did not fetch the expected values: %#v", values)
	}
}

func TestSubfieldsDifferentTag(t *testing.T) {
	ta1 := map[string]string{"tag": "a", "content": "X a"}
	tb1 := map[string]string{"tag": "b", "content": "X b"}
	tc1 := map[string]string{"tag": "c", "content": "X c"}
	field1 := Field{MarcTag: "100"}
	field1.Subfields = []map[string]string{ta1, tb1, tc1}

	ta2 := map[string]string{"tag": "a", "content": "Y a"}
	tb2 := map[string]string{"tag": "b", "content": "Y b"}
	field2 := Field{MarcTag: "100"}
	field2.Subfields = []map[string]string{ta2, tb2}

	fields := []Field{field1, field2}
	bib := Bib{VarFields: fields}
	values := bib.MarcValuesByField("100abc")

	// Each tag is in its own array element for each field.
	if !in(values[0], "X a") || !in(values[0], "X b") || !in(values[0], "X c") {
		t.Errorf("Did not fetch the expected values: %#v", values)
	}

	if !in(values[1], "Y a") || !in(values[1], "Y b") {
		t.Errorf("Did not fetch the expected values: %#v", values)
	}
}

func TestSubfieldsSameTag(t *testing.T) {
	t1 := map[string]string{"tag": "t", "content": "T1"}
	t2 := map[string]string{"tag": "t", "content": "T2"}
	t3 := map[string]string{"tag": "t", "content": "T3"}
	tn := map[string]string{"tag": "n", "content": "N"}
	t4 := map[string]string{"tag": "t", "content": "T4"}
	field1 := Field{MarcTag: "550"}
	field1.Subfields = []map[string]string{t1, t2, t3, tn, t4}

	fields := []Field{field1}
	bib := Bib{VarFields: fields}

	// Makes sure subfield values are combined for different subfields
	// (e.g. "T2 N") but kept separate for repeated the rest ("T1", "T2", "T4")
	values := bib.MarcValuesByField("550tnx")
	if !in(values[0], "T1") || !in(values[0], "T2") || !in(values[0], "T3 N") || !in(values[0], "T4") {
		t.Errorf("Did not fetch the expected values: %#v", values)
	}
}

func TestValuesFreestandingVernacular(t *testing.T) {
	t6 := map[string]string{"tag": "6", "content": "700-04/$1"}
	ta := map[string]string{"tag": "a", "content": "AAA"}
	tb := map[string]string{"tag": "b", "content": "BBB"}
	f880 := Field{MarcTag: "880"}
	f880.Subfields = []map[string]string{t6, ta, tb}
	fields := []Field{f880}
	bib := Bib{VarFields: fields}

	// Make sure fetching the 700 picks up vernacular values
	// even though there is no 700 field in the record.
	values := bib.MarcValues("700ab")
	if !in(values, "AAA BBB") {
		t.Errorf("Did not pick up freestanding vernacular values")
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

func TestTitleVernacularDisplay(t *testing.T) {
	// real sample: https://search.library.brown.edu/catalog/b8060012
	// title in english
	f2456 := map[string]string{"content": "880-03", "tag": "6"}
	f245a := map[string]string{"content": "whatever", "tag": "a"}
	f245 := Field{MarcTag: "245"}
	f245.Subfields = []map[string]string{f245a, f2456}

	// title in language
	f8806 := map[string]string{"content": "245-03/$1", "tag": "6"}
	f880a := map[string]string{"content": "titulo en español:", "tag": "a"}
	f880b := map[string]string{"content": "bb", "tag": "b"}
	f880c := map[string]string{"content": "cc", "tag": "c"}
	f880 := Field{MarcTag: "880"}
	f880.Subfields = []map[string]string{f8806, f880a, f880b, f880c}

	fields := []Field{f245, f880}
	bib := Bib{VarFields: fields}
	title := bib.TitleVernacularDisplay()
	if title != "titulo en español: bb" {
		t.Errorf("Invalid vernacular title found: %s", title)
	}
}

func TestUniformTitleTwoValues(t *testing.T) {
	// real sample https://search.library.brown.edu/catalog/b8060083
	f130a := map[string]string{"content": "Neues Licht.", "tag": "a"}
	f130l := map[string]string{"content": "English.", "tag": "l"}
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

	// real sample https://search.library.brown.edu/catalog/b8060295
	f240a := map[string]string{"content": "Poems.", "tag": "a"}
	f240k := map[string]string{"content": "Selections.", "tag": "k"}
	f240 := Field{MarcTag: "240"}
	f240.Subfields = []map[string]string{f240a, f240k}
	fields = []Field{f240}
	bib = Bib{VarFields: fields}
	titles = bib.UniformTitles(true)
	if len(titles) != 1 {
		t.Errorf("Invalid number of titles found (field 240): %d, %v", len(titles), titles)
	} else {
		if titles[0].Title[0].Display != "Poems." {
			t.Errorf("Subfield a not found: %v", titles)
		}
		if titles[0].Title[1].Display != "Selections." {
			t.Errorf("Subfield k not found: %v", titles)
		}
	}
}

func TestUniformTitleVernacular(t *testing.T) {
	// real sample: https://search.library.brown.edu/catalog/b8060012
	// title in english
	f2406 := map[string]string{"content": "880-02", "tag": "6"}
	f240a := map[string]string{"content": "title in english.", "tag": "a"}
	f240 := Field{MarcTag: "240"}
	f240.Subfields = []map[string]string{f240a, f2406}

	// title in language
	f8806 := map[string]string{"content": "240-02/$1", "tag": "6"}
	f880a := map[string]string{"content": "titulo en español.", "tag": "a"}
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
	f2406 := map[string]string{"content": "880-02", "tag": "6"}
	f240a := map[string]string{"content": "title in english.", "tag": "a"}
	f240l := map[string]string{"content": "English.", "tag": "l"}
	f240 := Field{MarcTag: "240"}
	f240.Subfields = []map[string]string{f240a, f240l, f2406}

	// title in language
	f8806 := map[string]string{"content": "240-02", "tag": "6"}
	f880a := map[string]string{"content": "titulo en español.", "tag": "a"}
	f880l := map[string]string{"content": "Spanish.", "tag": "l"}
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
		if t2.Display != "English." || t2.Query != "title in english.. English." {
			t.Errorf("Invalid values in first title (2/2): %v", t2)
		}
	}

	if len(titles[1].Title) != 2 {
		t.Errorf("Invalid values in second title: %v", titles)
	} else {
		t1 := titles[1].Title[0]
		if t1.Display != "titulo en español." || t1.Query != "titulo en español." {
			t.Errorf("Invalid values in second title (1/2): %#v", t1)
			t.Errorf("%#v", t1.Display)
			t.Errorf("%#v", t1.Query)
		}
		t2 := titles[1].Title[1]
		if t2.Display != "Spanish." || t2.Query != "titulo en español.. Spanish." {
			t.Errorf("Invalid values in second title (2/2): %v", t2)
			t.Errorf("%#v", t2.Display)
			t.Errorf("%#v", t2.Query)
		}
	}
}
