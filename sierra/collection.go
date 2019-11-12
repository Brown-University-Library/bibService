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

// I think these types should be "sql.xxx" rather than just "xxx"
// given that we use these objects just to represent database rows.
// The problem with that approach is that the data would seriaize
// weird into JSON because of the presence of null values (see for
// example what it does with the sql.NullTime values)
type CollectionItemRow struct {
	SierraList                    int
	BibRecordNum                  int
	RecordTypeCode                string
	Id                            int
	Title                         string
	LanguageCode                  string
	BCode1                        string
	BCode2                        string
	BCode3                        string
	CountryCode                   string
	IsCourseReserve               bool
	CatalogingDateGmt             sql.NullTime
	CreationDateGmt               sql.NullTime
	PublishYear                   int
	Author                        string
	ItemRecordNum                 int
	ItemTypeCode                  string
	BarCode                       string
	ICode2                        string
	ITypeCodeNum                  int
	LocationCode                  string
	ItemStatusCode                string
	LastCheckinGmt                sql.NullTime
	CheckoutTotal                 int
	RenewalTotal                  int
	LastYearToDateCheckoutTotal   int
	YearToDateCheckoutTotal       int
	CopyNum                       int
	CheckoutStatisticGroupCodeNum int
	Use3Count                     int
	LastCheckoutGmt               sql.NullTime
	InternalUseCount              int
	CopyUseCount                  int
	OldLocationCode               string
	IsSuppressed                  bool
	ItemCreationDateGmt           sql.NullTime
	CallnumberRaw                 string
	CallnumberNorm                string
	Publisher                     string
	MarcTag                       string
	MarcValue                     string
}

func (row CollectionItemRow) String() string {
	s := fmt.Sprintf("%s, %s, %s", row.BibRecordNum, row.ItemRecordNum, row.Title)
	return s
}

func SierraListForSubject(subject string) int {
	if subject == "econ" {
		return 334
	}
	return 0
}

