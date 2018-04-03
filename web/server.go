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

	http.HandleFunc("/bibutils/item/", itemController)
	http.HandleFunc("/bibutils/bib/", bibController)
	http.HandleFunc("/bibutils/bib", bibAbout)
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
	</ul>
	`
	fmt.Fprint(resp, html)
}

func bibAbout(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprint(resp, "Try with /bibutils/bib/your-bib-id")
}

func itemController(resp http.ResponseWriter, req *http.Request) {
	bib := bibFromPath(req.URL.Path)
	if bib == "" {
		fmt.Fprint(resp, "{\"error\": \"No BIB ID indicated\"}")
		return
	}

	model := bibModel.New(settings.SierraUrl, settings.KeySecret, settings.SessionFile)
	items, err := model.Items(bib)
	if err != nil {
		log.Printf("ERROR: %s", err)
		errMsg := fmt.Sprintf("Error getting items for BIB %s", bib)
		fmt.Fprint(resp, errMsg)
		return
	}

	body, err := toJSON(items, true)
	if err != nil {
		log.Printf("ERROR: %s", err)
		fmt.Fprint(resp, "Error converting response to JSON")
		return
	}

	fmt.Fprint(resp, body)
}

func bibController(resp http.ResponseWriter, req *http.Request) {
	bib := bibFromPath(req.URL.Path)
	if bib == "" {
		fmt.Fprint(resp, "{\"error\": \"No BIB ID indicated\"}")
		return
	}

	model := bibModel.New(settings.SierraUrl, settings.KeySecret, settings.SessionFile)
	body, err := model.Get(bib)
	if err != nil {
		log.Printf("ERROR: %s", err)
		body = fmt.Sprintf("Error getting BIB %s", bib)
	}
	fmt.Fprint(resp, body)
}
