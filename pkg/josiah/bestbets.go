package josiah

import (
	"context"
	"strings"

	"github.com/hectorcorrea/solr"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type bbRow struct {
	Name        string `json:"name"`
	Database    string `json:"database"`
	Queries     string `json:"queries"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

func (row bbRow) ID() string {
	return Slugify(row.Name)
}

func (row bbRow) Terms() []string {
	terms := []string{}
	for _, token := range strings.Split(row.Queries, ";") {
		term := strings.Trim(token, " ")
		if term != "" {
			term = strings.ToLower(term)
			terms = append(terms, term)
		}
	}
	return terms
}

func (t *BestBetsTable) AddRow(name, db, queries, url, description string) {
	row := bbRow{Name: name, Database: db, Queries: queries, URL: url, Description: description}
	t.Rows = append(t.Rows, row)
}

type BestBetsTable struct {
	Rows []bbRow `json:"rows"`
}

type BestBets struct {
	APIKey     string
	DocumentID string
}

// NewBestBets create a new class to download the BestBets Google Sheet data.
// 	apiKey: An API key defined in https://console.developers.google.com/apis/credentials?project=bestbets-143514
// 	docId: The ID of the Google Sheet with the BestBets data. This document must be
//			shared with access "Anyone with the link can view" so that they API Key can
//			have access to it.
//
func NewBestBets(apiKey, docID string) BestBets {
	return BestBets{APIKey: apiKey, DocumentID: docID}
}

// Download the content of the spreadsheet in the range indicated.
// The range is meant to be in the form "A1:D6".
// The result is a JSON string.
//
//
// References
// 		https://cloud.google.com/docs/authentication?_ga=2.72342995.-1974404554.1582571299
//		https://github.com/googleapis/google-api-go-client/blob/master/sheets/v4/sheets-gen.go
// 		https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/get?apix_params=%7B%22spreadsheetId%22%3A%221YACxwpx4HJUZnZwvqYBuAws_zY4sk1JaGSgju3IDhnY%22%2C%22range%22%3A%22A1%3AB2%22%7D
func (bb BestBets) Download(rangeStr string) (BestBetsTable, error) {

	// Connect to the Google Sheets service
	ctx := context.Background()
	sheetsService, err := sheets.NewService(ctx, option.WithAPIKey(bb.APIKey))
	if err != nil {
		return BestBetsTable{}, err
	}

	// Fetch the data for the BestBet Google sheet
	//
	// Note: This assume that the spreadsheet has been make available to
	// "Anyone who has the link can view". This is done via the Share
	// option in the Google sheet.
	sheet := sheetsService.Spreadsheets.Values.Get(bb.DocumentID, rangeStr)
	data, err := sheet.Context(ctx).Do()
	if err != nil {
		return BestBetsTable{}, err
	}

	// Copy the sheet data to our own struct
	table := BestBetsTable{}
	for _, row := range data.Values {
		var name, db, queries, url, description string
		if len(row) > 0 {
			name = row[0].(string)
		}
		if len(row) > 1 {
			db = row[1].(string)
		}
		if len(row) > 2 {
			queries = row[2].(string)
		}
		if len(row) > 3 {
			url = row[3].(string)
		}
		if len(row) > 4 {
			description = row[4].(string)
		}
		table.AddRow(name, db, queries, url, description)
	}

	return table, nil
}

func (bb BestBets) UpdateSolr(data BestBetsTable, solrURL string, deleteAll bool) error {
	solrCore := solr.New(solrURL, false)

	if deleteAll {
		err := solrCore.DeleteAll()
		if err != nil {
			return err
		}
	}

	docs := []solr.Document{}
	for _, row := range data.Rows {
		doc := solr.NewDocument()
		doc.Data["id"] = row.ID()
		doc.Data["name_display"] = row.Name
		doc.Data["url_display"] = row.URL
		doc.Data["description_display"] = row.Description
		doc.Data["term"] = row.Terms()
		docs = append(docs, doc)
	}
	err := solrCore.PostDocs(docs)
	return err
}
