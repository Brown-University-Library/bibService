package sierra

import (
	"testing"
)

func TestAuthors(t *testing.T) {
	a1 := map[string]string{"tag": "a", "content": "a1."}
	b1 := map[string]string{"tag": "b", "content": "b1."}
	b2 := map[string]string{"tag": "b", "content": "b2."}
	b3 := map[string]string{"tag": "b", "content": "b3,"}
	f110 := MarcField{MarcTag: "110"}
	f110.Subfields = []map[string]string{a1, b1, b2, b3}
	fields := MarcFields{f110}
	bib := Bib{VarFields: fields}
	authors := bib.AuthorsT()
	if authors[0] != "a1. b1. b2. b3" {
		t.Errorf("Unexpected values found: %#v", authors)
	}

	display := bib.AuthorDisplay()
	if display != "a1. b1. b2. b3" {
		t.Errorf("Unexpected values found: %#v", authors)
	}
}

func TestAuthorsAddl(t *testing.T) {
	a1 := map[string]string{"tag": "a", "content": "a1"}
	b1 := map[string]string{"tag": "b", "content": "b1"}
	f710_1 := MarcField{MarcTag: "710"}
	f710_1.Subfields = []map[string]string{a1, b1}

	a2 := map[string]string{"tag": "a", "content": "a2"}
	b2 := map[string]string{"tag": "b", "content": "b2"}
	b22 := map[string]string{"tag": "b", "content": "b22"}
	f710_2 := MarcField{MarcTag: "710"}
	f710_2.Subfields = []map[string]string{a2, b2, b22}

	a3 := map[string]string{"tag": "a", "content": "a3"}
	f710_3 := MarcField{MarcTag: "710"}
	f710_3.Subfields = []map[string]string{a3}

	fields := MarcFields{f710_1, f710_2, f710_3}
	bib := Bib{VarFields: fields}

	// Check authors additional display logic
	addlDisplay := bib.AuthorsAddlDisplay()
	if len(addlDisplay) != 3 {
		t.Errorf("Unexpected number of values found: %d. Values: %#v", len(addlDisplay), addlDisplay)
	}
	if addlDisplay[0] != "a1 b1" || addlDisplay[1] != "a2 b2 b22" || addlDisplay[2] != "a3" {
		t.Errorf("Unexpected values found: %#v", addlDisplay)
	}

	// Check authors additional logic
	authors := bib.AuthorsAddlT()
	if len(authors) != 3 {
		t.Errorf("Unexpected number of values found: %d. Values: %#v", len(authors), authors)
	}
	if authors[0] != "a1 b1" || authors[1] != "a2 b2 b22" || authors[2] != "a3" {
		t.Errorf("Unexpected values found: %#v", authors)
	}
}

func TestLanguage(t *testing.T) {
	lang1 := map[string]string{"content": "eng", "tag": "a"}
	lang2 := map[string]string{"content": "fre", "tag": "a"}
	lang3 := map[string]string{"content": "spa", "tag": "a"}

	field := MarcField{MarcTag: "041"}
	field.Subfields = []map[string]string{lang1, lang2, lang3}
	fields := MarcFields{field}
	bib := Bib{VarFields: fields}
	values := bib.Languages()
	if !in(values, "English") || !in(values, "Spanish") || !in(values, "French") {
		t.Errorf("Expected languages not found: %#v", values)
	}
}

func TestOclcNum(t *testing.T) {

	f001 := MarcField{MarcTag: "001", Content: "ocn987070476"}

	a1 := map[string]string{"tag": "a", "content": "ocn987070400"}
	f035_1 := MarcField{MarcTag: "035"}
	f035_1.Subfields = []map[string]string{a1}

	a2 := map[string]string{"tag": "a", "content": "ocm987070400"}
	f035_2 := MarcField{MarcTag: "035"}
	f035_2.Subfields = []map[string]string{a2}

	z := map[string]string{"tag": "z", "content": "ocn987070499"}
	f035_3 := MarcField{MarcTag: "035"}
	f035_3.Subfields = []map[string]string{z}

	fields := MarcFields{f001, f035_1, f035_2, f035_3}
	bib := Bib{VarFields: fields}

	nums := bib.OclcNum()
	if len(nums) != 3 || nums[0] != "987070476" || nums[1] != "987070400" || nums[2] != "987070499" {
		t.Errorf("Unexpected values: %#v", nums)
	}
}

