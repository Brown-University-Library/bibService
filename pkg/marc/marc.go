// Package MARC includes structs and functions to operate on MARC records.
//
// Struct MarcField represents a single field, whereas MarcFields is meant
// to represent a collection of fields (e.g. a MARC record)
package marc

import "regexp"

var reTrailingPunct *regexp.Regexp
var reTrailingPeriod *regexp.Regexp
var reSquareBracket *regexp.Regexp

func init() {
	// RegEx stolen from Traject's marc21.rb
	// https://github.com/traject/traject/blob/master/lib/traject/macros/marc21.rb
	//
	// # trailing: comma, slash, semicolon, colon (possibly preceded and followed by whitespace)
	// str = str.sub(/ *[ ,\/;:] *\Z/, '')
	reTrailingPunct = regexp.MustCompile(" *[ ,\\/;:] *$")

	// # trailing period if it is preceded by at least three letters (possibly preceded and followed by whitespace)
	// str = str.sub(/( *\w\w\w)\. *\Z/, '\1')
	reTrailingPeriod = regexp.MustCompile("( *\\w\\w\\w)\\. *$")

	// # single square bracket characters if they are the start
	// # and/or end chars and there are no internal square brackets.
	// str = str.sub(/\A\[?([^\[\]]+)\]?\Z/, '\1')
	reSquareBracket = regexp.MustCompile("^\\[?([^\\[\\]]+)\\]?$")
}
