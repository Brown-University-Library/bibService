package sierra

import (
	"strings"
)

var buildings map[string]string

func init() {
	buildings = map[string]string{
		"q":     "Annex",
		"h":     "Hay",
		"j":     "John Carter Brown",
		"o":     "Orwig",
		"r":     "Rockefeller",
		"s":     "Sciences",
		"a":     "Hay",
		"cass":  "Rockefeller",
		"chin":  "Rockefeller",
		"chref": "Rockefeller",
		"cours": "Rockefeller",
		"eacg":  "Rockefeller",
		"eacr":  "Rockefeller",
		"eacs":  "Rockefeller",
		"japan": "Rockefeller",
		"jaref": "Rockefeller",
		"jmap":  "Rockefeller",
		"koref": "Rockefeller",
		"linc":  "Hay",
		"linrf": "Hay",
		"lowc":  "Hay",
		"mddvd": "Sciences",
		"mdvid": "Sciences",
		"stor":  "Rockefeller",
		"vc":    "John Carter Brown",
		"xdoc":  "Rockefeller",
		"xfch":  "Rockefeller",
		"xrom":  "Rockefeller",
		"xxxxx": "Rockefeller",
		"zd":    "Rockefeller",
		"gar":   "Rockefeller",
		"zdcom": "Rockefeller",
	}
}

func buildingName(code string) string {
	if code == "" {
		return ""
	}
	code = strings.ToLower(code)
	name := buildings[code]
	if name == "" {
		name = buildings[code[0:1]]
	}
	return name
}
