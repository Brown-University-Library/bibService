package main

import (
	"bibService/web"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		displayHelp("No settings file indicated")
		return
	}
	settingsFile := os.Args[1]
	web.StartWebServer(settingsFile)
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
