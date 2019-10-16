package main

import (
	"bibService/josiah"
	"bibService/sierra"
	"bibService/web"
	"encoding/json"
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

	timeout := 300 // seconds
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require CommandTimeout=%d ",
		settings.DbHost, settings.DbPort, settings.DbUser, settings.DbPassword, settings.DbName, timeout)
	hayRows, err := sierra.HayQuery(connString)
	if err != nil {
		log.Printf("ERROR getting data from Sierra: %s", err)
		return
	}

	bytes, err := json.Marshal(hayRows)
	json := string(bytes)
	fmt.Printf("%s", json)
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
