package sierra

import (
	"encoding/json"
	"fmt"
)

type Bibs struct {
	Total   int   `json:"total"`
	Entries []Bib `json:"entries"`
}

type marcFileResp struct {
	File        string `json:"file"`
	InputCount  int    `json:"inputRecords"`
	OutputCount int    `json:"outputRecords"`
	ErrorCount  int    `json:"errors"`
}

// GetBibs retrieves the information about of a BIB record and its ITEM information.
//
// params is meant to include a key like
//	"id" : "the-id"						to fetch one record
//	"id" : "[fromId,toId]"				to fetch by ID range
//	"updatedDate": "[dateFrom,dateTo]"	to fetch by date range
//
// TODO: make these explicit parameters instead.
func (s *Sierra) GetBibs(params map[string]string, includeItems bool) (Bibs, error) {
	fields := "fields=default,available,orders,normTitle,normAuthor,locations,varFields,fixedFields"
	body, err := s.GetRaw(params, fields)
	if err != nil {
		return Bibs{}, err
	}

	var bibs Bibs
	err = json.Unmarshal([]byte(body), &bibs)
	if err != nil {
		return Bibs{}, err
	}

	for i, bib := range bibs.Entries {
		if bib.Deleted {
			continue
		}
		if includeItems {
			// TODO: Figure out a way to get items in batch.
			//			 In a previous attempt I tried passing the BIBs as a comma
			//			 delimited string but that did not work if any of the BIBs
			//			 was deleted. Need to revisit this.
			//
			items, err := s.Items(bib.Id)
			if err != nil {
				// TODO: Figure out why some records return "404 not found"
				// even though they have not been deleted (and I think they
				// do have items)
				errorMsg := fmt.Sprintf("Error fetching items for %s", bib.Id)
				s.log(errorMsg, err.Error())
			}
			bibs.Entries[i].Items = items.ForBib(bib.Id)
		}
	}
	return bibs, err
}

// Fetches minimal information about the records,
// we could eventually return an []string but I need to
// decide how to handle deleted records in that case.
func (s *Sierra) GetBibsMinimal(params map[string]string) (Bibs, error) {
	// TODO: could I use "id,deleted"?
	fields := "fields=default"
	body, err := s.GetRaw(params, fields)
	if err != nil {
		return Bibs{}, err
	}

	var bibs Bibs
	err = json.Unmarshal([]byte(body), &bibs)
	return bibs, err
}

func (s *Sierra) BibsUpdatedSince(date string) (Bibs, error) {
	err := s.authenticate()
	if err != nil {
		return Bibs{}, err
	}

	url := s.URL + "/bibs?updatedDate=" + date
	body, err := s.httpGet(url, s.Authorization.AccessToken)
	if err != nil {
		return Bibs{}, err
	}

	var bibs Bibs
	err = json.Unmarshal([]byte(body), &bibs)
	// Should this return an array of IDs instead?
	return bibs, err
}

// BibIDForItemID returns the BibID for a given ItemID
// If the itemID is associated with more than one BIB it will return the first one.
func (s *Sierra) BibIDForItemID(itemID string) (string, error) {
	item, err := s.Item(itemID)
	if err != nil {
		return "", err
	}
	if len(item.BibIds) == 0 {
		return "", fmt.Errorf("No BIB records found for item %s", itemID)
	}
	if len(item.BibIds) > 1 {
		// should I return error fmt.Errorf("Multiple BIB records found for item %s", itemID)
		return item.BibIds[0], nil
	}
	return item.BibIds[0], nil
}

func (s *Sierra) Search(value string) (string, error) {
	err := s.authenticate()
	if err != nil {
		return "", err
	}

	url := s.URL + "/bibs"
	url += "?deleted=false&suppressed=false&fields=title,author,publishYear,updatedDate"
	body, err := s.httpGet(url, s.Authorization.AccessToken)
	return body, err
}

func (s *Sierra) GetRaw(params map[string]string, fields string) (string, error) {
	err := s.authenticate()
	if err != nil {
		return "", err
	}

	if fields == "" {
		fields = "fields=default,available,orders,normTitle,normAuthor,locations,varFields,fixedFields"
	}

	url := s.URL + "/bibs?"
	for key, value := range params {
		url += key + "=" + value + "&"
	}
	url += fields
	return s.httpGet(url, s.Authorization.AccessToken)
}

// idRange can be a single ID or a comma delimited list of IDs.
// Becareful because it seems that Sierra's backend chokes when
// the list is to long (e.g. it fails with 50 IDs)
func (s *Sierra) Marc(idRange string) (string, error) {
	err := s.authenticate()
	if err != nil {
		return "", err
	}

	// The default export table in Sierra ("b2mtab") does not include the table
	// of contents information (MARC 970). The "b2mtab.toc" export table includes
	// this data. By passing the suffix "toc" to the API we indicate Sierra to
	// use the "b2mtab.toc" export table.
	url := s.URL + "/bibs/marc?id=" + idRange + "&mapping=toc"

	body, err := s.httpGet(url, s.Authorization.AccessToken)
	if err != nil {
		return "", err
	}

	var marcFile marcFileResp
	err = json.Unmarshal([]byte(body), &marcFile)
	if err != nil {
		return "", err
	}

	data, err := s.httpGet(marcFile.File, s.Authorization.AccessToken)
	return data, err
}

func (s *Sierra) Deleted(dateRange string) (string, error) {
	err := s.authenticate()
	if err != nil {
		return "", err
	}

	url := s.URL
	if dateRange == "" {
		url += "/bibs?deleted=true"
	} else {
		// TODO: validate dateRange is in the form a,b
		url += fmt.Sprintf("/bibs?deletedDate=[%s]", dateRange)
	}
	body, err := s.httpGet(url, s.Authorization.AccessToken)
	return body, err
}
