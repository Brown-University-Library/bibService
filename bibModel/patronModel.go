package bibModel

import (
	"bibService/sierra"
)

type PatronModel struct {
	settings Settings
	api      sierra.Sierra
}

func NewPatronModel(settings Settings) PatronModel {
	model := PatronModel{settings: settings}
	model.api = sierra.NewSierra(model.settings.SierraUrl, model.settings.KeySecret, model.settings.SessionFile)
	model.api.Verbose = settings.Verbose
	return model
}

func (model PatronModel) Checkouts(patronId string) (sierra.Checkouts, error) {
	data, err := model.api.Checkouts(patronId)
	return data, err
}

func (model PatronModel) GetBibs(checkouts sierra.Checkouts) ([]sierra.Bib, error) {
	bibs := []sierra.Bib{}
	for _, checkout := range checkouts.Entries {
		itemID := checkout.ItemID()
		item, err := model.api.GetItem(itemID)
		if err != nil {
			return []sierra.Bib{}, err
		}
		for _, bibID := range item.BibIds {
			params := map[string]string{"id": bibID}
			itemBibs, err := model.api.Get(params, true)
			if err != nil {
				return []sierra.Bib{}, err
			}
			for _, bib := range itemBibs.Entries {
				bibs = append(bibs, bib)
			}
		}
	}
	return bibs, nil
}
