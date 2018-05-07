package bibModel

import (
	"bibService/sierra"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hectorcorrea/solr"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

func x() {
	log.Printf("dummy")
}

// 2000 seems to be the limit that Sierra imposes
const pageSize = 1000

type ShelfResp struct {
	Aisle        string `json:"aisle"`
	DisplayAisle string `json:"display_aisle"`
	Floor        string `json:"floor"`
	Located      bool   `json:"located"`
	Location     string `json:"location"`
	Side         string `json:"side"`
}

type JosiahItem struct {
	Barcode    string    `json:"barcode"`
	Callnumber string    `json:"callnumber"`
	Location   string    `json:"location"`
	MapUrl     string    `json:"map"`
	Shelf      ShelfResp `json:"shelf"`
	Status     string    `json:"status"`
}

type JosiahItems struct {
	HasMore     bool         `json:"has_more"`
	Items       []JosiahItem `json:"items"`
	MoreLink    string       `json:"more_link"`
	Requestable bool         `json:"requestable"`
	Summary     []string     `json:"summary"`
}

type BibModel struct {
	settings Settings
	api      sierra.Sierra
	solrUrl  string
}

type Range struct {
	first int
	last  int
}

func New(settings Settings) BibModel {
	model := BibModel{settings: settings}
	model.api = sierra.NewSierra(model.settings.SierraUrl, model.settings.KeySecret, model.settings.SessionFile)
	model.api.Verbose = settings.Verbose
	model.solrUrl = settings.SolrUrl
	return model
}

func (model BibModel) GetBib(bibs string) (sierra.Bibs, error) {
	ids := idsFromBib(bibs)
	if ids == "" {
		return sierra.Bibs{}, errors.New("No ID was received")
	}

	params := map[string]string{
		"id": ids,
	}
	sierraBibs, err := model.api.Get(params, true)
	if err != nil {
		return sierra.Bibs{}, err
	}
	return sierraBibs, err
}

func (model BibModel) GetBibsUpdated(fromDate, toDate string, includeItems bool) (sierra.Bibs, error) {
	bibs := sierra.Bibs{}
	pageNum := 0
	for {
		pageNum += 1
		page, err := model.bibsUpdatedPaginated(fromDate, toDate, pageNum, includeItems)
		if err != nil {
			return sierra.Bibs{}, err
		}
		for _, entry := range page.Entries {
			if !entry.Deleted {
				bibs.Total += 1
				bibs.Entries = append(bibs.Entries, entry)
			}
		}
		if page.Total < pageSize {
			break
		}
	}
	return bibs, nil
}

func (model BibModel) GetBibsSuppressed(fromDate, toDate string) ([]string, error) {
	bibs := []string{}
	pageNum := 0
	for {
		pageNum += 1
		page, err := model.bibsSuppressedPaginated(fromDate, toDate, pageNum)
		if err != nil {
			return bibs, err
		}
		for _, entry := range page.Entries {
			bibs = append(bibs, entry.Bib())
		}
		if page.Total < pageSize {
			break
		}
	}
	return bibs, nil
}

func (model BibModel) GetBibsDeleted(fromDate, toDate string) ([]string, error) {
	bibs := []string{}
	pageNum := 0
	for {
		pageNum += 1
		page, err := model.bibsDeletedPaginated(fromDate, toDate, pageNum)
		if err != nil {
			if err.Error() == "Status code 404" {
				// nothing to delete, no big deal
				return bibs, nil
			}
			return bibs, err
		}
		for _, entry := range page.Entries {
			bibs = append(bibs, entry.Bib())
		}
		if page.Total < pageSize {
			break
		}
	}
	return bibs, nil
}

// Deletes from Solr the IDs of the records that have been deleted
// in Sierra or that have been marked as Suppressed in Sierra.
func (model BibModel) Delete(fromDate, toDate string) error {
	solrClient := solr.New(model.solrUrl, true)
	deleted, err := model.GetBibsDeleted(fromDate, toDate)
	if err != nil {
		return err
	}

	if len(deleted) != 0 {
		log.Printf("Submitting %d deleted IDs for delete to Solr", len(deleted))
		err = solrClient.Delete(deleted)
		if err != nil {
			return err
		}
	}

	suppressed, err := model.GetBibsSuppressed(fromDate, toDate)
	if err != nil {
		return err
	}

	if len(suppressed) != 0 {
		log.Printf("Submitting %d suppressed IDs for delete to Solr ", len(suppressed))
		err = solrClient.Delete(suppressed)
		if err != nil {
			return err
		}
	}
	return nil
}

func (model BibModel) GetSolrDeleteQuery(fromDate, toDate string) (string, error) {
	bibs, err := model.GetBibsDeleted(fromDate, toDate)
	if len(bibs) == 0 || err != nil {
		return "", err
	}

	query := "<delete>\r\n"
	for _, bib := range bibs {
		query += fmt.Sprintf("<id>%s</id>\r\n", bib)
	}
	query += "</delete>"
	return query, nil
}

func (model BibModel) bibsDeletedPaginated(fromDate, toDate string, page int) (sierra.Bibs, error) {
	offset := (page - 1) * pageSize
	params := map[string]string{
		"offset": strconv.Itoa(offset),
		"limit":  strconv.Itoa(pageSize),
	}

	if fromDate == "" && toDate == "" {
		return sierra.Bibs{}, errors.New("No date range was received")
	} else {
		params["deletedDate"] = dateRange(fromDate, toDate)
	}
	return model.api.Get(params, false)
}

func (model BibModel) SolrDocFromFile(fileName string) (SolrDoc, error) {
	body, err := ioutil.ReadFile(fileName)
	if err != nil {
		return SolrDoc{}, err
	}

	var bibs sierra.Bibs
	err = json.Unmarshal([]byte(body), &bibs)
	if err != nil {
		return SolrDoc{}, err
	}

	if bibs.Total == 0 {
		return SolrDoc{}, err
	}
	return NewSolrDoc(bibs.Entries[0])
}

func (model BibModel) bibsUpdatedPaginated(fromDate, toDate string, page int, includeItems bool) (sierra.Bibs, error) {
	offset := (page - 1) * pageSize
	params := map[string]string{
		"offset":      strconv.Itoa(offset),
		"limit":       strconv.Itoa(pageSize),
		"updatedDate": dateRange(fromDate, toDate),
	}

	return model.api.Get(params, includeItems)
}

func (model BibModel) bibsSuppressedPaginated(fromDate, toDate string, page int) (sierra.Bibs, error) {
	offset := (page - 1) * pageSize
	params := map[string]string{
		"offset":      strconv.Itoa(offset),
		"limit":       strconv.Itoa(pageSize),
		"suppressed":  "true",
		"updatedDate": dateRange(fromDate, toDate),
	}
	return model.api.GetBibs(params)
}

func (model BibModel) GetBibRaw(bib string) (string, error) {
	id := idFromBib(bib)
	if id == "" {
		return "", errors.New("No ID was detected on BIB")
	}

	params := map[string]string{
		"id": id,
	}
	return model.api.GetRaw(params, "")
}

func (model BibModel) Marc(bib string) (string, error) {
	id := idFromBib(bib)
	if id == "" {
		return "", errors.New("No ID was detected on BIB")
	}

	return model.api.Marc(id)
}

func (model BibModel) GetMarcUpdated(fromDate, toDate string) (string, error) {
	// Breaking by fixed size ranges is very inneficient.
	// If bib 100 and 80000 are modified it will get a lot of
	// records in between unnecessarily.
	//
	// Getting individual records is not good either because
	// we hit a rate limit on the III side after 100 requested
	// files.
	//
	// We could try to calculate batches to minimize the number
	// of records per batch without requesting more than 100. Yikes.
	return "bigMarc", errors.New("not implemented")
}

func (model BibModel) ItemsRaw(bib string) (string, error) {
	id := idFromBib(bib)
	if id == "" {
		return "", errors.New("No ID was detected on BIB")
	}

	return model.api.ItemsRaw(id)
}

func (model BibModel) Items(bib string) (JosiahItems, error) {
	id := idFromBib(bib)
	if id == "" {
		return JosiahItems{}, errors.New("No ID was detected on BIB")
	}

	sierraItems, err := model.api.Items(id)
	if err != nil {
		return JosiahItems{}, err
	}

	var items JosiahItems
	for _, sierraItem := range sierraItems.Entries {
		if sierraItem.IsForBib(id) {
			item := JosiahItem{
				Barcode:    sierraItem.BarcodeClean(),
				Callnumber: "",
				Location:   sierraItem.LocationName(),
				MapUrl:     "",
				Status:     sierraItem.StatusDisplay(),
			}
			items.Items = append(items.Items, item)
		}
	}
	return items, err
}

func idsFromBib(bibs string) string {
	ids := []string{}
	for _, bib := range strings.Split(bibs, ",") {
		id := idFromBib(bib)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return strings.Join(ids, ",")
}

func idFromBib(bib string) string {
	if len(bib) < 2 || bib[0] != 'b' {
		return ""
	}
	return bib[1:len(bib)]
}

func dateRange(fromDate, toDate string) string {
	// It seems that we cannot pass a time with the date. From what I gather
	// Sierra automatically appends "00:00:00" to the fromDate and "23:59:59"
	// to the `toDate`.
	return "[" + fromDate + "," + toDate + "]"
}
