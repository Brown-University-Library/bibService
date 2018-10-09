package web

import (
	"bibService/bibModel"
	"bibService/sierra"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

var settings bibModel.Settings

func StartWebServer(settingsFile string) {
	var err error

	log.Printf("Loading settings from: %s", settingsFile)
	settings, err = bibModel.LoadSettings(settingsFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%#v", settings)

	// Solr
	http.HandleFunc("/bibutils/solr/doc/", solrDoc)
	http.HandleFunc("/bibutils/solr/docFromFile/", solrDocFromFile)
	http.HandleFunc("/bibutils/solr/delete/", solrDelete)

	// Bib and Item level operation
	http.HandleFunc("/bibutils/bib/updated/", bibUpdated)
	http.HandleFunc("/bibutils/bib/deleted/", bibDeleted)
	http.HandleFunc("/bibutils/bib/suppressed/", bibSuppressed)
	http.HandleFunc("/bibutils/bib/", bibOne)
	http.HandleFunc("/bibutils/bibs/", bibRange)
	http.HandleFunc("/bibutils/item/", itemController)

	// MARC operations
	http.HandleFunc("/bibutils/marc/", marcController)

	// Misc
	http.HandleFunc("/bibutils/hayQuery.tsv", hayQueryTsv)
	http.HandleFunc("/bibutils/hayQuery", hayQuery)
	http.HandleFunc("/status", status)
	http.HandleFunc("/", homePage)
	log.Printf("Listening for requests at: http://%s", settings.ServerAddress)
	err = http.ListenAndServe(settings.ServerAddress, nil)
	if err != nil {
		log.Fatal("Failed to start the web server: ", err)
	}
}

func status(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprint(resp, "OK")
}

func hayQueryTsv(resp http.ResponseWriter, req *http.Request) {
	connString := fmt.Sprintf("host=%s port=1032 user=%s password=%s dbname=iii sslmode=require",
		settings.DbHost, settings.DbUser, settings.DbPassword)
	hayRows, _ := sierra.HayQuery(connString)
	text := ""
	for _, row := range hayRows {
		text += row.ToTSV() + "\r\n"
	}
	resp.Header().Add("Content-Type", "text/plain; charset=us-ascii")
	fmt.Fprint(resp, text)
}

func hayQuery(resp http.ResponseWriter, req *http.Request) {
	connString := fmt.Sprintf("host=%s port=1032 user=%s password=%s dbname=iii sslmode=require",
		settings.DbHost, settings.DbUser, settings.DbPassword)
	hayRows, _ := sierra.HayQuery(connString)
	html := `<html>
		<body>
		<style>
			table.noborder td {
			    margin: 0px 0px 0px 0px;
			    padding: 0px 0px 0px 0px;
			}
			table.noborder {
			    border-collapse: separate;
			    border-spacing: 0px;
			    *border-collapse: expression('separate', cellSpacing = '0px');
			}
			</style>`

	html += "<table class=noborder>"
	html += "<tr style=text-align:left;>"
	html += "<th>Record Num</th>"
	html += "<th>Order</th>"
	html += "<th>Code2</th>"
	html += "<th>Location</th>"
	html += "<th>Status </th>"
	html += "<th>Copy</th>"
	html += "<th>Barcode</th>"
	html += "<th>Callnumber</th>"
	html += "<th>Best Title</th>"
	html += "</tr>\r\n"

	for i, row := range hayRows {
		rowStyle := "style=background-color:#dbe9fe;"
		if (i % 2) != 0 {
			rowStyle = ""
		}
		html += fmt.Sprintf("<tr %s>", rowStyle)
		html += "<td>" + row.RecordNum + "</td>"
		html += "<td>" + row.DisplayOrder + "</td>"
		html += "<td>" + row.Code2 + "</td>"
		html += "<td>" + row.LocationCode + "</td>"
		html += "<td>" + row.StatusCode + "</td>"
		html += "<td>" + row.CopyNum + "</td>"
		html += "<td style=width:150px>" + row.BarCode + "</td>"
		html += "<td style=width:150px>" + row.CallNumber + "</td>"
		html += "<td>" + row.BestTitle + "</td>"
		html += "</tr>\r\n"
	}
	html += "</table></body></html>"
	fmt.Fprint(resp, html)
}

func bibOne(resp http.ResponseWriter, req *http.Request) {
	bib := qsParam("bib", req)
	if bib == "" {
		err := errors.New("No bib parameter was received")
		renderJSON(resp, nil, err, "bibOne")
		return
	}
	model := bibModel.New(settings)
	if qsParam("raw", req) == "true" {
		log.Printf("Fetching BIB data for bib: %s %v(raw)", bib, req.URL.Query())
		body, err := model.GetBibRaw(bib)
		renderJSON(resp, body, err, "bibOne")
	} else {
		log.Printf("Fetching BIB data for bib: %s %v", bib, req.URL.Query())
		bibs, err := model.GetBib(bib)
		renderJSON(resp, bibs, err, "bibOne")
	}
}

func bibRange(resp http.ResponseWriter, req *http.Request) {
	from := qsParam("from", req)
	to := qsParam("to", req)
	if from == "" || to == "" {
		err := errors.New("No from/to parameters were received")
		renderJSON(resp, nil, err, "bibRange")
		return
	}
	model := bibModel.New(settings)
	log.Printf("Fetching BIB data for bibs: %s - %s", from, to)
	bibs, err := model.GetBibRange(from, to)
	renderJSON(resp, bibs, err, "bibRange")
}

func bibUpdated(resp http.ResponseWriter, req *http.Request) {
	from := qsParam("from", req)
	to := qsParam("to", req)
	if from == "" || to == "" {
		err := errors.New("No from/to parameters were received")
		renderJSON(resp, nil, err, "bibUpdated")
		return
	}
	log.Printf("Fetching BIB updated (%s - %s)", from, to)
	model := bibModel.New(settings)
	body, err := model.GetBibsUpdated(from, to, true)
	renderJSON(resp, body, err, "bibUpdated")
}

func bibDeleted(resp http.ResponseWriter, req *http.Request) {
	from := qsParam("from", req)
	to := qsParam("to", req)
	days, _ := strconv.Atoi(qsParam("days", req))
	if days != 0 {
		from, to = rangeFromDays(days)
	}
	if from == "" || to == "" {
		err := errors.New("No from/to parameters were received")
		renderJSON(resp, nil, err, "bibDeleted")
		return
	}
	log.Printf("Fetching BIB deleted (%s - %s)", from, to)
	model := bibModel.New(settings)
	body, err := model.GetBibsDeleted(from, to)
	renderJSON(resp, body, err, "bibDeleted")
}

func bibSuppressed(resp http.ResponseWriter, req *http.Request) {
	from := qsParam("from", req)
	to := qsParam("to", req)
	days, _ := strconv.Atoi(qsParam("days", req))
	if days != 0 {
		from, to = rangeFromDays(days)
	}
	if from == "" || to == "" {
		err := errors.New("No from/to parameters were received")
		renderJSON(resp, nil, err, "bibSuppressed")
		return
	}
	log.Printf("Fetching BIB suppressed (%s - %s)", from, to)
	model := bibModel.New(settings)
	body, err := model.GetBibsSuppressed(from, to)
	renderJSON(resp, body, err, "bibSuppressed")
}

func solrDoc(resp http.ResponseWriter, req *http.Request) {
	bib := qsParam("bib", req)
	if bib == "" {
		err := errors.New("No bib parameter was received")
		renderJSON(resp, nil, err, "bibController")
		return
	}
	log.Printf("Fetching SolrDoc for %s", bib)
	model := bibModel.New(settings)
	bibs, err := model.GetBib(bib)
	if err != nil {
		renderJSON(resp, nil, err, "bibController")
		return
	}
	if len(bibs.Entries) > 0 {
		doc := bibModel.NewSolrDoc(bibs.Entries[0])
		renderJSON(resp, doc, nil, "solrDoc")
	} else {
		renderJSON(resp, "", errors.New("no bibs returned"), "solrDoc")
	}
}

func solrDocFromFile(resp http.ResponseWriter, req *http.Request) {
	bib := qsParam("bib", req)
	if bib == "" {
		err := errors.New("No bib parameter was received")
		renderJSON(resp, nil, err, "solrDocFromFile")
		return
	}
	log.Printf("Generating SolrDoc from file for BIB: %s", bib)
	model := bibModel.New(settings)
	path := settings.CachedDataPath
	fileName := path + bib + ".json"
	doc, err := model.SolrDocFromFile(fileName)
	renderJSON(resp, doc, err, "solrDocFromFile")
}

func solrDelete(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		renderJSON(resp, "", errors.New("Must use HTTP POST"), "solrDelete")
		return
	}
	from := qsParam("from", req)
	to := qsParam("to", req)
	days, _ := strconv.Atoi(qsParam("days", req))
	if days != 0 {
		from, to = rangeFromDays(days)
	}
	log.Printf("Deleting from Solr (%s - %s)", from, to)
	model := bibModel.New(settings)
	err := model.Delete(from, to)
	renderJSON(resp, "OK", err, "solrDelete")
}

func itemController(resp http.ResponseWriter, req *http.Request) {
	bib := qsParam("bib", req)
	if bib == "" {
		err := errors.New("No bib parameter was received")
		renderJSON(resp, nil, err, "itemController")
		return
	}
	model := bibModel.New(settings)
	if qsParam("raw", req) == "true" {
		log.Printf("Fetching item data for bib: %s (raw)", bib)
		body, err := model.ItemsRaw(bib)
		renderJSON(resp, body, err, "itemController")
	} else {
		log.Printf("Fetching item data for bib: %s", bib)
		items, err := model.Items(bib)
		renderJSON(resp, items, err, "itemController")
	}
}

func marcController(resp http.ResponseWriter, req *http.Request) {
	bib := qsParam("bib", req)
	if bib == "" {
		err := errors.New("No bib parameter was received")
		renderJSON(resp, nil, err, "marcController")
		return
	}
	log.Printf("Fetching MARC for bib: %s", bib)
	model := bibModel.New(settings)
	marcData, err := model.Marc(bib)
	if err != nil {
		log.Printf("ERROR (marcController): %s", err)
		fmt.Fprint(resp, "Error fetching MARC data")
		return
	}
	fmt.Fprint(resp, marcData)
}

func renderJSON(resp http.ResponseWriter, data interface{}, errFetch error, info string) {
	if errFetch != nil {
		log.Printf("ERROR (%s): %s", info, errFetch)
		// Tweak this, sometimes we want to provide more information to the client.
		fmt.Fprint(resp, "Error retrieving information")
		return
	}

	if _, isString := data.(string); isString {
		resp.Header().Add("Content-Type", "application/json")
		fmt.Fprint(resp, data)
		return
	}

	// Convert the object to a string with the JSON representation
	json, err := toJSON(data, true)
	if err != nil {
		log.Printf("ERROR (%s): %s", info, err)
		fmt.Fprint(resp, "Error converting response to JSON")
		return
	}
	resp.Header().Add("Content-Type", "application/json")
	fmt.Fprint(resp, json)
}
