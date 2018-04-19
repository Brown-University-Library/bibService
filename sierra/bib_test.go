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

//"Designing effective strategies for environmental education : ",
//"an evaluation of the center for environmental studies' partnership with Providence public schools /",

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
