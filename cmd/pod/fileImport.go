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
	doc.IsbnT = getValuesBySpec(rec, "020a:020z")

	doc.OclcT = getOclcValues(rec)
	if len(doc.OclcT) == 0 {
		doc.OclcT = append(doc.OclcT, id)
	}

	titleSpec := "100tflnp:110tflnp:111tfklpsv:130adfklmnoprst:210ab:222ab:"
	titleSpec += "240adfklmnoprs:242abnp:246abnp:247abnp:"
	titleSpec += "700fklmtnoprsv:710fklmorstv:711fklpt:730adfklmnoprstv:740ap"
	doc.TitleT = getValuesBySpec(rec, titleSpec)

	titleDisplay := getValuesBySpec(rec, "245apbfgkn")
	if len(titleDisplay) > 0 {
		doc.TitleDisplay = []string{titleDisplay[0]}
	}

	authors := getValuesBySpec(rec, "100abcdq:110abcd:111abcd")
	if len(authors) > 0 {
		doc.AuthorDisplay = []string{authors[0]}
	}
	// doc.AuthorVernDisplay
	doc.AuthorAddlDisplay = getValuesBySpec(rec, "700abcd:710ab:711ab")
	doc.AuthorT = getValuesBySpec(rec, "100abcdq:110abcd:111abcdeq")
	doc.AuthorAddlT = getValuesBySpec(rec, "700abcdq:710abcd:711abcdeq:810abc:811aqdce")
	//doc.AuthorSort = getValuesBySpec("100abcd:110abcd:111abc:110ab:700abcd:710ab:711ab")
	doc.AuthorFacet = getValuesBySpec(rec, "100abcd:110ab:111ab:700abcd:710ab:711ab")

	doc.PublishedDisplay = getValuesBySpec(rec, "260a")
	// doc.PublishedVernDisplay

	doc.TopicFacet = getValuesBySpec(rec, "650a:690a")

	subjectSpec := "600a:600abcdefghjklmnopqrstuvxyz:"
	subjectSpec += "610a:610abcdefghklmnoprstuvxyz:"
	subjectSpec += "611a:611acdefghjklnpqstuvxyz:"
	subjectSpec += "630a:630adefghklmnoprstvxyz:"
	subjectSpec += "648a:648avxyz:"
	subjectSpec += "650a:650abcdevxyz:"
	subjectSpec += "651a:651aevxyz:"
	subjectSpec += "653a:654abevyz:"
	subjectSpec += "654a:655abvxyz:"
	subjectSpec += "655a:656akvxyz:"
	subjectSpec += "656a:657avxyz:"
	subjectSpec += "657a:658ab:"
	subjectSpec += "658a:662abcdefgh:"
	subjectSpec += "690a:690abcdevxyz"
	doc.SubjectsT = getValuesBySpec(rec, subjectSpec)

	lang := getControlField(rec, "008")
	if len(lang) > 37 {
		doc.LanguageFacet = append(doc.LanguageFacet, lang[35:38])
	}
	langs := getValuesBySpec(rec, "041a:041d:041e:041j")
	doc.LanguageFacet = safeAppendMany(doc.LanguageFacet, langs)

	doc.AccessFacet = []string{idPrefix}
	doc.SourceInstitution = idPrefix
	return doc, nil
}

func getControlField(rec marc.Record, tag string) string {
	for _, field := range rec.FieldsByTag(tag) {
		if field.Tag == tag {
			return field.Value
		}
	}
	return ""
}

func getValuesBySpec(rec marc.Record, specStr string) []string {
	values := []string{}
	for _, spec := range marcimport.NewFieldSpecs(specStr) {
		for _, field := range rec.FieldsByTag(spec.MarcTag) {
			subCount := len(spec.Subfields)
			if subCount == 0 {
				// No subfields indicated, return the string version of the field
				panic("not supported")
				// value = field.String()
			} else if subCount == 1 {
				// Single subfield requested, add each match as its own value
				for _, sub := range field.SubFields {
					if spec.ContainsSub(sub.Code) {
						values = safeAppend(values, sub.Value)
					}
				}
			} else {
				// Multiple subfields requested, concatenate them before adding
				// them as a value
				value := ""
				for _, sub := range field.SubFields {
					if spec.ContainsSub(sub.Code) {
						value += sub.Value + " "
					}
				}
				values = safeAppend(values, value)
			}
		}
	}
	return values
}

func getOclcValues(rec marc.Record) []string {
	values := []string{}
	for _, value := range getValuesBySpec(rec, "035a:035z") {
		if strings.HasPrefix(value, "(OCoLC)") {
			values = safeAppend(values, value[7:])
		} else {
			values = safeAppend(values, value)
		}
	}
	return values
}

func safeAppend(values []string, value string) []string {
	value = strings.Trim(value, " ")
	if value == "" {
		return values
	}
	for _, v := range values {
		if v == value {
			return values
		}
	}
	return append(values, value)
}

func safeAppendMany(values []string, values2 []string) []string {
	for _, value := range values2 {
		values = safeAppend(values, value)
	}
	return values
}
