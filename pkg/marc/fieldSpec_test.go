package marc

import (
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
