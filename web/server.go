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
		<li> <a href="/bibutils/bib/b8060910">BIB Record</a>
		<li> <a href="/bibutils/item/b8060910">Item level data (availability)</a>
		<li> <a href="/bibutils/marc/b8060910">MARC data for a BIB Record</a>
	</ul>
	<p>Troubleshooting: /bibutils/status</p>
	`
	fmt.Fprint(resp, html)
}

func bibController(resp http.ResponseWriter, req *http.Request) {
	bib := bibFromPath(req.URL.Path)
	if bib == "" {
		fmt.Fprint(resp, "{\"error\": \"No BIB ID indicated\"}")
		return
	}

	model := bibModel.New(settings.SierraUrl, settings.KeySecret, settings.SessionFile)
	bibs, err := model.Get(bib)
	renderJSON(resp, bibs, err, "bibController")
}

func marcController(resp http.ResponseWriter, req *http.Request) {
	bib := bibFromPath(req.URL.Path)
	dates := req.URL.Query()["since"]
	sinceDate := ""
	if len(dates) > 0 {
		sinceDate = dates[0]
	}
	if bib == "" && sinceDate == "" {
		fmt.Fprint(resp, "{\"error\": \"No BIB ID indicated\"}")
		return
	}

	var marcData string
	var err error
	model := bibModel.New(settings.SierraUrl, settings.KeySecret, settings.SessionFile)
	if bib != "" {
		log.Printf("Fetching BIB: %s", bib)
		marcData, err = model.Marc(bib, "")
	} else if sinceDate != "" {
		log.Printf("Fetching since: %s", sinceDate)
		marcData, err = model.Marc(bib, sinceDate)
	}

	if err != nil {
		log.Printf("ERROR (marcController): %s", err)
		fmt.Fprint(resp, "Error fetching MARC data")
		return
	}
	fmt.Fprint(resp, marcData)
}

func itemController(resp http.ResponseWriter, req *http.Request) {
	bib := bibFromPath(req.URL.Path)
	if bib == "" {
		fmt.Fprint(resp, "{\"error\": \"No BIB ID indicated\"}")
		return
	}

	model := bibModel.New(settings.SierraUrl, settings.KeySecret, settings.SessionFile)
	items, err := model.Items(bib)
	renderJSON(resp, items, err, "itemController")
}

func renderJSON(resp http.ResponseWriter, data interface{}, errFetch error, info string) {
	if errFetch != nil {
		log.Printf("ERROR (%s): %s", info, errFetch)
		fmt.Fprint(resp, "Error retrieving information")
		return
	}

	json, err := toJSON(data, true)
	if err != nil {
		log.Printf("ERROR (%s): %s", info, err)
		fmt.Fprint(resp, "Error converting response to JSON")
		return
	}
	fmt.Fprint(resp, json)
}
