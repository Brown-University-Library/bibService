package web

import (
	"bibService/bibModel"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var settings bibModel.Settings

func StartWebServer(settingsFile string) {
	var err error

	log.Printf("Loading settings from: %s", settingsFile)
	settings, err = bibModel.LoadSettings(settingsFile)
	if err != nil {
		log.Fatal(err)
	}

	// Solr
	http.HandleFunc("/bibutils/solr/doc/", solrDoc)
	http.HandleFunc("/bibutils/solr/docFromFile/", solrDocFromFile)
	http.HandleFunc("/bibutils/solr/delete/", solrDelete)

	// Bib and Item level operation
	http.HandleFunc("/bibutils/bib/updated/", bibUpdated)
	http.HandleFunc("/bibutils/bib/deleted/", bibDeleted)
	http.HandleFunc("/bibutils/bib/suppressed/", bibSuppressed)
	http.HandleFunc("/bibutils/bib/", bibController)
	http.HandleFunc("/bibutils/item/", itemController)

	// MARC operations
	http.HandleFunc("/bibutils/marc/", marcController)

	// Misc
	http.HandleFunc("/status", status)
	http.HandleFunc("/", home)
	log.Printf("Listening for requests at: http://%s", settings.ServerAddress)
	err = http.ListenAndServe(settings.ServerAddress, nil)
	if err != nil {
		log.Fatal("Failed to start the web server: ", err)
	}
}

func status(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprint(resp, "OK")
}

func home(resp http.ResponseWriter, req *http.Request) {
	log.Printf("Home: %v", req.URL)

	html := `<h1>bibService</h1>
	<p>Service for BIB record utilities</p>

	<h2>BIB Record</h2>
	<ul>
		<li> <a href="/bibutils/bib/?bib=b8060910">BIB Record</a>
		<li> <a href="/bibutils/bib/?bib=b8060910&raw=true">BIB Record (raw)</a>
		<li> <a href="/bibutils/bib/updated/?from=2018-05-04&to=2018-05-07">BIB records updated</a>
		<li> <a href="/bibutils/bib/deleted/?from=2018-05-04&to=2018-05-07">BIB records deleted (IDs only)</a>
		<li> <a href="/bibutils/bib/suppressed/?from=2018-05-04&to=2018-05-07">BIB records suppressed (IDs only)</a>
	</ul>

	<h2>Item level</h2>
	<ul>
		<li> <a href="/bibutils/item/?bib=b8060910">Item level data (availability)</a>
		<li> <a href="/bibutils/item/?bib=b8060910&raw=true">Item level data (availability - raw)</a>
	</ul>

	<h2>MARC</h2>
	<ul>
		<li> <a href="/bibutils/marc/?bib=b8060910">MARC data for a BIB Record</a>
	</ul>

	<p>Troubleshooting: <a href="/status">/status</a></p>
	`
	if settings.RootUrl != "" {
		html = strings.Replace(html, "/bibutils/", settings.RootUrl, -1)
	}
	fmt.Fprint(resp, html)
}

func bibController(resp http.ResponseWriter, req *http.Request) {
	bib := qsParam("bib", req)
	if bib == "" {
		err := errors.New("No bib parameter was received")
		renderJSON(resp, nil, err, "bibController")
		return
	}
	model := bibModel.New(settings)
	if qsParam("raw", req) == "true" {
		log.Printf("Fetching BIB data for bib: %s %v(raw)", bib, req.URL.Query())
		body, err := model.GetBibRaw(bib)
		renderJSON(resp, body, err, "bibController")
	} else {
		log.Printf("Fetching BIB data for bib: %s %v", bib, req.URL.Query())
		bibs, err := model.GetBib(bib)
		renderJSON(resp, bibs, err, "bibController")
	}
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
	path := "/Users/hectorcorrea/dev/marc_files_json/" // Make this a parameter
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
