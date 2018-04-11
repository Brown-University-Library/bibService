package bibModel

import (
	"bibService/sierra"
	"errors"
	"fmt"
	"strings"
)

type ShelfResp struct {
	Aisle        string `json:"aisle"`
	DisplayAisle string `json:"display_aisle"`
	Floor        string `json:"floor"`
	Located      bool   `json:"located"`
	Location     string `json:"location"`
	Side         string `json:"side"`
}

type ItemResp struct {
	Barcode    string    `json:"barcode"`
	Callnumber string    `json:"callnumber"`
	Location   string    `json:"location"`
	MapUrl     string    `json:"map"`
	Shelf      ShelfResp `json:"shelf"`
	Status     string    `json:"status"`
}

type ItemsResp struct {
	HasMore     bool       `json:"has_more"`
	Items       []ItemResp `json:"items"`
	MoreLink    string     `json:"more_link"`
	Requestable bool       `json:"requestable"`
	Summary     []string   `json:"summary"`
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

func (model BibModel) GetBib(bib string) (sierra.BibsResp, error) {
	id := idFromBib(bib)
	if id == "" {
		return sierra.BibsResp{}, errors.New("No ID was detected on BIB")
	}

	params := map[string]string{
		"id": id,
	}
	sierraBibs, err := model.api.Get(params)
	if err != nil {
		return sierra.BibsResp{}, err
	}
	return sierraBibs, err
}

func (model BibModel) GetBibsUpdated(fromDate, toDate string) (sierra.BibsResp, error) {
	params := map[string]string{
		"updatedDate": dateRange(fromDate, toDate),
	}
	sierraBibs, err := model.api.Get(params)
	if err != nil {
		return sierra.BibsResp{}, err
	}
	return sierraBibs, err
}

func (model BibModel) GetBibsDeleted(fromDate, toDate string) (sierra.BibsResp, error) {
	// TODO: add support for "deleted": true
	// will the response serialize correctly?
	params := map[string]string{
		"deletedDate": dateRange(fromDate, toDate),
		"limit":       "100",
	}
	sierraBibs, err := model.api.Get(params)
	if err != nil {
		return sierra.BibsResp{}, err
	}
	return sierraBibs, err
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
	if err != nil {
		return "", err
	}
	query := fmt.Sprintf("<delete><query>id:(%s)</query></delete>", strings.Join(bibs, " OR "))
	return query, nil
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

func (model BibModel) Items(bib string) (ItemsResp, error) {
	id := idFromBib(bib)
	if id == "" {
		return ItemsResp{}, errors.New("No ID was detected on BIB")
	}

	sierraItems, err := model.api.Items(id)
	if err != nil {
		return ItemsResp{}, err
	}

	var items ItemsResp
	for _, sierraItem := range sierraItems.Entries {
		if sierraItem.IsForBib(id) {
			item := ItemResp{
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

func idFromBib(bib string) string {
	if len(bib) < 2 || bib[0] != 'b' {
		return ""
	}
	return bib[1:len(bib)]
}

func dateRange(fromDate, toDate string) string {
	// TODO: handle times and their URL encoding more gracefully
	// %3A means ":"
	return "[" + fromDate + "," + toDate + "]"
}
