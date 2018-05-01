package bibModel

import (
	"bibService/sierra"
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestX(t *testing.T) {

	bytes, err := ioutil.ReadFile("b1318777_test.json")
	if err != nil {
		t.Errorf("Error reading test file: %s", err)
	}

	var bibs sierra.Bibs
	err = json.Unmarshal(bytes, &bibs)
	if err != nil {
		t.Errorf("Error parsing JSON: %s", err)
	}

	_, err = NewSolrDoc(bibs.Entries[0])
	if err != nil {
		t.Errorf("Error getting MARC data: %s", err)
	}
	// t.Errorf("%v", doc)
}
