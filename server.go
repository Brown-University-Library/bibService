package main

import (
	"bibService/bibModel"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var settings Settings

func StartWebServer(settingsFile string) {
	var err error

	log.Printf("Loading settings from: %s", settingsFile)
	settings, err = LoadSettings(settingsFile)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/bibutils/bib/", bibUtils)
	http.HandleFunc("/bibutils/bib", bibUtilsAbout)
	http.HandleFunc("/", home)
	log.Printf("Listening for requests at: http://%s", settings.ServerAddress)
	err = http.ListenAndServe(settings.ServerAddress, nil)
	if err != nil {
		log.Fatal("Failed to start the web server: ", err)
	}
}

func home(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprint(resp, "home")
}

func bibUtilsAbout(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprint(resp, "Try with /bibutils/bib/your-bib-id")
}

func bibUtils(resp http.ResponseWriter, req *http.Request) {
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

// Extracts the BIB from a URL Path. Assumes the BIB is the last segment
// of the path. For example: /whatever/whatever/bib
func bibFromPath(path string) string {
	tokens := strings.Split(path, "/")
	if len(tokens) == 0 {
		return ""
	}
	return tokens[len(tokens)-1]
}
