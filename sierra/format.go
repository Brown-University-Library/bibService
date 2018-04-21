package sierra

var formats map[string]string

func init() {
	formats = map[string]string{
		"MANUSCRIPT":   "Archives/Manuscripts",
		"2-D GRAPHIC":  "Visual Material",
		"BOOK":         "Book",
		"MAP":          "Map",
		"MUSIC RECORD": "Sound Recording",
		"AUDIOVISUAL":  "Video",
	}

	// TODO: handle these duplicate ones
	// "2-D GRAPHIC":  "Book",
	// "BOOK":         "Periodical Title",
}

func formatName(code string) string {
	name := formats[code]
	if name == "" {
		return "Book"
	}
	return name
}

// func FormatForBib(bib BibResp) string {
// 	fixed := ""
// 	f008 := bib.MarcValue("008")
// 	if len(f008) >= 22 {
// 		fixed = f008[21:22]
// 	}
// 	return ""
// }

// Traject values
// formats = map[string]string{
// 	"AJ":  "Journal",
// 	"AN":  "Newspaper",
// 	"BI":  "Biography",
// 	"BK":  "Book",
// 	"CE":  "Data File",
// 	"CR":  "CDROM",
// 	"CS":  "Software",
// 	"DI":  "Dictionaries",
// 	"DR":  "Directories",
// 	"EN":  "Encyclopedias",
// 	"HT":  "HathiTrust",
// 	"MN":  "Maps-Atlas",
// 	"MP":  "Map",
// 	"MS":  "Musical Score",
// 	"MU":  "Music",
// 	"MV":  "Archive",
// 	"MW":  "Manuscript",
// 	"MX":  "Mixed Material",
// 	"PP":  "Photographs & Pictorial Works",
// 	"RC":  "Audio CD",
// 	"RL":  "Audio LP",
// 	"RM":  "Audio (music)",
// 	"RS":  "Audio (spoken word)",
// 	"RU":  "Audio",
// 	"SE":  "Serial",
// 	"SX":  "Serial",
// 	"VB":  "Video (Blu-ray)",
// 	"VD":  "Video (DVD)",
// 	"VG":  "Video Games",
// 	"VH":  "Video (VHS)",
// 	"VL":  "Motion Picture",
// 	"VM":  "Visual Material",
// 	"WM":  "Microform",
// 	"XC":  "Conference",
// 	"XS":  "Statistics",
// 	"XX":  "Unknown",
// 	"CF":  "Computer File",
// 	"BAM": "Archives/Manuscripts",
// 	"BV":  "Video",
// 	"BSR": "Sound Recording",
// 	"BP":  "Periodical Title",
// 	"B3D": "3D object",
// 	"BTD": "Thesis/Dissertation",
// }
