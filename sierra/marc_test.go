package sierra

import (
	"testing"
)

//
// BASIC TESTS
//
func TestMarcValuesJoin(t *testing.T) {
	a1 := map[string]string{"tag": "a", "content": "A1"}
	b1 := map[string]string{"tag": "b", "content": "B1"}
	f1 := MarcField{MarcTag: "520"}
	f1.Subfields = []map[string]string{a1, b1}

	a2 := map[string]string{"tag": "a", "content": "A2"}
	f2 := MarcField{MarcTag: "520"}
	f2.Subfields = []map[string]string{a2}

	a3 := map[string]string{"tag": "a", "content": "A3"}
	b3 := map[string]string{"tag": "b", "content": "B3"}
	f3 := MarcField{MarcTag: "520"}
	f3.Subfields = []map[string]string{a3, b3}

	fields := MarcFields{f1, f2, f3}

	joinedValues := fields.MarcValuesByField("520ab", true)
	if len(joinedValues) != 3 ||
		joinedValues[0][0] != "A1 B1" ||
		joinedValues[1][0] != "A2" ||
		joinedValues[2][0] != "A3 B3" {
		t.Errorf("Unexpected joinedValues: %#v", joinedValues)
	}

	values := fields.MarcValuesByField("520ab", false)
	if len(values) != 3 ||
		values[0][0] != "A1" || values[0][1] != "B1" ||
		values[1][0] != "A2" ||
		values[2][0] != "A3" || values[2][1] != "B3" {
		t.Errorf("Unexpected values: %#v", values)
	}
}

func TestTwoFields(t *testing.T) {
	ta1 := map[string]string{"tag": "a", "content": "a1"}
	field1 := MarcField{MarcTag: "100"}
	field1.Subfields = []map[string]string{ta1}

	ta2 := map[string]string{"tag": "a", "content": "a2"}
	field2 := MarcField{MarcTag: "100"}
	field2.Subfields = []map[string]string{ta2}

	fields := MarcFields{field1, field2}

	// Two fields should result in two results
	// (even if they are the same MARC field).
	values := fields.MarcValuesByField("100abc", true)
	if len(values) != 2 {
		t.Errorf("Unexpected number of results: %d", len(values))
	}

	if !in(values[0], "a1") || !in(values[1], "a2") {
		t.Errorf("Did not fetch the expected values: %#v", values)
	}
}

func TestFieldValuesJoined(t *testing.T) {
	ta1 := map[string]string{"tag": "a", "content": "X a"}
	tb1 := map[string]string{"tag": "b", "content": "X b"}
	tc1 := map[string]string{"tag": "c", "content": "X c"}
	field1 := MarcField{MarcTag: "100"}
	field1.Subfields = []map[string]string{ta1, tb1, tc1}

	ta2 := map[string]string{"tag": "a", "content": "Y a"}
	tb2 := map[string]string{"tag": "b", "content": "Y b"}
	field2 := MarcField{MarcTag: "100"}
	field2.Subfields = []map[string]string{ta2, tb2}

	fields := MarcFields{field1, field2}
	values := fields.MarcValuesByField("100abc", true)

	// Joins the values per-field when the spec has multiple
	// subfields (e.g. "abc")
	if !in(values[0], "X a X b X c") {
		t.Errorf("Did not fetch the expected values: %#v", values)
	}

	if !in(values[1], "Y a Y b") {
		t.Errorf("Did not fetch the expected values: %#v", values)
	}
}

func TestFieldValuesNotJoined(t *testing.T) {
	// Sample record b8060047
	t1 := map[string]string{"tag": "t", "content": "T1"}
	f1 := MarcField{MarcTag: "505"}
	f1.Subfields = []map[string]string{t1}

	t2 := map[string]string{"tag": "t", "content": "T2"}
	f2 := MarcField{MarcTag: "505"}
	f2.Subfields = []map[string]string{t2}

	t3 := map[string]string{"tag": "t", "content": "T3"}
	f3 := MarcField{MarcTag: "505"}
	f3.Subfields = []map[string]string{t3}
	fields := MarcFields{f1, f2, f3}

	// Does not join the values per-field when the spec has a single
	// subfield (e.g. "t")
	values := fields.MarcValuesByField("505t", true)
	if values[0][0] != "T1" || values[1][0] != "T2" || values[2][0] != "T3" {
		t.Errorf("Did not fetch the expected values: %#v", values)
	}
}

