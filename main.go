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

	download := len(os.Args) == 3 && os.Args[2] == "download"
	if download {
		downloadMarc(settingsFile)
		return
	}

	web.StartWebServer(settingsFile)
}

/*
table of contents (sierra_export_0007.mrc)
=907  \\$a.b10158698$b02-25-16$c05-09-94
=998  \\$ar0001$b05-09-94$cm$da  $e-$feng$gmau$h4$i1
=970  11$lCh. 1$tStatus Quaestionis$p1
*/
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
