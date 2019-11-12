package josiah

import (
	"bibService/sierra"
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type Ecosystem struct {
	sierraConnString string
	josiahConnString string
}

func NewEcosystem(s, j string) Ecosystem {
	return Ecosystem{sierraConnString: s, josiahConnString: j}
}

func (e Ecosystem) DownloadCollection(subject string) error {
	// Get data from Sierra's database
	items, err := sierra.CollectionItemsForSubject(e.sierraConnString, subject)
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
	listID := sierra.SierraListForSubject(subject)
	log.Printf("Deleting previous saved data in Josiah for this list %d\r\n", listID)
	sqlDelete := `DELETE FROM eco_details WHERE sierra_list = ?`
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
			callnumber_raw, callnumber_norm, publisher, marc_tag, marc_value
		)
		VALUES(
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?
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
			item.CallnumberRaw, item.CallnumberNorm, item.Publisher, item.MarcTag, item.MarcValue)
		if err != nil {
			return err
		}
	}

	return nil
}