func TestRegionFacetWithParent(t *testing.T) {
	z1 := map[string]string{"content": "usa", "tag": "z"}
	z2 := map[string]string{"content": "ri", "tag": "z"}
	field := MarcField{MarcTag: "650"}
	field.Subfields = []map[string]string{z1, z2}
	fields := MarcFields{field}
	bib := Bib{VarFields: fields}
	facets := bib.RegionFacet()
	if !in(facets, "usa") || !in(facets, "ri (usa)") {
		t.Errorf("Failed to detect parent region: %#v", facets)
	}
}

func TestTitleT(t *testing.T) {
	a1 := map[string]string{"tag": "a", "content": "a1."}
	p1 := map[string]string{"tag": "p", "content": "p1."}
	p2 := map[string]string{"tag": "p", "content": "p2"}
	field := MarcField{MarcTag: "130"}
	field.Subfields = []map[string]string{a1, p1, p2}
	fields := MarcFields{field}
	bib := Bib{VarFields: fields}
	titles := bib.TitleT()
	// when multiple subfields are present, value is joined
	if titles[0] != "a1. p1. p2" {
		t.Errorf("Unexpected titles found: %#v", titles)
	}

	t1 := map[string]string{"tag": "t", "content": "t1 /"}
	t2 := map[string]string{"tag": "t", "content": "t2 /"}
	t3 := map[string]string{"tag": "t", "content": "t3 /"}
	field = MarcField{MarcTag: "505"}
	field.Subfields = []map[string]string{t1, t2, t3}
	fields = MarcFields{field}
	bib = Bib{VarFields: fields}
	titles = bib.TitleT()
	// when a single subfield is repeated, values are NOT joines
	if titles[0] != "t1" || titles[1] != "t2" || titles[2] != "t3" {
		t.Errorf("Unexpected titles found: %#v", titles)
	}
}

func TestRegionFacet(t *testing.T) {
	z1 := map[string]string{"content": "usa", "tag": "z"}
	z2 := map[string]string{"content": "ri", "tag": "z"}
	z3 := map[string]string{"content": "zz", "tag": "z"}
	field := MarcField{MarcTag: "650"}
	field.Subfields = []map[string]string{z1, z2, z3}
	fields := MarcFields{field}
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
	f245 := MarcField{MarcTag: "245"}
	f245.Subfields = []map[string]string{f245a, f2456}

	// title in language
	f8806 := map[string]string{"content": "245-03/$1", "tag": "6"}
	f880a := map[string]string{"content": "titulo en español:", "tag": "a"}
	f880b := map[string]string{"content": "bb", "tag": "b"}
	f880c := map[string]string{"content": "cc", "tag": "c"}
	f880 := MarcField{MarcTag: "880"}
	f880.Subfields = []map[string]string{f8806, f880a, f880b, f880c}

	fields := MarcFields{f245, f880}
	bib := Bib{VarFields: fields}
	title := bib.TitleVernacularDisplay()
	if title != "titulo en español: bb" {
		t.Errorf("Invalid vernacular title found: %s", title)
	}
}

