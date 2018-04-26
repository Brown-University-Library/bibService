package sierra

import (
	// "log"
	"strings"
	"testing"
)

type value struct {
	tag     string
	content string
}

func TestSubfieldParsing(t *testing.T) {
	tagsWanted := []string{"a", "x", "b"}
	fieldValues := []value{
		value{tag: "a", content: "A"},
		value{tag: "x", content: "X"},
		value{tag: "y", content: "Y"},
		value{tag: "b", content: "B"},
		value{tag: "a", content: "A"},
		value{tag: "b", content: "B"},
		value{tag: "a", content: "A"},
		value{tag: "y", content: "Y"},
		value{tag: "a", content: "A"},
		value{tag: "a", content: "A"},
	}

	// Processes the values in a Field and outputs the tags requested.
	// The logic to group the output is a bit complex because it combines
	// the values for different tags into a single value. For example,
	// if we want tags "abc" from a field with the following information:
	//
	//    tag   content
	//    ---   -------
	//    a      A1
	//    b      B1
	//    a      A2
	//    a      A3
	//    c      C3
	//
	// it will output:
	//
	//      "A1 B1"        // combined two tags
	//      "A2"           // single tag
	//      "A3 C3"        // combined two tags
	//
	output := []string{}
	processedTags := []string{}
	batchValues := []string{}
	for _, value := range fieldValues {
		tag := value.tag
		content := value.content
		tagAlreadyProcessed := in(processedTags, tag)

		if tagAlreadyProcessed {
			// output whatever we've gathered so far...
			if len(batchValues) > 0 {
				output = append(output, strings.Join(batchValues, ""))
			}

			// start a new batch...
			processedTags = []string{}
			batchValues = []string{}
		}

		if in(tagsWanted, tag) && content != "" {
			// add value to the batch
			batchValues = append(batchValues, content)
		}
		processedTags = append(processedTags, tag)
	}

	if len(batchValues) > 0 {
		// output the last batch
		output = append(output, strings.Join(batchValues, ""))
	}

	expected := []string{"AXB", "AB", "A", "A", "A"}
	if len(output) != len(expected) {
		t.Errorf("Unexpected number of values. Got: %d Expected: %d. %#v", len(output), len(expected), output)
	} else {
		for i, value := range expected {
			if value != output[i] {
				t.Errorf("Value %d mismatch. Got: %s Expected: %s", i, output[i], value)
			}
		}
	}
}
