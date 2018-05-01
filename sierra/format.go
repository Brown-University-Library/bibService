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

func formatCode(leader string) string {
	// Leader = "00000nas a2200445 i 4500"
	//					 0123456789-123456789-123
	var recType, level string
	if len(leader) >= 8 {
		recType = leader[6:7]
		level = leader[7:8]
	}

	if recType == "m" {
		return "CF" //computer file
	}

	if recType == "t" || recType == "p" {
		return "BAM" // archival material
	}

	if recType == "r" {
		return "B3D" // 3D object
	}

	if isMusicFormat(recType, level) {
		if recType == "c" {
			return "MS" // musical score
		} else {
			return "BSR" // sound recording
		}
	}

	if isVisualMaterial(recType, level) {
		return "VM"
	}

	if isSerialFormat(recType, level) {
		return "BP"
	}

	if isMap(recType, level) {
		return "MP"
	}
	// JCB items with old style codes - consider them books.
	if recType == "a" && level == "p" {
		return "BK"
	}

	if isMixedMaterial(recType, level) {
		return "MX"
	}

	if isBookFormat(recType, level) {
		return "BK"
	}

	return "XX"
}

func isMap(recType, level string) bool {
	types := []string{"e", "f"}
	levels := []string{"a", "b", "c", "d", "i", "m", "s"}
	return in(types, recType) && in(levels, level)
}

func isMusicFormat(recType, level string) bool {
	types := []string{"c", "d", "i", "j"}
	levels := []string{"a", "b", "c", "d", "i", "m", "s"}
	return in(types, recType) && in(levels, level)
}

func isMixedMaterial(recType, level string) bool {
	types := []string{"b", "p"}
	levels := []string{"a", "b", "c", "d", "m", "s"}
	return in(types, recType) && in(levels, level)
}

func isVisualMaterial(recType, level string) bool {
	types := []string{"g", "k", "o", "r"}
	levels := []string{"a", "b", "c", "d", "i", "m", "s"}
	return in(types, recType) && in(levels, level)
}

func isSerialFormat(recType, level string) bool {
	levels := []string{"b", "s", "i"}
	if recType == "a" && in(levels, level) {
		return true
	}
	if level == "e" {
		return true
	}
	return false
}

func isBookFormat(recType, level string) bool {
	types := []string{"a", "t"}
	levels := []string{"a", "c", "d", "m"}
	return in(types, recType) && in(levels, level)
}

func formatName(code string) string {
	name := formats[code]
	if name == "" {
		return "Book"
	}
	return name
}
