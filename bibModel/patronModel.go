package bibModel

import (
	"bibService/sierra"
)

// PatronModel handles patron interactions with Sierra.
type PatronModel struct {
	settings Settings
	sierra   sierra.Sierra
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
	model.sierra = sierra.NewSierra(model.settings.SierraURL, model.settings.KeySecret, model.settings.SessionFile)
	model.sierra.Verbose = settings.Verbose
	return model
}

// Checkedouts returns the raw checked out data for a given patron
func (patron PatronModel) checkedouts(patronID string) (sierra.Checkouts, error) {
	data, err := patron.sierra.Checkouts(patronID)
	return data, err
}

// CheckedoutBibs returns the BIB information for the items checked out by a given patron.
func (patron PatronModel) CheckedoutBibs(patronID string) ([]CheckedoutItem, error) {
	checkouts, err := patron.checkedouts(patronID)
	if err != nil {
		return []CheckedoutItem{}, err
	}

	// Get the BIB data associated with each checked out ITEM.
	items := []CheckedoutItem{}
	for _, checkout := range checkouts.Entries {
		itemID := checkout.ItemID()
		bibID, err := patron.sierra.BibIDForItemID(itemID)
		if err != nil {
			return []CheckedoutItem{}, err
		}
		bib, err := patron.sierra.GetBib(bibID)
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
