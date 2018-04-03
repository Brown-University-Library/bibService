package bibModel

import (
	"bibService/sierra"
	"errors"
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
	sierraUrl   string
	keySecret   string
	sessionFile string
}

func New(sierraUrl, keySecret, sessionFile string) BibModel {
	model := BibModel{
		sierraUrl:   sierraUrl,
		keySecret:   keySecret,
		sessionFile: sessionFile,
	}
	return model
}

func (model BibModel) Get(bib string) (sierra.BibsResp, error) {
	id := idFromBib(bib)
	if id == "" {
		return sierra.BibsResp{}, errors.New("No ID was detected on BIB")
	}

	api := sierra.NewSierra(model.sierraUrl, model.keySecret, model.sessionFile)
	api.Verbose = true
	sierraBibs, err := api.Get(id)
	if err != nil {
		return sierra.BibsResp{}, err
	}
	return sierraBibs, err
}

func (model BibModel) Marc(bib, sinceDate string) (string, error) {
	var id string
	if bib != "" {
		id = idFromBib(bib)
		if id == "" {
			return "", errors.New("No ID was detected on BIB")
		}
	} else {
		// validate date
	}

	api := sierra.NewSierra(model.sierraUrl, model.keySecret, model.sessionFile)
	api.Verbose = true
	if id != "" {
		return api.Marc(id)
	} else {
		// TODO: use a date range
		bibsData, err := api.BibsUpdatedSince("[2018-03-28,2018-04-03]")
		if err != nil {
			return "", err
		}
		ids := []string{}
		for i, bibRecord := range bibsData.Entries {
			ids = append(ids, bibRecord.Id)
			if i == 219 {
				break
			}
		}
		return api.Marc(strings.Join(ids, ","))
	}
}

func (model BibModel) Items(bib string) (ItemsResp, error) {
	id := idFromBib(bib)
	if id == "" {
		return ItemsResp{}, errors.New("No ID was detected on BIB")
	}

	sierra := sierra.NewSierra(model.sierraUrl, model.keySecret, model.sessionFile)
	sierra.Verbose = true
	sierraItems, err := sierra.Items(id)
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
