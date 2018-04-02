package main

import (
	"flag"
)

func main() {
	var settingsFile string
	flag.StringVar(&settingsFile, "settings", "settings.json", "Required. Settings file")
	flag.Parse()
	StartWebServer(settingsFile)
}
