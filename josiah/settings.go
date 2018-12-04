package josiah

import (
	"encoding/json"
	"io/ioutil"
)

// Settings represents a shared set of values for all models including
// information about how to connect to Sierra's API, the Sierra database,
// or our Solr server.
type Settings struct {
	ServerAddress  string `json:"serverAddress"`
	SessionFile    string `json:"sessionFile"`
	SierraURL      string `json:"sierraUrl"`
	KeySecret      string `json:"keySecret"`
	Verbose        bool   `json:"verbose"`
	SolrURL        string `json:"solrUrl"`
	RootURL        string `json:"rootUrl"`
	CachedDataPath string `json:"cachedDataPath"`
	DbUser         string `json:"dbUser"`
	DbPassword     string `json:"dbPassword"`
	DbHost         string `json:"dbHost"`
	DbPort         int    `json:"dbPort"`
	DbName         string `json:"dbName"`
}

// LoadSettings fetches settings information from a JSON file.
func LoadSettings(filename string) (Settings, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return Settings{}, err
	}

	var settings Settings
	err = json.Unmarshal(bytes, &settings)
	return settings, err
}
