package web

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func homePage(resp http.ResponseWriter, req *http.Request) {
	log.Printf("Home: %v", req.URL)

	html := `<h1>Sierra utilities</h1>

	<h2>BIB Record</h2>
	<ul>
		<li> <a href="/bibutils/bib/?bib=b8060910">BIB Record</a>
		<li> <a href="/bibutils/bib/?bib=b8060910&raw=true">BIB Record (raw)</a>
		<li> <a href="/bibutils/bib/updated/?from=2018-05-04&to=2018-05-07">BIB records updated</a>
		<li> <a href="/bibutils/bib/deleted/?from=2018-05-04&to=2018-05-07">BIB records deleted (IDs only)</a>
		<li> <a href="/bibutils/bib/suppressed/?from=2018-05-04&to=2018-05-07">BIB records suppressed (IDs only)</a>
	</ul>

	<h2>Item level</h2>
	<ul>
		<li> <a href="/bibutils/item/?bib=b8060910">Item level data (availability)</a>
		<li> <a href="/bibutils/item/?bib=b8060910&raw=true">Item level data (availability - raw)</a>
	</ul>

	<h2>MARC</h2>
	<ul>
		<li> <a href="/bibutils/marc/?bib=b8060910">MARC data for a BIB Record</a>
	</ul>

	<h2>Pull Slips</h2>
	<ul>
		<li> <a href="/bibutils/pullSlips?id=171">Pull Slips (for Sierra List 171)</a>
	</ul>

	<h2>Collections</h2>
	<ul>
		<li> <a href="/collection/details?id=334">Download collection data (for Sierra List 334)</a>
		<li> <a href="/collection/import?id=334">Import collection data into Josiah (for Sierra List 334)</a>
	</ul>

	<p>Troubleshooting: <a href="/status">/status</a></p>
	`
	if settings.RootURL != "" {
		html = strings.Replace(html, "/bibutils/", settings.RootURL, -1)
	}
	fmt.Fprint(resp, html)
}
