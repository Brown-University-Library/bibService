package bibModel

import (
	"encoding/json"
	"io/ioutil"
)

type Settings struct {
	ServerAddress string `json:"serverAddress"`
	SessionFile   string `json:"sessionFile"`
	SierraUrl     string `json:"sierraUrl"`
	KeySecret     string `json:"keySecret"`
	Verbose       bool   `json:"verbose"`
	SolrUrl       string `json:"solrUrl"`
	RootUrl       string `json:"rootUrl"`
}

func LoadSettings(filename string) (Settings, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return Settings{}, err
	}

	var settings Settings
	err = json.Unmarshal(bytes, &settings)
	return settings, err
}
