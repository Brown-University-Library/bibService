package sierra

import (
	"testing"
)

func TestValuesWithVernacular(t *testing.T) {
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
	values := fields.MarcValues("700ab")
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

func TestSubfieldsDifferentTag(t *testing.T) {
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

	if !in(values[0], "X a X b X c") {
		t.Errorf("Did not fetch the expected values: %#v", values)
	}

	if !in(values[1], "Y a Y b") {
		t.Errorf("Did not fetch the expected values: %#v", values)
	}
}

func TestSubfieldsSameTag(t *testing.T) {
	// Sample record b8060047
	t1 := map[string]string{"tag": "t", "content": "T1"}
	t2 := map[string]string{"tag": "t", "content": "T2"}
	t3 := map[string]string{"tag": "t", "content": "T3"}
	tn := map[string]string{"tag": "n", "content": "N"}
	t4 := map[string]string{"tag": "t", "content": "T4"}
	field1 := MarcField{MarcTag: "550"}
	field1.Subfields = []map[string]string{t1, t2, t3, tn, t4}

	fields := MarcFields{field1}

	// Makes sure subfield values are combined for different subfields
	// (e.g. "T2 N") but kept separate for repeated the rest ("T1", "T2", "T4")
	values := fields.MarcValuesByField("550tnx", true)
	if !in(values[0], "T1") || !in(values[0], "T2") || !in(values[0], "T3 N") || !in(values[0], "T4") {
		t.Errorf("Did not fetch the expected values: %#v", values)
	}
}

func TestValuesFreestandingVernacular(t *testing.T) {
	t6 := map[string]string{"tag": "6", "content": "700-04/$1"}
	ta := map[string]string{"tag": "a", "content": "AAA"}
	tb := map[string]string{"tag": "b", "content": "BBB"}
	f880 := MarcField{MarcTag: "880"}
	f880.Subfields = []map[string]string{t6, ta, tb}
	fields := MarcFields{f880}

	// Make sure fetching the 700 picks up vernacular values
	// even though there is no 700 field in the record.
	values := fields.MarcValues("700ab")
	if !in(values, "AAA BBB") {
		t.Errorf("Did not pick up freestanding vernacular values")
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
}
