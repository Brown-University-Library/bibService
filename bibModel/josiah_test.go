package bibModel

import (
	"bibService/sierra"
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestX(t *testing.T) {

	bytes, err := ioutil.ReadFile("b1318777.json")
	if err != nil {
		t.Errorf("Error reading test file: %s", err)
	}

	var bibs sierra.BibsResp
	err = json.Unmarshal(bytes, &bibs)
	if err != nil {
		t.Errorf("Error parsing JSON: %s", err)
	}

	_, err = NewJosiahSolr(bibs.Entries[0])
	if err != nil {
		t.Errorf("Error getting MARC data: %s", err)
	}
	// t.Errorf("%v", doc)
}
