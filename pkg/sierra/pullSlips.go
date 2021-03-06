package sierra

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type PullSlipRow struct {
	FlagDate       sql.NullTime
	DisplayOrder   int
	ProjectTitle   string
	ListDate       sql.NullTime
	OrderNum       string
	CallNumber     string
	CopyNum        int
	Volume         string
	BarCode        string
	Code2          string
	ItemStatusCode string
	BibRecordNum   string
	ItemRecordNum  string
	LocalTag       string
	Title          string
	Edition        string
	Publisher      string
	PubYear        string
	Author         string
	Description    string
	ItemLocation   string
	LocalNotes     string
	BndWith        bool
}

func (row PullSlipRow) String() string {
	s := fmt.Sprintf("%s, %s, %s", row.BibRecordNum, row.ItemRecordNum, row.Title)
	return s
}

func PullSlipsForList(connString string, listID int) ([]PullSlipRow, error) {
	log.Printf("Connecting to DB: %s", connString)
	// https://godoc.org/github.com/lib/pq
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return []PullSlipRow{}, err
	}
	defer db.Close()

	// Query provided by Kylene
	sqlSelect := `
	DROP TABLE IF EXISTS temp_item_data;
	DROP TABLE IF EXISTS temp_dupe;

	CREATE TEMP TABLE temp_item_data AS
	SELECT
		bo.display_order,
		(
			SELECT boi.name
			FROM sierra_view.bool_info as boi
			WHERE boi.id = bo.bool_info_id
		) as project_title,
		(
			SELECT boi.bool_gmt
			FROM sierra_view.bool_info as boi
			WHERE boi.id = bo.bool_info_id
		) as list_date,
		(
			SELECT o.record_type_code || o.record_num || 'a'
			FROM sierra_view.order_view as o, sierra_view.bib_record_order_record_link as ol
			WHERE o.id = ol.order_record_id AND ol.bib_record_id=l.bib_record_id
			LIMIT 1
		) as ordernum,
		p.call_number_norm,
		i.copy_num,
		(
			SELECT vv.field_content
	 		FROM sierra_view.varfield as vv
	 		WHERE vv.record_id = l.item_record_id AND vv.varfield_type_code = 'v'
			LIMIT 1
		) as volume,
		p.barcode,
		i.icode2,
		i2.item_status_code,
		rb.record_type_code || rb.record_num || 'a' as bib_record_num,
		ri.record_type_code || ri.record_num || 'a' as item_record_num,
		(
			SELECT regexp_replace(trim(v.field_content), '(\|[a-z]{1}Hathi Trust Report)', '', 'ig')
			FROM sierra_view.varfield as v
			WHERE v.record_id = l.bib_record_id
				AND v.marc_tag='910'
				AND v.field_content like '%Hathi%'
			ORDER BY v.occ_num
			LIMIT 1
		) as localtag,
		(
			SELECT sv.content
			FROM sierra_view.subfield_view as sv
			WHERE sv.record_id = l.bib_record_id
				AND sv.marc_tag='245'
				AND sv.tag='a'
			ORDER BY sv.occ_num
			LIMIT 1
		) as title,
		(
			SELECT regexp_replace(trim(v.field_content), '(\|[a-z]{1})', '', 'ig')
			FROM sierra_view.varfield as v
			WHERE v.record_id = l.bib_record_id AND v.marc_tag='250'
			ORDER BY v.occ_num
			LIMIT 1
		) as edition,
		(
			SELECT regexp_replace(trim(v.field_content), '(\|[a-z]{1})', '', 'ig')
			FROM sierra_view.varfield as v
			WHERE v.record_id = l.bib_record_id AND v.marc_tag='260'
			ORDER BY v.occ_num
			LIMIT 1
		) as publisher,
		(
			SELECT cf.p07 || cf.p08 || cf.p09 || cf.p10 as pubYear
			FROM sierra_view.control_field as cf
			WHERE cf.record_id = l.bib_record_id AND cf.control_num = '8'
			ORDER BY cf.occ_num
			LIMIT 1
		) as pubyear,
		(
			SELECT brp.best_author
			FROM sierra_view.bib_record_property as brp
			WHERE brp.bib_record_id = l.bib_record_id
		) as author,
		(
			SELECT regexp_replace(trim(v.field_content), '(\|[a-z]{1})', '', 'ig')
			FROM sierra_view.varfield as v
			WHERE v.record_id = l.bib_record_id AND v.marc_tag='300'
			ORDER BY v.occ_num
			LIMIT 1
		) as description,
		i.location_code iloc,
		(
			SELECT regexp_replace(trim(v.field_content), '(\|[a-z]{1})', '', 'ig')
			FROM sierra_view.varfield as v
			WHERE v.record_id = l.bib_record_id AND v.marc_tag='590'
			ORDER BY v.occ_num
			LIMIT 1
		) as localnotes
	FROM sierra_view.item_record as i
		JOIN sierra_view.bib_record_item_record_link as l ON (l.item_record_id = i.record_id)
		JOIN sierra_view.item_view as i2 ON (i2.id=i.record_id)
		JOIN sierra_view.item_record_property as p ON (p.item_record_id=i.record_id)
		JOIN sierra_view.bool_set as bo ON (bo.record_metadata_id=i.record_id)
		JOIN sierra_view.record_metadata as ri ON (ri.id = i.record_id)
		JOIN sierra_view.record_metadata as rb ON (rb.id = l.bib_record_id) AND (rb.campus_code = '')
	WHERE bo.bool_info_id={listID};

	CREATE TEMP TABLE temp_dupe AS
	SELECT count(l.bib_record_id)>1 as BNDWITH, bo.display_order
	FROM sierra_view.bib_record_item_record_link as l
		JOIN sierra_view.bool_set as bo ON (bo.record_metadata_id=l.item_record_id)
	WHERE bo.bool_info_id={listID}
	GROUP BY l.item_record_id, bo.display_order;

	SELECT Current_Date as Flag_date, t.*, du.BNDWITH
	FROM temp_item_data as t
		JOIN temp_dupe as du ON (du.display_order = t.display_order)
	ORDER BY t.display_order;`

	sqlSelect = strings.ReplaceAll(sqlSelect, "{listID}", strconv.Itoa(listID))
	log.Printf("Running query: \r\n%s\r\n", sqlSelect)

	rows, err := db.Query(sqlSelect)
	if err != nil {
		return []PullSlipRow{}, err
	}
	defer rows.Close()

	values := []PullSlipRow{}
	log.Printf("Fetching rows...")
	for rows.Next() {
		row, err := scanPullSlipRow(rows)
		if err != nil {
			return []PullSlipRow{}, err
		}
		values = append(values, row)

		count := len(values)
		if count > 0 && (count%100) == 0 {
			log.Printf("Fetched %d rows...", count)
		}
	}
	log.Printf("Found %d rows\r\n", len(values))
	return values, nil
}

