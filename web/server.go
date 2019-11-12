package web

import (
	"bibService/josiah"
	"bibService/sierra"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

var settings josiah.Settings

// StartWebServer runs the web server.
func StartWebServer(settingsFile string) {
	var err error

	log.Printf("Loading settings from: %s", settingsFile)
	settings, err = josiah.LoadSettings(settingsFile)
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

	// Patron operations
	http.HandleFunc("/bibutils/patron/checkout/", checkoutController)

	// MARC operations
	http.HandleFunc("/bibutils/marc/", marcController)

	// Collection Dashboard
	http.HandleFunc("/collection/details", collectionDetails)
	http.HandleFunc("/collection/import", collectionImport)

	// Misc
	http.HandleFunc("/bibutils/pullSlips", pullSlips)
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

func pullSlips(resp http.ResponseWriter, req *http.Request) {
	listID := qsParamInt("id", req)
	if listID == 0 {
		err := errors.New("No id parameter was received")
		renderJSON(resp, nil, err, "pullSlips")
		return
	}

	timeout := 300 // seconds
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require connect_timeout=%d",
		settings.DbHost, settings.DbPort, settings.DbUser, settings.DbPassword, settings.DbName, timeout)
	rows, err := sierra.PullSlipsForList(connString, listID)
	if err != nil {
		log.Printf("ERROR getting data from Sierra: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Header().Add("Content-Type", "application/json")
		fmt.Fprint(resp, "[]")
		return
	}

	bytes, err := json.Marshal(rows)
	json := string(bytes)
	resp.Header().Add("Content-Type", "application/json")
	fmt.Fprint(resp, json)
}

// Downloads into Josiah's database the data for a collection
func collectionImport(resp http.ResponseWriter, req *http.Request) {
	subject := qsParam("subject", req)
	timeout := 300 // seconds

	sierraConnString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require connect_timeout=%d",
		settings.DbHost, settings.DbPort, settings.DbUser, settings.DbPassword, settings.DbName, timeout)
	josiahConnString := fmt.Sprintf("%s:%s@/%s?parseTime=true", settings.JosiahDbUser, settings.JosiahDbPassword, settings.JosiahDbName)
	e := josiah.NewEcosystem(sierraConnString, josiahConnString)
	err := e.DownloadCollection(subject)
	if err != nil {
		log.Printf("ERROR downloading collection %s: %s", subject, err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Header().Add("Content-Type", "application/json")
		fmt.Fprint(resp, "[]")
		return
	}

	resp.Header().Add("Content-Type", "application/json")
	fmt.Fprint(resp, "[]")
}

// Returns the data for a collection
func collectionDetails(resp http.ResponseWriter, req *http.Request) {
	subject := qsParam("subject", req)
	timeout := 300 // seconds
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require connect_timeout=%d",
		settings.DbHost, settings.DbPort, settings.DbUser, settings.DbPassword, settings.DbName, timeout)
	rows, err := sierra.CollectionItemsForSubject(connString, subject)
	if err != nil {
		log.Printf("ERROR getting data from Sierra: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Header().Add("Content-Type", "application/json")
		fmt.Fprint(resp, "[]")
		return
	}

	bytes, err := json.Marshal(rows)
	json := string(bytes)
	resp.Header().Add("Content-Type", "application/json")
	fmt.Fprint(resp, json)
}

func bibOne(resp http.ResponseWriter, req *http.Request) {
	bib := qsParam("bib", req)
	if bib == "" {
		err := errors.New("No bib parameter was received")
		renderJSON(resp, nil, err, "bibOne")
		return
	}
	model := josiah.NewBibModel(settings)
	if qsParam("raw", req) == "true" {
		log.Printf("Fetching BIB data for bib: %s %v(raw)", bib, req.URL.Query())
		body, err := model.GetBibRaw(bib)
		renderJSON(resp, body, err, "bibOne")
	} else {
		log.Printf("Fetching BIB data for bib: %s %v", bib, req.URL.Query())
		bibs, err := model.GetBibs(bib)
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
	model := josiah.NewBibModel(settings)
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
	model := josiah.NewBibModel(settings)
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
	model := josiah.NewBibModel(settings)
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
	model := josiah.NewBibModel(settings)
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
	model := josiah.NewBibModel(settings)
	bibs, err := model.GetBibs(bib)
	if err != nil {
		renderJSON(resp, nil, err, "bibController")
		return
	}
	if len(bibs.Entries) > 0 {
		doc := josiah.NewSolrDoc(bibs.Entries[0])
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
	model := josiah.NewBibModel(settings)
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
	model := josiah.NewBibModel(settings)
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
	model := josiah.NewBibModel(settings)
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

func checkoutController(resp http.ResponseWriter, req *http.Request) {
	patronID := qsParam("patronId", req)
	if patronID == "" {
		err := errors.New("No patronId parameter was received")
		renderJSON(resp, nil, err, "checkoutController")
		return
	}
	log.Printf("Fetching checkout information for patronId: %s", patronID)
	model := josiah.NewPatronModel(settings)
	checkouts, err := model.CheckedoutBibs(patronID)
	if err != nil {
		log.Printf("ERROR (checkoutController): %s", err)
		fmt.Fprint(resp, "Error fetching patron checkouts")
		return
	}
	renderJSON(resp, checkouts, err, "checkoutController")
}

func marcController(resp http.ResponseWriter, req *http.Request) {
	bib := qsParam("bib", req)
	if bib == "" {
		err := errors.New("No bib parameter was received")
		renderJSON(resp, nil, err, "marcController")
		return
	}
	log.Printf("Fetching MARC for bib: %s", bib)
	model := josiah.NewBibModel(settings)
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