func CollectionItemsForSubject(connString string, subject string) ([]CollectionItemRow, error) {
	listID := SierraListForSubject(subject)
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
	SELECT main.*,
		var.marc_tag as marc_tag,
		var.field_content as marc_value
	FROM (
		SELECT 	bib.record_num, bib.record_type_code, bib.id, bib.title, bib.language_code,
			bib.bcode1, bib.bcode2, bib.bcode3, bib.country_code, bib.is_on_course_reserve,
			bib.cataloging_date_gmt, bib.record_creation_date_gmt,
			bibprop.publish_year, bibprop.best_author,
			i.record_num as item_record_num, i.record_type_code as item_type_code,
			i.barcode, i.icode2, i.itype_code_num, i.location_code, i.item_status_code,
			i.last_checkin_gmt, i.checkout_total, i.renewal_total, i.last_year_to_date_checkout_total,
			i.year_to_date_checkout_total, i.copy_num, i.checkout_statistic_group_code_num,
			i.use3_count, i.last_checkout_gmt, i.internal_use_count, i.copy_use_count,
			i.old_location_code, i.is_suppressed, i.record_creation_date_gmt as item_creation_date_gmt,
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
		WHERE list.bool_info_id = {listID}
	) AS main
	LEFT OUTER JOIN sierra_view.varfield AS var ON (main.id = var.record_id AND var.marc_tag >= '600' AND var.marc_tag <= '699')`

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
		row.SierraList = listID
		values = append(values, row)

		count := len(values)
		if count > 0 && (count%5000) == 0 {
			log.Printf("Fetched %d rows...", count)
		}
	}
	log.Printf("Found %d rows\r\n", len(values))
	return values, nil
}

func scanCollectionItemRow(rows *sql.Rows) (CollectionItemRow, error) {
	var BibRecordNum sql.NullInt64
	var RecordTypeCode sql.NullString
	var Id sql.NullInt64
	var Title sql.NullString
	var LanguageCode sql.NullString
	var BCode1, BCode2, BCode3 sql.NullString
	var CountryCode sql.NullString
	var IsCourseReserve sql.NullBool
	var CatalogingDateGmt sql.NullTime
	var CreationDateGmt sql.NullTime
	var PublishYear sql.NullInt64
	var Author sql.NullString
	var ItemRecordNum sql.NullInt64
	var ItemTypeCode sql.NullString
	var BarCode sql.NullString
	var ICode2 sql.NullString
	var ITypeCodeNum sql.NullInt64
	var LocationCode sql.NullString
	var ItemStatusCode sql.NullString
	var LastCheckinGmt sql.NullTime
	var CheckoutTotal sql.NullInt64
	var RenewalTotal sql.NullInt64
	var LastYearToDateCheckoutTotal sql.NullInt64
	var YearToDateCheckoutTotal sql.NullInt64
	var CopyNum sql.NullInt64
	var CheckoutStatisticGroupCodeNum sql.NullInt64
	var Use3Count sql.NullInt64
	var LastCheckoutGmt sql.NullTime
	var InternalUseCount sql.NullInt64
	var CopyUseCount sql.NullInt64
	var OldLocationCode sql.NullString
	var IsSuppressed sql.NullBool
	var ItemCreationDateGmt sql.NullTime
	var CallnumberRaw sql.NullString
	var CallnumberNorm sql.NullString
	var Publisher sql.NullString
	var MarcTag sql.NullString
	var MarcValue sql.NullString

	err := rows.Scan(
		&BibRecordNum,
		&RecordTypeCode,
		&Id,
		&Title,
		&LanguageCode,
		&BCode1, &BCode2, &BCode3,
		&CountryCode,
		&IsCourseReserve,
		&CatalogingDateGmt,
		&CreationDateGmt,
		&PublishYear,
		&Author,
		&ItemRecordNum,
		&ItemTypeCode,
		&BarCode,
		&ICode2,
		&ITypeCodeNum,
		&LocationCode,
		&ItemStatusCode,
		&LastCheckinGmt,
		&CheckoutTotal,
		&RenewalTotal,
		&LastYearToDateCheckoutTotal,
		&YearToDateCheckoutTotal,
		&CopyNum,
		&CheckoutStatisticGroupCodeNum,
		&Use3Count,
		&LastCheckoutGmt,
		&InternalUseCount,
		&CopyUseCount,
		&OldLocationCode,
		&IsSuppressed,
		&ItemCreationDateGmt,
		&CallnumberRaw,
		&CallnumberNorm,
		&Publisher,
		&MarcTag,
		&MarcValue)
	if err != nil {
		return CollectionItemRow{}, err
	}

	row := CollectionItemRow{}
	row.BibRecordNum = intLongValue(BibRecordNum)
	row.RecordTypeCode = stringValue(RecordTypeCode)
	row.Id = intLongValue(Id)
	row.Title = stringValue(Title)
	row.LanguageCode = stringValue(LanguageCode)
	row.BCode1 = stringValue(BCode1)
	row.BCode2 = stringValue(BCode2)
	row.BCode3 = stringValue(BCode3)
	row.CountryCode = stringValue(CountryCode)
	row.IsCourseReserve = boolValue(IsCourseReserve)
	row.CatalogingDateGmt = CatalogingDateGmt
	row.CreationDateGmt = CreationDateGmt
	row.PublishYear = intLongValue(PublishYear)
	row.Author = stringValue(Author)
	row.ItemRecordNum = intLongValue(ItemRecordNum)
	row.ItemTypeCode = stringValue(ItemTypeCode)
	row.BarCode = stringValue(BarCode)
	row.ICode2 = stringValue(ICode2)
	row.ITypeCodeNum = intLongValue(ITypeCodeNum)
	row.LocationCode = stringValue(LocationCode)
	row.ItemStatusCode = stringValue(ItemStatusCode)
	row.LastCheckinGmt = LastCheckinGmt
	row.CheckoutTotal = intLongValue(CheckoutTotal)
	row.RenewalTotal = intLongValue(RenewalTotal)
	row.LastYearToDateCheckoutTotal = intLongValue(LastYearToDateCheckoutTotal)
	row.YearToDateCheckoutTotal = intLongValue(YearToDateCheckoutTotal)
	row.CopyNum = intLongValue(CopyNum)
	row.CheckoutStatisticGroupCodeNum = intLongValue(CheckoutStatisticGroupCodeNum)
	row.Use3Count = intLongValue(Use3Count)
	row.LastCheckoutGmt = LastCheckoutGmt
	row.InternalUseCount = intLongValue(InternalUseCount)
	row.CopyUseCount = intLongValue(CopyUseCount)
	row.OldLocationCode = stringValue(OldLocationCode)
	row.IsSuppressed = boolValue(IsSuppressed)
	row.ItemCreationDateGmt = ItemCreationDateGmt
	row.CallnumberRaw = stringValue(CallnumberRaw)
	row.CallnumberNorm = stringValue(CallnumberNorm)
	row.Publisher = stringValue(Publisher)
	row.MarcTag = stringValue(MarcTag)
	row.MarcValue = stringValue(MarcValue)
	return row, nil
}