func scanPullSlipRow(rows *sql.Rows) (PullSlipRow, error) {
	var displayOrder, copyNum int
	var bndWith bool
	var flagDate, listDate sql.NullTime
	var projectTitle, orderNum, callNumber, volume, barCode, code2,
		itemStatusCode, bibRecordNum, itemRecordNum, localTag, title, edition,
		publisher, pubYear, author, description, itemLocation, localNotes sql.NullString

	err := rows.Scan(&flagDate, &displayOrder, &projectTitle, &listDate,
		&orderNum, &callNumber, &copyNum, &volume,
		&barCode, &code2, &itemStatusCode, &bibRecordNum, &itemRecordNum,
		&localTag, &title, &edition, &publisher, &pubYear, &author,
		&description, &itemLocation, &localNotes, &bndWith)

	if err != nil {
		return PullSlipRow{}, err
	}

	row := PullSlipRow{}
	row.FlagDate = flagDate
	row.DisplayOrder = displayOrder
	row.ProjectTitle = stringValue(projectTitle)
	row.ListDate = listDate
	row.OrderNum = stringValue(orderNum)
	row.CallNumber = stringValue(callNumber)
	row.CopyNum = copyNum
	row.Volume = stringValue(volume)
	row.BarCode = stringValue(barCode)
	row.Code2 = stringValue(code2)
	row.ItemStatusCode = stringValue(itemStatusCode)
	row.BibRecordNum = stringValue(bibRecordNum)
	row.ItemRecordNum = stringValue(itemRecordNum)
	row.LocalTag = stringValue(localTag)
	row.Title = stringValue(title)
	row.Edition = stringValue(edition)
	row.Publisher = stringValue(publisher)
	row.PubYear = stringValue(pubYear)
	row.Author = stringValue(author)
	row.Description = stringValue(description)
	row.ItemLocation = stringValue(itemLocation)
	row.LocalNotes = stringValue(localNotes)
	row.BndWith = bndWith

	return row, nil
}
