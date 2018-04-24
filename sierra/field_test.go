package sierra

import (
// "testing"
)

// func TestIsVernacularFor(t *testing.T) {
// 	sub1 := map[string]string{"content": "710-05/$1", "tag": "6"}
// 	sub2 := map[string]string{"content": "gabriel garcia", "tag": "a"}
// 	sub3 := map[string]string{"content": "author", "tag": "e"}
// 	field := Field{MarcTag: "880"}
// 	field.Subfields = []map[string]string{sub1, sub2, sub3}
//
// 	spec710, _ := NewFieldSpec("710ab")
// 	if !field.IsVernacularFor(spec710) {
// 		t.Errorf("Failed to detect 710 vernacular")
// 	}
//
// 	spec245, _ := NewFieldSpec("245ab")
// 	if field.IsVernacularFor(spec245) {
// 		t.Errorf("Detected unexpected vernacular")
// 	}
// }

// func TestVernacularValues(t *testing.T) {
// 	sub1 := map[string]string{"content": "710-05/$1", "tag": "6"}
// 	sub2 := map[string]string{"content": "gabriel x garcia", "tag": "a"}
// 	sub3 := map[string]string{"content": "author", "tag": "e"}
// 	sub4 := map[string]string{"content": "gabriel y garcia", "tag": "a"}
// 	field := Field{MarcTag: "880"}
// 	field.Subfields = []map[string]string{sub1, sub2, sub3, sub4}
//
// 	spec710, _ := NewFieldSpec("710ab")
// 	values := field.VernacularValues(spec710)
// 	if len(values) != 2 {
// 		t.Errorf("Failed to fetch vernacular values (%d)", len(values))
// 	}
//
// 	spec245, _ := NewFieldSpec("245ab")
// 	values = field.VernacularValues(spec245)
// 	if len(values) > 0 {
// 		t.Errorf("Fetched unexpected vernacular values (%d)", len(values))
// 	}
// }
