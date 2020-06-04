package main

import (
	"bibService/pkg/josiah"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		displayHelp("No settings file indicated")
		return
	}
	settingsFile := os.Args[1]

	testRun := len(os.Args) == 3 && os.Args[2] == "smoketest"
	if testRun {
		smokeTest(settingsFile)
		return
	}

	download := len(os.Args) == 3 && os.Args[2] == "download"
	if download {
		downloadMarc(settingsFile)
		return
	}

	delete := len(os.Args) == 3 && os.Args[2] == "deleteBib"
	if delete {
		deleteBib(settingsFile)
		return
	}

	StartWebServer(settingsFile)
}

func deleteBib(settingsFile string) {
	settings, err := josiah.LoadSettings(settingsFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%#v", settings)

	model := josiah.NewBibModel(settings)
	from, to := RangeFromDays(10)
	err = model.Delete(from, to)
	if err != nil {
		log.Printf("%#v", err)
		return
	}
	log.Printf("OK")
}

func downloadMarc(settingsFile string) {
	settings, err := josiah.LoadSettings(settingsFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%#v", settings)

	d := josiah.NewDownloader(settings)
	d.AddDefaultBatches()
	toc := false
	err = d.DownloadAll(toc)
	if err != nil {
		log.Printf("%#v", err)
		return
	}
	log.Printf("OK")
}

func smokeTest(settingsFile string) {
	settings, err := josiah.LoadSettings(settingsFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%#v", settings)
}

func displayHelp(msg string) {
	syntax := `

Syntax:
	bibService settingsFile [action]

where settingsFile is the name of the JSON file with the configuration to use.
Content of settingsFile should be as follows:

	{
	  "serverAddress": "localhost:9001",
	  "sierraUrl": "https://your-iii-domain/iii/sierra-api/v5",
	  "keySecret": "your-key:your-secret",
	  "sessionFile": "iii_session.json"
	}

By default a web server will be loaded to process requests. The [action]
parameter can be used to perform one-off individual actions rather than
loading the web server. The valid actions are:

	smoketest - loads the settings file and print its values
	download - downloads from Sierra all bib records as MARC files (takes 20+ hours)
	deleteBib - deletes from Solr bib records deleted from Sierra in the last 10 days
	`
	fmt.Printf("%s%s\r\n", msg, syntax)
}
