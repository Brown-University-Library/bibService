package sierra

import (
	"bibService/pkg/marc"
	"testing"
)

func TestAuthors(t *testing.T) {
	a1 := map[string]string{"tag": "a", "content": "a1."}
	b1 := map[string]string{"tag": "b", "content": "b1."}
	b2 := map[string]string{"tag": "b", "content": "b2."}
	b3 := map[string]string{"tag": "b", "content": "b3,"}
	f110 := marc.MarcField{MarcTag: "110"}
	f110.Subfields = []map[string]string{a1, b1, b2, b3}
	fields := marc.MarcFields{f110}
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
	f710_1 := marc.MarcField{MarcTag: "710"}
	f710_1.Subfields = []map[string]string{a1, b1}

	a2 := map[string]string{"tag": "a", "content": "a2"}
	b2 := map[string]string{"tag": "b", "content": "b2"}
	b22 := map[string]string{"tag": "b", "content": "b22"}
	f710_2 := marc.MarcField{MarcTag: "710"}
	f710_2.Subfields = []map[string]string{a2, b2, b22}

	a3 := map[string]string{"tag": "a", "content": "a3"}
	f710_3 := marc.MarcField{MarcTag: "710"}
	f710_3.Subfields = []map[string]string{a3}

	fields := marc.MarcFields{f710_1, f710_2, f710_3}
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

	field := marc.MarcField{MarcTag: "041"}
	field.Subfields = []map[string]string{lang1, lang2, lang3}
	fields := marc.MarcFields{field}
	bib := Bib{VarFields: fields}
	values := bib.Languages()
	if !in(values, "English") || !in(values, "Spanish") || !in(values, "French") {
		t.Errorf("Expected languages not found: %#v", values)
	}
}

func TestOclcNum(t *testing.T) {

	f001 := marc.MarcField{MarcTag: "001", Content: "ocn987070476"}

	a1 := map[string]string{"tag": "a", "content": "ocn987070400"}
	f035_1 := marc.MarcField{MarcTag: "035"}
	f035_1.Subfields = []map[string]string{a1}

	a2 := map[string]string{"tag": "a", "content": "ocm987070400"}
	f035_2 := marc.MarcField{MarcTag: "035"}
	f035_2.Subfields = []map[string]string{a2}

	z := map[string]string{"tag": "z", "content": "ocn987070499"}
	f035_3 := marc.MarcField{MarcTag: "035"}
	f035_3.Subfields = []map[string]string{z}

	fields := marc.MarcFields{f001, f035_1, f035_2, f035_3}
	bib := Bib{VarFields: fields}

	nums := bib.OclcNum()
	if len(nums) != 3 || nums[0] != "987070476" || nums[1] != "987070400" || nums[2] != "987070499" {
		t.Errorf("Unexpected values: %#v", nums)
	}
}

func TestRegionFacetWithParent(t *testing.T) {
	z1 := map[string]string{"content": "usa", "tag": "z"}
	z2 := map[string]string{"content": "ri", "tag": "z"}
	field := marc.MarcField{MarcTag: "650"}
	field.Subfields = []map[string]string{z1, z2}
	fields := marc.MarcFields{field}
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
	field := marc.MarcField{MarcTag: "130"}
	field.Subfields = []map[string]string{a1, p1, p2}
	fields := marc.MarcFields{field}
	bib := Bib{VarFields: fields}
	titles := bib.TitleT()
	// when multiple subfields are present, value is joined
	if titles[0] != "a1. p1. p2" {
		t.Errorf("Unexpected titles found: %#v", titles)
	}

	t1 := map[string]string{"tag": "t", "content": "t1 /"}
	t2 := map[string]string{"tag": "t", "content": "t2 /"}
	t3 := map[string]string{"tag": "t", "content": "t3 /"}
	field = marc.MarcField{MarcTag: "505"}
	field.Subfields = []map[string]string{t1, t2, t3}
	fields = marc.MarcFields{field}
	bib = Bib{VarFields: fields}
	titles = bib.TitleT()
	// when a single subfield is repeated, values are NOT joines
	if titles[0] != "t1" || titles[1] != "t2" || titles[2] != "t3" {
		t.Errorf("Unexpected titles found: %#v", titles)
	}
}

func TestTitleSeries(t *testing.T) {
	f1 := map[string]string{"tag": "f", "content": "F1"}
	l1 := map[string]string{"tag": "l", "content": "P1"}
	f400 := marc.MarcField{MarcTag: "400"}
	f400.Subfields = []map[string]string{f1, l1}

	a1 := map[string]string{"tag": "a", "content": "A1"}
	b1 := map[string]string{"tag": "B", "content": "B1"}
	a2 := map[string]string{"tag": "a", "content": "A2"}
	f490 := marc.MarcField{MarcTag: "490"}
	f490.Subfields = []map[string]string{a1, b1, a2}

	fields := marc.MarcFields{f400, f490}
	bib := Bib{VarFields: fields}
	titles := bib.TitleSeries()

	if len(titles) != 3 ||
		titles[0] != "F1 P1" ||
		titles[1] != "A1" || titles[2] != "A2" {
		t.Errorf("Unexpected titles found: %#v", titles)
	}
}