//
// VERNACULAR
//
func TestVernacular(t *testing.T) {
	// Two author values
	s1 := map[string]string{"tag": "6", "content": "880-04"}
	a1 := map[string]string{"tag": "a", "content": "aaa"}
	b1 := map[string]string{"tag": "b", "content": "bbb"}
	f700_1 := MarcField{MarcTag: "700"}
	f700_1.Subfields = []map[string]string{s1, a1, b1}

	s2 := map[string]string{"tag": "6", "content": "880-05"}
	a2 := map[string]string{"tag": "a", "content": "ccc"}
	f700_2 := MarcField{MarcTag: "700"}
	f700_2.Subfields = []map[string]string{s2, a2}

	// and their vernacular values
	s3 := map[string]string{"tag": "6", "content": "700-04/$1"}
	a3 := map[string]string{"tag": "a", "content": "AAA"}
	b3 := map[string]string{"tag": "b", "content": "BBB"}
	f880_1 := MarcField{MarcTag: "880"}
	f880_1.Subfields = []map[string]string{s3, a3, b3}

	s4 := map[string]string{"tag": "6", "content": "700-05/$1"}
	a4 := map[string]string{"tag": "a", "content": "CCC"}
	z4 := map[string]string{"tag": "z", "content": "ZZZ"} // should not be picked up
	f880_2 := MarcField{MarcTag: "880"}
	f880_2.Subfields = []map[string]string{s4, a4, z4}

	fields := MarcFields{f700_1, f700_2, f880_1, f880_2}

	// Make sure fetching the 700 picks up the associated 880 fields
	values := fields.MarcValues("700ab", true)
	if !in(values, "aaa bbb") || !in(values, "ccc") {
		t.Errorf("700 field values not found: %#v", values)
	}

	if !in(values, "AAA BBB") || !in(values, "CCC") {
		t.Errorf("880 field values not found: %#v", values)
	}

	if len(values) != 4 {
		t.Errorf("Unexpected number of values found: %#v", values)
	}
}

func TestVernacularFreestanding(t *testing.T) {
	t6 := map[string]string{"tag": "6", "content": "700-04/$1"}
	ta := map[string]string{"tag": "a", "content": "AAA"}
	tb := map[string]string{"tag": "b", "content": "BBB"}
	f880 := MarcField{MarcTag: "880"}
	f880.Subfields = []map[string]string{t6, ta, tb}
	fields := MarcFields{f880}

	vern := fields.VernacularValuesByField("700ab")
	if vern[0][0] != "AAA BBB" {
		t.Errorf("Did not pick up freestanding vernacular values")
	}

	// Make sure fetching the 700 picks up vernacular values
	// even though there is no 700 field in the record.
	values := fields.MarcValues("700ab", true)
	if !in(values, "AAA BBB") {
		t.Errorf("Did not pick up freestanding vernacular values")
	}
}

func TestVernacularIncompleteLinking(t *testing.T) {
	x1 := map[string]string{"tag": "6", "content": "880-01"}
	x2 := map[string]string{"tag": "a", "content": "XXX"}
	x3 := map[string]string{"tag": "d", "content": "xxx"}
	f700 := MarcField{MarcTag: "700"}
	f700.Subfields = []map[string]string{x1, x2, x3}

	a1 := map[string]string{"tag": "6", "content": "700-00/$1"}
	a2 := map[string]string{"tag": "a", "content": "AAA"}
	a3 := map[string]string{"tag": "d", "content": "aaa"}
	f880a := MarcField{MarcTag: "880"}
	f880a.Subfields = []map[string]string{a1, a2, a3}

	b1 := map[string]string{"tag": "6", "content": "700-00/$1"}
	b2 := map[string]string{"tag": "a", "content": "BBB"}
	b3 := map[string]string{"tag": "d", "content": "bbb"}
	f880b := MarcField{MarcTag: "880"}
	f880b.Subfields = []map[string]string{b1, b2, b3}

	c1 := map[string]string{"tag": "6", "content": "700-01/$1"}
	c2 := map[string]string{"tag": "a", "content": "CCC"}
	c3 := map[string]string{"tag": "d", "content": "ccc"}
	f880c := MarcField{MarcTag: "880"}
	f880c.Subfields = []map[string]string{c1, c2, c3}

	fields := MarcFields{f700, f880a, f880b, f880c}

	// Make sure the original value (700) is detected separated
	// from the vernacular (880s). Each of the 880s should come
	// as an independent field, hence len(all) == 4.
	all := fields.MarcValuesByField("700abcd", true)
	if len(all) != 4 {
		t.Errorf("Invalid values detected: %#v", all)
	}

	// Make sure all three vernacular values are picked up even if
	// their linking is incomplete (notice how one of them matches
	// "700-01" but not the other two only partially match "700")
	vern := fields.VernacularValuesByField("700abcd")
	if len(vern) != 3 {
		t.Errorf("Invalid vernacular values detected: %#v", vern)
	}
}

