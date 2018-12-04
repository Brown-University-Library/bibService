package bibModel

import (
	"bibService/sierra"
	"fmt"
)

// PatronModel handles patron interactions with Sierra.
type PatronModel struct {
	settings Settings
	api      sierra.Sierra
}

// CheckedoutItem represents the bib information for a checked out item.
type CheckedoutItem struct {
	BibID     string
	BibNumber string
	Title     string
	Author    string
	DueDate   string
	ItemID    string
}

// NewPatronModel creates a new PatronModel
func NewPatronModel(settings Settings) PatronModel {
	model := PatronModel{settings: settings}
	model.api = sierra.NewSierra(model.settings.SierraUrl, model.settings.KeySecret, model.settings.SessionFile)
	model.api.Verbose = settings.Verbose
	return model
}

// Checkouts returns the raw checked out data for a given patron
func (model PatronModel) Checkouts(patronID string) (sierra.Checkouts, error) {
	data, err := model.api.Checkouts(patronID)
	return data, err
}

// CheckedoutBibs returns the BIB information for the items checked out by a given patron.
func (model PatronModel) CheckedoutBibs(patronID string) ([]CheckedoutItem, error) {

	checkouts, err := model.Checkouts(patronID)
	if err != nil {
		return []CheckedoutItem{}, err
	}

	items := []CheckedoutItem{}
	for _, checkout := range checkouts.Entries {
		itemID := checkout.ItemID()
		bibID, err := model.bibIDForItemID(itemID)
		if err != nil {
			return []CheckedoutItem{}, err
		}
		bib, err := model.api.GetBib(bibID)
		if err != nil {
			return []CheckedoutItem{}, err
		}
		item := CheckedoutItem{
			BibID:     bib.Id,
			BibNumber: "b" + bib.Id,
			Title:     bib.Title,
			Author:    bib.Author,
			DueDate:   checkout.DueDate,
			ItemID:    itemID,
		}
		items = append(items, item)
	}
	return items, nil
}

func (model PatronModel) bibIDForItemID(itemID string) (string, error) {
	sierraItem, err := model.api.GetItem(itemID)
	if err != nil {
		return "", err
	}
	if len(sierraItem.BibIds) == 0 {
		return "", fmt.Errorf("No BIB records found for item %s", itemID)
	}
	if len(sierraItem.BibIds) > 1 {
		// should I return error fmt.Errorf("Multiple BIB records found for item %s", itemID)
		return sierraItem.BibIds[0], nil
	}
	return sierraItem.BibIds[0], nil
}