func TestUniformTitleTwoValues(t *testing.T) {
	// real sample https://search.library.brown.edu/catalog/b8060083
	f130a := map[string]string{"tag": "a", "content": "Neues Licht."}
	f130l := map[string]string{"tag": "l", "content": "English."}
	f130 := MarcField{MarcTag: "130"}
	f130.Subfields = []map[string]string{f130a, f130l}
	fields := MarcFields{f130}
	bib := Bib{VarFields: fields}
	titles := bib.UniformTitles(false)
	if len(titles) != 1 {
		t.Errorf("Invalid number of titles found (field 130): %d, %v", len(titles), titles)
	} else {
		if titles[0].Title[0].Display != "Neues Licht." {
			t.Errorf("Subfield a not found: %#v", titles)
		}

		if titles[0].Title[1].Display != "English." {
			t.Errorf("Subfield l not found: %#v", titles)
		}
	}

	// real sample https://search.library.brown.edu/catalog/b8060295
	f240a := map[string]string{"content": "Poems.", "tag": "a"}
	f240k := map[string]string{"content": "Selections.", "tag": "k"}
	f240 := MarcField{MarcTag: "240"}
	f240.Subfields = []map[string]string{f240a, f240k}
	fields = MarcFields{f240}
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
	f240 := MarcField{MarcTag: "240"}
	f240.Subfields = []map[string]string{f240a, f2406}

	// title in language
	f8806 := map[string]string{"content": "240-02/$1", "tag": "6"}
	f880a := map[string]string{"content": "titulo en español.", "tag": "a"}
	f880 := MarcField{MarcTag: "880"}
	f880.Subfields = []map[string]string{f8806, f880a}

	fields := MarcFields{f240, f880}
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

func TestPublishedDisplay(t *testing.T) {
	// Sample record b8060074
	s1 := map[string]string{"tag": "a", "content": "new haven"}
	s2 := map[string]string{"tag": "b", "content": "yale"}
	s3 := map[string]string{"tag": "a", "content": "london"}
	s4 := map[string]string{"tag": "b", "content": "humphrey"}
	s5 := map[string]string{"tag": "b", "content": "oxford"}
	s6 := map[string]string{"tag": "c", "content": "1942"}
	field := MarcField{MarcTag: "260"}
	field.Subfields = []map[string]string{s1, s2, s3, s4, s5, s6}
	fields := MarcFields{field}
	bib := Bib{VarFields: fields}
	values := bib.PublishedDisplay()
	if !in(values, "new haven") || !in(values, "london") {
		t.Errorf("Incorrect values found: %#v", values)
	}
}

func TestUniformTitleVernacularMany(t *testing.T) {
	// real sample: https://search.library.brown.edu/catalog/b8060012
	// title in english
	f2406 := map[string]string{"content": "880-02", "tag": "6"}
	f240a := map[string]string{"content": "title in english.", "tag": "a"}
	f240l := map[string]string{"content": "English.", "tag": "l"}
	f240 := MarcField{MarcTag: "240"}
	f240.Subfields = []map[string]string{f240a, f240l, f2406}

	// title in language
	f8806 := map[string]string{"content": "240-02", "tag": "6"}
	f880a := map[string]string{"content": "titulo en español.", "tag": "a"}
	f880l := map[string]string{"content": "Spanish.", "tag": "l"}
	f880 := MarcField{MarcTag: "880"}
	f880.Subfields = []map[string]string{f8806, f880a, f880l}

	fields := MarcFields{f240, f880}
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
			t.Errorf("%#v", t1.Display)
			t.Errorf("%#v", t1.Query)
		}
		t2 := titles[1].Title[1]
		if t2.Display != "Spanish." || t2.Query != "titulo en español. Spanish." {
			t.Errorf("Invalid values in second title (2/2): %v", t2)
			t.Errorf("%#v", t2.Display)
			t.Errorf("%#v", t2.Query)
		}
	}
}

func TestFormatCode(t *testing.T) {
	// Video format depends on MARC 007
	leader := MarcField{FieldTag: "_", Content: "00000ngm 2200000Ia 4500"}
	f007 := MarcField{MarcTag: "007", Content: "vf mbahos"}
	fields := MarcFields{leader, f007}
	bib := Bib{VarFields: fields}
	code := bib.FormatCode()
	if code != "BV" {
		t.Errorf("Failed to detect a video: %#v", code)
	}
}
