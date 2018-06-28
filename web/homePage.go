package web

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func homePage(resp http.ResponseWriter, req *http.Request) {
	log.Printf("Home: %v", req.URL)

	html := `<h1>bibService</h1>
	<p>Service for BIB record utilities</p>

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

	<p>Troubleshooting: <a href="/status">/status</a></p>
	`
	if settings.RootUrl != "" {
		html = strings.Replace(html, "/bibutils/", settings.RootUrl, -1)
	}
	fmt.Fprint(resp, html)
}
