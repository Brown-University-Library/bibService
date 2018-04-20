package sierra

var formats map[string]string

func init() {
	formats = map[string]string{
		"AJ":  "Journal",
		"AN":  "Newspaper",
		"BI":  "Biography",
		"BK":  "Book",
		"CE":  "Data File",
		"CR":  "CDROM",
		"CS":  "Software",
		"DI":  "Dictionaries",
		"DR":  "Directories",
		"EN":  "Encyclopedias",
		"HT":  "HathiTrust",
		"MN":  "Maps-Atlas",
		"MP":  "Map",
		"MS":  "Musical Score",
		"MU":  "Music",
		"MV":  "Archive",
		"MW":  "Manuscript",
		"MX":  "Mixed Material",
		"PP":  "Photographs & Pictorial Works",
		"RC":  "Audio CD",
		"RL":  "Audio LP",
		"RM":  "Audio (music)",
		"RS":  "Audio (spoken word)",
		"RU":  "Audio",
		"SE":  "Serial",
		"SX":  "Serial",
		"VB":  "Video (Blu-ray)",
		"VD":  "Video (DVD)",
		"VG":  "Video Games",
		"VH":  "Video (VHS)",
		"VL":  "Motion Picture",
		"VM":  "Visual Material",
		"WM":  "Microform",
		"XC":  "Conference",
		"XS":  "Statistics",
		"XX":  "Unknown",
		"CF":  "Computer File",
		"BAM": "Archives/Manuscripts",
		"BV":  "Video",
		"BSR": "Sound Recording",
		"BP":  "Periodical Title",
		"B3D": "3D object",
		"BTD": "Thesis/Dissertation",
	}
}

func formatName(code string) string {
	return formats[code]
}
