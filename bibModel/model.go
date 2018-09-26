package bibModel

import (
	"bibService/sierra"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/hectorcorrea/solr"
)

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

func (model BibModel) GetBibRange(fromBib, toBib string) (sierra.Bibs, error) {
	bibs := sierra.Bibs{}
	fromId := idFromBib(fromBib)
	toId := idFromBib(toBib)
	pageNum := 0
	for {
		pageNum += 1
		page, err := model.bibRangePaginated(fromId, toId, pageNum)
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

func (model BibModel) bibRangePaginated(fromBib, toBib string, page int) (sierra.Bibs, error) {
	offset := (page - 1) * pageSize
	params := map[string]string{
		"offset": strconv.Itoa(offset),
		"limit":  strconv.Itoa(pageSize),
	}

	if fromBib == "" && fromBib == "" {
		return sierra.Bibs{}, errors.New("No BIB range was received")
	} else {
		params["id"] = fmt.Sprintf("[%s,%s]", fromBib, toBib)
	}
	return model.api.Get(params, true)
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
	beginCount, err := solrClient.Count()
	if err != nil {
		return err
	}

	deleted, err := model.GetBibsDeleted(fromDate, toDate)
	if err != nil {
		return err
	}

	if len(deleted) != 0 {
		err = solrClient.Delete(deleted)
		if err != nil {
			log.Printf("Error deleting from Solr deleted records in Sierra (%d)", len(deleted))
			return err
		}
	}

	suppressed, err := model.GetBibsSuppressed(fromDate, toDate)
	if err != nil {
		return err
	}

	if len(suppressed) != 0 {
		err = solrClient.Delete(suppressed)
		if err != nil {
			log.Printf("Error deleting from Solr suppressed records in Sierra (%d)", len(suppressed))
			return err
		}
	}

	endCount, err := solrClient.Count()
	if err != nil {
		return err
	}

	// It's possible that the totals don't add up. For example, running the delete
	// for the same date range twice will report 0 records deleted in Solr the
	// second time (even if Sierra reports that there are records deleted and
	// suppressed)
	log.Printf("Deleted %d documents from Solr (D=%d, S=%d)", beginCount-endCount, len(deleted), len(suppressed))
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
	return NewSolrDoc(bibs.Entries[0]), nil
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
