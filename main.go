package main

import (
	"bibService/josiah"
	"bibService/web"
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

	web.StartWebServer(settingsFile)
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
	bibService settingsFile

where settingsFile is the name of the JSON file with the configuration to use.
Content of settingsFile should be as follows:

	{
	  "serverAddress": "localhost:9001",
	  "sierraUrl": "https://your-iii-domain/iii/sierra-api/v5",
	  "keySecret": "your-key:your-secret",
	  "sessionFile": "iii_session.json"
	}
	`
	fmt.Printf("%s%s\r\n", msg, syntax)
}
