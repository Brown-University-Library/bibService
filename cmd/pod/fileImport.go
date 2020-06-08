package main

import (
	"bibService/pkg/marcimport"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hectorcorrea/marcli/pkg/marc"
	"github.com/hectorcorrea/solr"
)

func ImportFile(filename string, idPrefix string, solrUrl string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	i := 0
	stdOut := (solrUrl == "")
	solrData := ""
	marc := marc.NewMarcFile(file)
	if stdOut {
		fmt.Printf("[\n")
	}

	for marc.Scan() {
		r, err := marc.Record()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		d, err := newSolrDocFromRecord(r, idPrefix)
		if err != nil {
			return err
		}

		bytes, err := json.Marshal(d)
		if err != nil {
			return err
		}
		json := string(bytes)

		if stdOut {
			if i > 0 {
				fmt.Printf(",\n")
			}
			fmt.Printf("%s", json)
		} else {
			if i > 0 {
				solrData += ","
			}
			solrData += json
		}

		i++
	}

	if stdOut {
		fmt.Printf("\n]\n")
	} else {
		// TODO: post in batches
		solrCore := solr.New(solrUrl, false)
		err := solrCore.PostString("[" + solrData + "]")
		if err != nil {
			return err
		}
	}

	return marc.Err()
}

func newSolrDocFromRecord(rec marc.Record, idPrefix string) (marcimport.SolrDoc, error) {
	doc := marcimport.SolrDoc{}

	id := rec.ControlNum()
	if id == "" {
		return doc, errors.New("No id found")
	}
	doc.Id = []string{idPrefix + "-" + id}
	doc.IssnT = getValuesBySpec(rec, "022a:022l:022y:773x:774x:776x")

	doc.OclcT = getValuesBySpec(rec, "035a:035z")
	doc.OclcT = append(doc.OclcT, id)

	titleSpec := "100tflnp:110tflnp:111tfklpsv:130adfklmnoprst:210ab:222ab:"
	titleSpec += "240adfklmnoprs:242abnp:246abnp:247abnp:"
	titleSpec += "700fklmtnoprsv:710fklmorstv:711fklpt:730adfklmnoprstv:740ap"
	doc.TitleT = getValuesBySpec(rec, titleSpec)

	titleDisplay := getValuesBySpec(rec, "245apbfgkn")
	if len(titleDisplay) > 0 {
		doc.TitleDisplay = []string{titleDisplay[0]}
	}

	authors := getValuesBySpec(rec, "700abcd:710ab:711ab")
	if len(authors) > 0 {
		doc.AuthorDisplay = []string{authors[0]}
	}
	doc.AuthorFacet = getValuesBySpec(rec, "100abcd:110ab:111ab:700abcd:711ab")
	doc.AuthorT = getValuesBySpec(rec, "100abcdq:110abcd:111abcdeq")

	doc.AccessFacet = []string{idPrefix}
	doc.SourceInstitution = idPrefix
	return doc, nil
}

func getValuesBySpec(rec marc.Record, specStr string) []string {
	values := []string{}
	for _, spec := range marcimport.NewFieldSpecs(specStr) {
		for _, field := range rec.FieldsByTag(spec.MarcTag) {
			value := ""
			if len(spec.Subfields) == 0 {
				// No subfields indicated, return the string version of the field
				value = field.String()
			} else {
				for _, sub := range field.SubFields {
					if spec.ContainsSub(sub.Code) {
						value = sub.Value + " "
					}
				}
			}
			value = strings.Trim(value, " ")
			if value != "" {
				values = append(values, value)
			}
		}
	}
	return values
}