func TestTitleSeries(t *testing.T) {
	// real sample: https://search.library.brown.edu/catalog/b8060352

	// field 490
	f4906 := map[string]string{"tag": "6", "content": "880-04"}
	f490a := map[string]string{"tag": "a", "content": "Rekishi bunka raiburarī ;"}
	f490v := map[string]string{"tag": "v", "content": "451"}
	f490 := MarcField{MarcTag: "490"}
	f490.Subfields = []map[string]string{f4906, f490a, f490v}

	// vernacular for 490
	f8806 := map[string]string{"tag": "6", "content": "490-04/$1"}
	f880a := map[string]string{"tag": "a", "content": "歴史文化ライブラリー ;"}
	f880v := map[string]string{"tag": "v", "content": "451"}
	f880 := MarcField{MarcTag: "880"}
	f880.Subfields = []map[string]string{f8806, f880a, f880v}

	// field 830
	f8306 := map[string]string{"tag": "6", "content": "880-05/$1"}
	f830a := map[string]string{"tag": "a", "content": "Rekishi bunka raiburarī ;"}
	f830v := map[string]string{"tag": "v", "content": "451"}
	f830 := MarcField{MarcTag: "830"}
	f830.Subfields = []map[string]string{f8306, f830a, f830v}

	// vernacular for 830
	f8806x := map[string]string{"tag": "6", "content": "830-05/$1"}
	f880ax := map[string]string{"tag": "a", "content": "歴史文化ライブラリー ;"}
	f880vx := map[string]string{"tag": "v", "content": "451"}
	f880x := MarcField{MarcTag: "880"}
	f880x.Subfields = []map[string]string{f8806x, f880ax, f880vx}

	fields := MarcFields{f490, f830, f880, f880x}

	// Make sure it picks up the correct set of subfields for each field:
	//
	// 	values[0] => 490a
	// 	values[1] => 880a for 490a
	// 	values[2] => 830av
	// 	values[3] => 880av for 830av
	specsStr := "490a:830adv"
	values := fields.MarcValuesByField(specsStr, true)
	if values[0][0] != "Rekishi bunka raiburarī ;" ||
		values[1][0] != "歴史文化ライブラリー ;" ||
		values[2][0] != "Rekishi bunka raiburarī ; 451" ||
		values[3][0] != "歴史文化ライブラリー ; 451" {
		t.Errorf("Unexpected values were found: %#v")
	}
}

func TestPubYear008X(t *testing.T) {
	test1 := "760629c19749999ne tr pss o   0   a0eng  cas   "
	year, ok := pubYear008(test1, 15)
	if !ok || year != 1974 {
		t.Errorf("Failed on %s (%v, %v)", test1, ok, year)
	}

	test2 := "061108c200u9999nyuar ss 0 0eng ccas a "
	year, ok = pubYear008(test2, 15)
	if !ok || year != 2005 {
		t.Errorf("Failed on %s (%v, %v)", test2, ok, year)
	}

	// Eventually we want to handle this.
	test3 := "061108duuuu2002nyuar ss 0 0eng ccas a "
	year, ok = pubYear008(test3, 15)
	if ok {
		t.Errorf("Failed on %s (%v, %v)", test3, ok, year)
	}

	test4 := "061108q19501980nyuar ss 0 0eng ccas a "
	_, ok = pubYear008(test4, 15)
	if ok {
		t.Errorf("Should have returned false on questionable date %s (%v, %v)", test4)
	}
}
