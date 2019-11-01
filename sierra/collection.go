package sierra

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type CollectionItemRow struct {
	BibRecordNum   string
	ItemRecordNum  string
	Title          string
	PublishYear    int
	BarCode        string
	LocationCode   string
	CallnumberRaw  string
	CallnumberNorm string
	Publisher      string
}

func (row CollectionItemRow) String() string {
	s := fmt.Sprintf("%s, %s, %s", row.BibRecordNum, row.ItemRecordNum, row.Title)
	return s
}

func CollectionItemsForSubject(connString string, subject string) ([]CollectionItemRow, error) {
	listID := 0
	if subject == "econ" {
		listID = 334
	}
	if listID == 0 {
		msg := fmt.Sprintf("Invalid subject (%s)", subject)
		return []CollectionItemRow{}, errors.New(msg)
	}
	return CollectionItemsForList(connString, listID)
}

func CollectionItemsForList(connString string, listID int) ([]CollectionItemRow, error) {
	log.Printf("Connecting to DB: %s", connString)
	// https://godoc.org/github.com/lib/pq
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return []CollectionItemRow{}, err
	}
	defer db.Close()

	sqlSelect := `
		SELECT bib.record_num, bib.title, bibprop.publish_year,
			i.record_num as item_record_num, i.barcode, i.location_code,
			iprop.call_number, iprop.call_number_norm,
			(
				SELECT v.field_content
				FROM sierra_view.varfield as v
				WHERE v.record_id = bib.id AND v.marc_tag='260'
				ORDER BY v.occ_num
				LIMIT 1
			) as publisher
		FROM sierra_view.bool_set AS list
		INNER JOIN sierra_view.bib_view AS bib ON (bib.id = list.record_metadata_id)
		INNER JOIN sierra_view.bib_record_item_record_link AS lk ON (bib.id = lk.bib_record_id)
		INNER JOIN sierra_view.item_view AS i ON (i.id = lk.item_record_id)
		INNER JOIN sierra_view.bib_record_property AS bibprop ON (bib.id = bibprop.bib_record_id)
		INNER JOIN sierra_view.item_record_property AS iprop ON (i.id = iprop.item_record_id)
		WHERE list.bool_info_id = {listID};`

	sqlSelect = strings.ReplaceAll(sqlSelect, "{listID}", strconv.Itoa(listID))
	log.Printf("Running query: \r\n%s\r\n", sqlSelect)

	rows, err := db.Query(sqlSelect)
	if err != nil {
		return []CollectionItemRow{}, err
	}
	defer rows.Close()

	values := []CollectionItemRow{}
	log.Printf("Fetching rows...")
	for rows.Next() {
		row, err := scanCollectionItemRow(rows)
		if err != nil {
			return []CollectionItemRow{}, err
		}
		values = append(values, row)

		count := len(values)
		if count > 0 && (count%1000) == 0 {
			log.Printf("Fetched %d rows...", count)
		}
	}
	log.Printf("Found %d rows\r\n", len(values))
	return values, nil
}

func scanCollectionItemRow(rows *sql.Rows) (CollectionItemRow, error) {
	var recordNum, itemRecordNum, pubYear sql.NullInt64
	var title, barcode, location, callNum, callNumNorm, publisher sql.NullString

	err := rows.Scan(&recordNum, &title, &pubYear, &itemRecordNum,
		&barcode, &location, &callNum, &callNumNorm, &publisher)
	if err != nil {
		return CollectionItemRow{}, err
	}

	row := CollectionItemRow{}
	row.BibRecordNum = fmt.Sprintf("b%d", intLongValue(recordNum))
	row.ItemRecordNum = fmt.Sprintf("i%d", intLongValue(itemRecordNum))
	row.Title = stringValue(title)
	row.PublishYear = int(intLongValue(pubYear))
	row.BarCode = stringValue(barcode)
	row.LocationCode = stringValue(location)
	row.CallnumberRaw = stringValue(callNum)
	row.CallnumberNorm = stringValue(callNumNorm)
	row.Publisher = stringValue(publisher)
	return row, nil
}
