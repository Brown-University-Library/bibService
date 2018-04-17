package bibModel

import (
	"bibService/sierra"
	"errors"
	"fmt"
	"strconv"
	"strings"
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
}

func New(settings Settings) BibModel {
	model := BibModel{settings: settings}
	model.api = sierra.NewSierra(model.settings.SierraUrl, model.settings.KeySecret, model.settings.SessionFile)
	model.api.Verbose = settings.Verbose
	return model
}

func (model BibModel) GetBib(bibs string) (sierra.BibsResp, error) {
	ids := idsFromBib(bibs)
	if ids == "" {
		return sierra.BibsResp{}, errors.New("No ID was received")
	}

	params := map[string]string{
		"id": ids,
	}
	sierraBibs, err := model.api.Get(params)
	if err != nil {
		return sierra.BibsResp{}, err
	}
	return sierraBibs, err
}

func (model BibModel) GetBibsUpdated(fromDate, toDate string) (sierra.BibsResp, error) {
	bibs := sierra.BibsResp{}
	pageNum := 0
	for {
		pageNum += 1
		page, err := model.bibsUpdatedPaginated(fromDate, toDate, pageNum)
		if err != nil {
			return sierra.BibsResp{}, err
		}
		bibs.Total += page.Total
		for _, entry := range page.Entries {
			bibs.Entries = append(bibs.Entries, entry)
		}
		if page.Total <= pageSize {
			break
		}
	}
	return bibs, nil
}

func (model BibModel) GetBibsDeleted(fromDate, toDate string) (sierra.BibsResp, error) {
	bibs := sierra.BibsResp{}
	pageNum := 0
	for {
		pageNum += 1
		page, err := model.bibsDeletedPaginated(fromDate, toDate, pageNum)
		if err != nil {
			return sierra.BibsResp{}, err
		}
		bibs.Total += page.Total
		for _, entry := range page.Entries {
			bibs.Entries = append(bibs.Entries, entry)
		}
		if page.Total < pageSize {
			break
		}
	}
	return bibs, nil
}

func (model BibModel) GetSolrBibsToDelete(fromDate, toDate string) ([]string, error) {
	sierraBibs, err := model.GetBibsDeleted(fromDate, toDate)
	if err != nil {
		return []string{}, err
	}

	bibs := []string{}
	for _, bib := range sierraBibs.Entries {
		bibs = append(bibs, "b"+bib.Id)
	}
	return bibs, nil
}

func (model BibModel) GetSolrDeleteQuery(fromDate, toDate string) (string, error) {
	bibs, err := model.GetSolrBibsToDelete(fromDate, toDate)
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

func (model BibModel) bibsDeletedPaginated(fromDate, toDate string, page int) (sierra.BibsResp, error) {
	offset := (page - 1) * pageSize
	params := map[string]string{
		"offset": strconv.Itoa(offset),
		"limit":  strconv.Itoa(pageSize),
	}

	if fromDate == "" && toDate == "" {
		params["deleted"] = "true"
	} else {
		params["deletedDate"] = dateRange(fromDate, toDate)
	}

	return model.api.Get(params)
}

func (model BibModel) bibsUpdatedPaginated(fromDate, toDate string, page int) (sierra.BibsResp, error) {
	offset := (page - 1) * pageSize
	params := map[string]string{
		"offset":      strconv.Itoa(offset),
		"limit":       strconv.Itoa(pageSize),
		"updatedDate": dateRange(fromDate, toDate),
	}

	return model.api.Get(params)
}

func (model BibModel) GetBibRaw(bib string) (string, error) {
	id := idFromBib(bib)
	if id == "" {
		return "", errors.New("No ID was detected on BIB")
	}

	params := map[string]string{
		"id": id,
	}
	return model.api.GetRaw(params)
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
