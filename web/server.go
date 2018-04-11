package web

import (
	"bibService/bibModel"
	"fmt"
	"log"
	"net/http"
)

var settings Settings

func StartWebServer(settingsFile string) {
	var err error

	log.Printf("Loading settings from: %s", settingsFile)
	settings, err = LoadSettings(settingsFile)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/bibutils/bib/", bibController)
	http.HandleFunc("/bibutils/marc/", marcController)
	http.HandleFunc("/bibutils/item/", itemController)
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
	html := `<h1>bibService</h1>
	<p>Service for BIB record utilities</p>
	<p>Examples:</p>
	<ul>
		<li> <a href="/bibutils/bib/?bib=b8060910">BIB Record</a>
		<li> <a href="/bibutils/bib/?bib=b8060910&raw=true">BIB Record (raw)</a>
		<li> <a href="/bibutils/item/?bib=b8060910">Item level data (availability)</a>
		<li> <a href="/bibutils/item/?bib=b8060910&raw=true">Item level data (availability) (raw)</a>
		<li> <a href="/bibutils/marc/?bib=b8060910">MARC data for a BIB Record</a>
	</ul>
	<p>Troubleshooting: /bibutils/status</p>
	`
	fmt.Fprint(resp, html)
}

func bibController(resp http.ResponseWriter, req *http.Request) {
	bib := qsParam("bib", req)
	if bib != "" {
		if qsParam("raw", req) == "true" {
			log.Printf("Fetching BIB data for bib: %s %v(raw)", bib, req.URL.Query())
			model := NewBibModel()
			body, err := model.GetBibRaw(bib)
			renderJSON(resp, body, err, "bibController")
		} else {
			log.Printf("Fetching BIB data for bib: %s %v", bib, req.URL.Query())
			model := NewBibModel()
			bibs, err := model.GetBib(bib)
			renderJSON(resp, bibs, err, "bibController")
		}
	}
	from := qsParam("from", req)
	to := qsParam("to", req)
	if from != "" && to != "" {
		log.Printf("Fetching BIB data for bib since: %s-%s %v(raw)", from, to, req.URL.Query())
		model := NewBibModel()
		body, err := model.GetBibsUpdated(from, to)
		renderJSON(resp, body, err, "bibController")
	}
}

func marcController(resp http.ResponseWriter, req *http.Request) {
	bib := qsParam("bib", req)
	log.Printf("Fetching MARC for bib: %s", bib)
	model := NewBibModel()
	marcData, err := model.Marc(bib)
	if err != nil {
		log.Printf("ERROR (marcController): %s", err)
		fmt.Fprint(resp, "Error fetching MARC data")
		return
	}
	fmt.Fprint(resp, marcData)
}

func itemController(resp http.ResponseWriter, req *http.Request) {
	bib := qsParam("bib", req)
	if qsParam("raw", req) == "true" {
		log.Printf("Fetching item data for bib: %s (raw)", bib)
		model := NewBibModel()
		body, err := model.ItemsRaw(bib)
		renderJSON(resp, body, err, "itemController")
	} else {
		log.Printf("Fetching item data for bib: %s", bib)
		model := NewBibModel()
		items, err := model.Items(bib)
		renderJSON(resp, items, err, "itemController")
	}
}

func NewBibModel() bibModel.BibModel {
	return bibModel.New(settings.SierraUrl, settings.KeySecret, settings.SessionFile, settings.Verbose)
}

func renderJSON(resp http.ResponseWriter, data interface{}, errFetch error, info string) {
	if errFetch != nil {
		log.Printf("ERROR (%s): %s", info, errFetch)
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
