package sierra

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type HayRow struct {
	RecordNum    string
	DisplayOrder string
	Code2        string
	LocationCode string
	StatusCode   string
	CopyNum      string
	BarCode      string
	CallNumber   string
	BestTitle    string
}

func (row HayRow) ToTSV() string {
	s := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s",
		row.RecordNum, row.DisplayOrder, row.Code2,
		row.LocationCode, row.StatusCode, row.CopyNum,
		row.BarCode, row.CallNumber, row.BestTitle)
	return s
}

func stringValue(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func HayQuery(connString string) ([]HayRow, error) {
	log.Printf("Connecting to DB: %s", connString)
	// https://godoc.org/github.com/lib/pq
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return []HayRow{}, err
	}
	defer db.Close()

	// Query provided by Kylene
	sqlSelect := `
	SELECT i2.record_num, bo.display_order, i.icode2, i.location_code,
		i.item_status_code, i.copy_num, p.barcode, p.call_number_norm, b.best_title
	FROM sierra_view.item_view AS i2,
		sierra_view.bool_set AS bo,
		sierra_view.item_record AS i,
		sierra_view.item_record_property AS p,
		sierra_view.bib_record_item_record_link AS l,
		sierra_view.bib_record_property AS b
	WHERE ((bo.bool_info_id)=171) AND ((bo.record_metadata_id)=i2.id) AND
		i2.id = i.id AND
		i.id = p.item_record_id AND
		p.item_record_id = l.item_record_id AND
		l.bib_record_id = b.bib_record_id
	ORDER BY bo.display_order;`
	log.Printf("Running query: \r\n%s\r\n", sqlSelect)

	rows, err := db.Query(sqlSelect)
	if err != nil {
		return []HayRow{}, err
	}
	defer rows.Close()

	values := []HayRow{}
	for rows.Next() {
		row, err := scanHayRow(rows)
		if err != nil {
			return []HayRow{}, err
		}
		values = append(values, row)
	}
	return values, nil
}

func scanHayRow(rows *sql.Rows) (HayRow, error) {
	var recordNum, displayOrder, code2, locationCode, statusCode,
		copyNum, barCode, callNumber, bestTitle sql.NullString

	err := rows.Scan(&recordNum, &displayOrder, &code2,
		&locationCode, &statusCode, &copyNum, &barCode,
		&callNumber, &bestTitle)
	if err != nil {
		return HayRow{}, err
	}

	row := HayRow{}
	row.RecordNum = stringValue(recordNum)
	row.DisplayOrder = stringValue(displayOrder)
	row.Code2 = stringValue(code2)
	row.LocationCode = stringValue(locationCode)
	row.StatusCode = stringValue(statusCode)
	row.CopyNum = stringValue(copyNum)
	row.BarCode = stringValue(barCode)
	row.CallNumber = stringValue(callNumber)
	row.BestTitle = stringValue(bestTitle)
	return row, nil
}