func TestRegionFacet(t *testing.T) {
	z1 := map[string]string{"content": "usa", "tag": "z"}
	z2 := map[string]string{"content": "ri", "tag": "z"}
	z3 := map[string]string{"content": "zz", "tag": "z"}
	field := marc.MarcField{MarcTag: "650"}
	field.Subfields = []map[string]string{z1, z2, z3}
	fields := marc.MarcFields{field}
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
	f245 := marc.MarcField{MarcTag: "245"}
	f245.Subfields = []map[string]string{f245a, f2456}

	// title in language
	f8806 := map[string]string{"content": "245-03/$1", "tag": "6"}
	f880a := map[string]string{"content": "titulo en español:", "tag": "a"}
	f880b := map[string]string{"content": "bb", "tag": "b"}
	f880c := map[string]string{"content": "cc", "tag": "c"}
	f880 := marc.MarcField{MarcTag: "880"}
	f880.Subfields = []map[string]string{f8806, f880a, f880b, f880c}

	fields := marc.MarcFields{f245, f880}
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
	f130 := marc.MarcField{MarcTag: "130"}
	f130.Subfields = []map[string]string{f130a, f130l}
	fields := marc.MarcFields{f130}
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
	f240 := marc.MarcField{MarcTag: "240"}
	f240.Subfields = []map[string]string{f240a, f240k}
	fields = marc.MarcFields{f240}
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
	f240 := marc.MarcField{MarcTag: "240"}
	f240.Subfields = []map[string]string{f240a, f2406}

	// title in language
	f8806 := map[string]string{"content": "240-02/$1", "tag": "6"}
	f880a := map[string]string{"content": "titulo en español.", "tag": "a"}
	f880 := marc.MarcField{MarcTag: "880"}
	f880.Subfields = []map[string]string{f8806, f880a}

	fields := marc.MarcFields{f240, f880}
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
	field := marc.MarcField{MarcTag: "260"}
	field.Subfields = []map[string]string{s1, s2, s3, s4, s5, s6}
	fields := marc.MarcFields{field}
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
	f240 := marc.MarcField{MarcTag: "240"}
	f240.Subfields = []map[string]string{f240a, f240l, f2406}

	// title in language
	f8806 := map[string]string{"content": "240-02", "tag": "6"}
	f880a := map[string]string{"content": "titulo en español.", "tag": "a"}
	f880l := map[string]string{"content": "Spanish.", "tag": "l"}
	f880 := marc.MarcField{MarcTag: "880"}
	f880.Subfields = []map[string]string{f8806, f880a, f880l}

	fields := marc.MarcFields{f240, f880}
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
	leader := marc.MarcField{FieldTag: "_", Content: "00000ngm 2200000Ia 4500"}
	f007 := marc.MarcField{MarcTag: "007", Content: "vf mbahos"}
	fields := marc.MarcFields{leader, f007}
	bib := Bib{VarFields: fields}
	code := bib.FormatCode()
	if code != "BV" {
		t.Errorf("Failed to detect a video: %#v", code)
	}
}

func TestSubjects(t *testing.T) {
	// Sample record b1000980
	a1 := map[string]string{"tag": "a", "content": "A1"}
	d1 := map[string]string{"tag": "d", "content": "D1"}
	t1 := map[string]string{"tag": "t", "content": "T1"}
	f1 := marc.MarcField{MarcTag: "600"}
	f1.Subfields = []map[string]string{a1, d1, t1}

	a2 := map[string]string{"tag": "a", "content": "A2"}
	d2 := map[string]string{"tag": "d", "content": "D2"}
	t2 := map[string]string{"tag": "t", "content": "T2"}
	f2 := marc.MarcField{MarcTag: "600"}
	f2.Subfields = []map[string]string{a2, d2, t2}

	v3 := map[string]string{"tag": "6", "content": "600-02/$1"}
	a3 := map[string]string{"tag": "a", "content": "A3"}
	d3 := map[string]string{"tag": "d", "content": "D3"}
	t3 := map[string]string{"tag": "t", "content": "T3"}
	f3 := marc.MarcField{MarcTag: "880"}
	f3.Subfields = []map[string]string{v3, a3, d3, t3}

	fields := marc.MarcFields{f1, f2, f3}
	bib := Bib{VarFields: fields}
	subjects := bib.Subjects()
	// Make sure the subjects are picked correctly...
	if subjects[0] != "A1 D1 T1" ||
		subjects[1] != "A2 D2 T2" ||
		subjects[2] != "A3 D3 T3" {
		t.Errorf("Unexpected values found (1/2): %#v", subjects)
	}

	// ...and that the "a" subfields are also picked on their own
	if subjects[3] != "A1" || subjects[4] != "A2" || subjects[5] != "A3" {
		t.Errorf("Unexpected values found (2/2): %#v", subjects)
	}
}

func TestIsDissertation(t *testing.T) {
	a1 := map[string]string{"tag": "a", "content": "something Brown University something"}
	c1 := map[string]string{"tag": "c", "content": "c"}
	f1 := marc.MarcField{MarcTag: "502"}
	f1.Subfields = []map[string]string{a1, c1}
	fields1 := marc.MarcFields{f1}
	bib1 := Bib{VarFields: fields1}
	if !bib1.IsDissertaion() {
		t.Errorf("Failed to detect dissertation")
	}

	a2 := map[string]string{"tag": "a", "content": "a"}
	d2 := map[string]string{"tag": "d", "content": "something Brown University something"}
	f2 := marc.MarcField{MarcTag: "502"}
	f2.Subfields = []map[string]string{a2, d2}
	fields2 := marc.MarcFields{f2}
	bib2 := Bib{VarFields: fields2}
	if bib2.IsDissertaion() {
		t.Errorf("Incorrectly detected a dissertation")
	}
}
