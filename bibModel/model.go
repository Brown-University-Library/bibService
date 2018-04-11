package bibModel

import (
	"bibService/sierra"
	"errors"
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
	verbose     bool
}

func New(sierraUrl, keySecret, sessionFile string, verbose bool) BibModel {
	model := BibModel{
		sierraUrl:   sierraUrl,
		keySecret:   keySecret,
		sessionFile: sessionFile,
		verbose:     verbose,
	}
	return model
}

func (model BibModel) GetBib(bib string) (sierra.BibsResp, error) {
	id := idFromBib(bib)
	if id == "" {
		return sierra.BibsResp{}, errors.New("No ID was detected on BIB")
	}

	api := sierra.NewSierra(model.sierraUrl, model.keySecret, model.sessionFile)
	api.Verbose = model.verbose
	params := map[string]string{
		"id": id,
	}
	sierraBibs, err := api.Get(params)
	if err != nil {
		return sierra.BibsResp{}, err
	}
	return sierraBibs, err
}

func (model BibModel) GetBibsUpdated(fromDate, toDate string) (sierra.BibsResp, error) {
	api := sierra.NewSierra(model.sierraUrl, model.keySecret, model.sessionFile)
	api.Verbose = model.verbose
	params := map[string]string{
		"updatedDate": "[" + fromDate + "," + toDate + "]",
	}
	sierraBibs, err := api.Get(params)
	if err != nil {
		return sierra.BibsResp{}, err
	}
	return sierraBibs, err
}

func (model BibModel) GetBibRaw(bib string) (string, error) {
	id := idFromBib(bib)
	if id == "" {
		return "", errors.New("No ID was detected on BIB")
	}

	api := sierra.NewSierra(model.sierraUrl, model.keySecret, model.sessionFile)
	api.Verbose = model.verbose
	params := map[string]string{
		"id": id,
	}
	return api.GetRaw(params)
}

func (model BibModel) Marc(bib string) (string, error) {
	id := idFromBib(bib)
	if id == "" {
		return "", errors.New("No ID was detected on BIB")
	}

	api := sierra.NewSierra(model.sierraUrl, model.keySecret, model.sessionFile)
	api.Verbose = model.verbose
	return api.Marc(id)
}

func (model BibModel) ItemsRaw(bib string) (string, error) {
	id := idFromBib(bib)
	if id == "" {
		return "", errors.New("No ID was detected on BIB")
	}

	api := sierra.NewSierra(model.sierraUrl, model.keySecret, model.sessionFile)
	api.Verbose = model.verbose
	return api.ItemsRaw(id)
}

func (model BibModel) Items(bib string) (ItemsResp, error) {
	id := idFromBib(bib)
	if id == "" {
		return ItemsResp{}, errors.New("No ID was detected on BIB")
	}

	api := sierra.NewSierra(model.sierraUrl, model.keySecret, model.sessionFile)
	api.Verbose = model.verbose
	sierraItems, err := api.Items(id)
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
