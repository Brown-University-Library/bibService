package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		displayHelp("")
		return
	}

	filename := os.Args[1]
	tokens := strings.Split(filepath.Base(filename), "_")
	if len(tokens) != 2 {
		log.Fatal("Filename must be in the format prefix_nnnnn.xml")
		return
	}

	prefix := tokens[0]

	solrURL := ""
	if len(os.Args) > 2 {
		solrURL = os.Args[2]
	}

	err := ImportFile(filename, prefix, solrURL)
	if err != nil {
		log.Fatal(err)
		log.Printf(solrURL)
	}
}

func displayHelp(msg string) {
	syntax := `
Imports a MARC file to Solr

Syntax:
	pod filename [solrUrl]


filename is the name of the MARC file with the data to import and
must be in the format prefix_nnnnn.xml or prefix_nnnnn.mrc
where prefix indicates the source institution (e.g. penn_ or duke_).

solrUrl is optional, if provided data will be submitted to Solr via
this URL. If not provided output is stdout.

	`
	fmt.Printf("%s%s\r\n", msg, syntax)
}
