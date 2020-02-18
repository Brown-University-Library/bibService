package josiah

import (
	"bibService/sierra"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Ecosystem struct {
	sierraConnString string
	josiahConnString string
}

type SummaryRow struct {
	Name  string
	Count int
}

func NewEcosystem(s, j string) Ecosystem {
	return Ecosystem{sierraConnString: s, josiahConnString: j}
}

func NewSummaryRowFromSql(name sql.NullString, count sql.NullInt64) SummaryRow {
	row := SummaryRow{}
	if name.Valid {
		row.Name = name.String
	}

	if count.Valid {
		row.Count = int(count.Int64)
	}

	return row
}

func (e Ecosystem) DownloadCollection(listID int) error {
	// Get data from Sierra's database
	items, err := sierra.CollectionItemsForList(e.sierraConnString, listID)
	if err != nil {
		return err
	}

	log.Printf("Connecting to DB: %s", e.josiahConnString)
	db, err := sql.Open("mysql", e.josiahConnString)
	if err != nil {
		return err
	}
	defer db.Close()

	// Delete previous information
	log.Printf("Deleting previous saved data in Josiah for this list %d\r\n", listID)
	sqlDelete := `DELETE FROM eco_details WHERE sierra_list = ?`
	_, err = db.Exec(sqlDelete, listID)
	if err != nil {
		return err
	}

	sqlDelete = `DELETE FROM eco_summaries WHERE sierra_list = ?`
	_, err = db.Exec(sqlDelete, listID)
	if err != nil {
		return err
	}

	// Save Sierra data in the Josiah SQL database.
	log.Printf("Saving in Josiah %d records for list %d\r\n", len(items), listID)
	batch := []sierra.CollectionItemRow{}
	for _, item := range items {
		batch = append(batch, item)
		if len(batch) == 5000 {
			err := e.saveBatch(db, batch)
			if err != nil {
				return err
			}
			batch = []sierra.CollectionItemRow{}
		}
	}
	err = e.saveBatch(db, batch)
	if err != nil {
		return err
	}

	// Calculate and save summary record for this list
	log.Printf("Saving summary in Josiah for list %d\r\n", listID)
	err = e.saveSummary(db, listID)
	if err != nil {
		return err
	}

	log.Printf("Done saving %d records in Josiah for list %d\r\n", len(items), listID)
	return nil
}

func (e Ecosystem) saveBatch(db *sql.DB, batch []sierra.CollectionItemRow) error {
	if len(batch) == 0 {
		return nil
	}

	log.Printf("Saving batch (%d records)\r\n", len(batch))
	sqlInsert := `
		INSERT INTO eco_details(
			sierra_list, bib_record_num, record_type_code, bib_id, title,
			language_code, b_code1, b_code2, b_code3, country_code,
			cataloging_date_gmt, creation_date_gmt, publish_year, author, item_record_num,
			item_type_code, barcode, i_code2, i_type_code_num, location_code,
			item_status_code, last_checkin_gmt, checkout_total, renewal_total, last_year_to_date_checkout_total,
			year_to_date_checkout_total, copy_num, checkout_statistic_group_code_num, use3_count, last_checkout_gmt,
			internal_use_count, copy_use_count, old_location_code, is_suppressed, item_creation_date_gmt,
			callnumber_raw, callnumber_norm, publisher,
			ord_record_num, fund_code, fund_code_num, fund_code_master,
			marc_tag, marc_value
		)
		VALUES(
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?, ?,
			?, ?
		)
	`
	for _, item := range batch {
		_, err := db.Exec(sqlInsert,
			item.SierraList, item.BibRecordNum, item.RecordTypeCode, item.Id, item.Title,
			item.LanguageCode, item.BCode1, item.BCode2, item.BCode3, item.CountryCode,
			item.CatalogingDateGmt, item.CreationDateGmt, item.PublishYear, item.Author, item.ItemRecordNum,
			item.ItemTypeCode, item.BarCode, item.ICode2, item.ITypeCodeNum, item.LocationCode,
			item.ItemStatusCode, item.LastCheckinGmt, item.CheckoutTotal, item.RenewalTotal, item.LastYearToDateCheckoutTotal,
			item.YearToDateCheckoutTotal, item.CopyNum, item.CheckoutStatisticGroupCodeNum, item.Use3Count, item.LastCheckoutGmt,
			item.InternalUseCount, item.CopyUseCount, item.OldLocationCode, item.IsSuppressed, item.ItemCreationDateGmt,
			item.CallnumberRaw, item.CallnumberNorm, item.Publisher,
			item.OrderRecordNum, item.FundCode, item.FundCodeNum, item.FundCodeMaster,
			item.MarcTag, item.MarcValue)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e Ecosystem) saveSummary(db *sql.DB, listID int) error {
	listName, err := sierra.CollectionName(e.sierraConnString, listID)
	if err != nil {
		return err
	}

	bibCount, itemCount, err := e.getBibCounts(db, listID)
	if err != nil {
		return err
	}

	// Calculate summary by location
	sqlSelect := `SELECT location_code AS code, count(*) AS count
		FROM eco_details
		WHERE sierra_list = {listID}
		GROUP BY location_code
		ORDER BY 2 DESC`
	locationCounts, err := e.getSummaryCounts(db, sqlSelect, listID)
	if err != nil {
		return err
	}

	// Calculate summary by call number
	sqlSelect = `SELECT substring_index(callnumber_norm,' ', 1) AS code, count(*) AS count
		FROM eco_details
    	WHERE sierra_list = {listID}
    	GROUP BY substring_index(callnumber_norm,' ', 1)
    	ORDER BY 2 DESC`
	callNoCounts, err := e.getSummaryCounts(db, sqlSelect, listID)
	if err != nil {
		return err
	}

	// Calculate summary by checkout counts
	sqlSelect = `SELECT checkout_total, count(checkout_total)
		FROM eco_details
		WHERE sierra_list = {listID}
		GROUP BY checkout_total
		ORDER BY 1 DESC`
	checkoutCounts, err := e.getSummaryCounts(db, sqlSelect, listID)
	if err != nil {
		return err
	}

	// Calculate summary by fund codes
	fundCounts, err := e.getSummaryFundCodes(db, listID)
	if err != nil {
		return err
	}

	// Save them
	sqlInsert := `
		INSERT INTO eco_summaries(
			sierra_list, list_name, bib_count, item_count, updated_date_gmt,
			locations_str, callnumbers_str, checkouts_str, fundcodes_str
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?
		)`

	log.Printf("Updating summary record for list %d", listID)
	_, err = db.Exec(sqlInsert,
		listID, listName, bibCount, itemCount, DbUtcNow(),
		ToJSON(locationCounts),
		ToJSON(callNoCounts),
		ToJSON(checkoutCounts),
		ToJSON(fundCounts))
	if err != nil {
		return err
	}

	// Calculate summary by subjects separate.
	//
	// Notice that we only get the first N subjects (this is so that
	// we don't go over the limit for TEXT fields in MySQL)
	sqlSelect = `SELECT substring_index(marc_value, '|', 2) AS code, count(id) AS count
		FROM eco_details
		WHERE sierra_list = {listID}
		GROUP BY substring_index(marc_value, '|', 2)
		ORDER BY 2 DESC
		LIMIT 100`
	subjectCounts, err := e.getSummaryCounts(db, sqlSelect, listID)
	if err != nil {
		return err
	}

	sqlUpdate := "UPDATE eco_summaries SET subjects_str = ? WHERE sierra_list = ?"
	log.Printf("Updating subjects_str for list %d", listID)
	_, err = db.Exec(sqlUpdate, ToJSON(subjectCounts), listID)
	if err != nil {
		return err
	}

	return err
}

func (e Ecosystem) getBibCounts(db *sql.DB, listID int) (int, int, error) {
	sqlSelect := `SELECT count(distinct bib_record_num), count(distinct item_record_num)
		FROM eco_details
		WHERE sierra_list = {listID}`
	sqlSelect = strings.ReplaceAll(sqlSelect, "{listID}", strconv.Itoa(listID))
	log.Printf("Running query: \r\n%s\r\n", sqlSelect)

	row := db.QueryRow(sqlSelect)
	var bibCount, itemCount int
	err := row.Scan(&bibCount, &itemCount)
	return bibCount, itemCount, err
}

func (e Ecosystem) getSummaryCounts(db *sql.DB, sqlSelect string, listID int) ([]SummaryRow, error) {
	sqlSelect = strings.ReplaceAll(sqlSelect, "{listID}", strconv.Itoa(listID))
	log.Printf("Running query: \r\n%s\r\n", sqlSelect)

	rows, err := db.Query(sqlSelect)
	if err != nil {
		return []SummaryRow{}, err
	}
	defer rows.Close()

	log.Printf("Fetching rows...")
	values := []SummaryRow{}
	var name sql.NullString
	var count sql.NullInt64
	for rows.Next() {
		err := rows.Scan(&name, &count)
		if err != nil {
			return []SummaryRow{}, err
		}
		row := NewSummaryRowFromSql(name, count)
		values = append(values, row)
	}
	return values, nil
}

func (e Ecosystem) getSummaryFundCodes(db *sql.DB, listID int) ([]SummaryRow, error) {
	// Notice that we select 3 fields here
	// (and we combined them into two below)
	sqlSelect := `SELECT fund_code, fund_code_master, count(fund_code)
		FROM eco_details
		WHERE sierra_list = {listID}
		GROUP BY fund_code, fund_code_master
		ORDER BY 3 DESC, 1 ASC`
	sqlSelect = strings.ReplaceAll(sqlSelect, "{listID}", strconv.Itoa(listID))
	log.Printf("Running query: \r\n%s\r\n", sqlSelect)

	rows, err := db.Query(sqlSelect)
	if err != nil {
		return []SummaryRow{}, err
	}
	defer rows.Close()

	log.Printf("Fetching rows...")
	values := []SummaryRow{}
	var code, master sql.NullString
	var count sql.NullInt64
	for rows.Next() {
		err := rows.Scan(&code, &master, &count)
		if err != nil {
			return []SummaryRow{}, err
		}
		row := NewSummaryRowFromSql(code, count)
		row.Name = fmt.Sprintf("%s|%s", code.String, master.String)
		values = append(values, row)
	}
	return values, nil
}
