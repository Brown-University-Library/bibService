package josiah

import (
	"encoding/json"
	"io/ioutil"
)

// Settings represents a shared set of values for all models including
// information about how to connect to Sierra's API, the Sierra database,
// or our Solr server.
type Settings struct {
	ServerAddress    string `json:"serverAddress"`
	SessionFile      string `json:"sessionFile"`
	SierraURL        string `json:"sierraUrl"`
	KeySecret        string `json:"keySecret"`
	Verbose          bool   `json:"verbose"`
	SolrURL          string `json:"solrUrl"`
	RootURL          string `json:"rootUrl"`
	CachedDataPath   string `json:"cachedDataPath"`
	DbUser           string `json:"dbUser"`           // Sierra Postgres DB
	DbPassword       string `json:"dbPassword"`       // Sierra Postgres DB
	DbHost           string `json:"dbHost"`           // Sierra Postgres DB
	DbPort           int    `json:"dbPort"`           // Sierra Postgres DB
	DbName           string `json:"dbName"`           // Sierra Postgres DB
	JosiahDbHost     string `json:"josiahDbHost"`     // Josiah MySQL DB
	JosiahDbUser     string `json:"josiahDbUser"`     // Josiah MySQL DB
	JosiahDbPassword string `json:"josiahDbPassword"` // Josiah MySQL DB
	JosiahDbName     string `json:"josiahDbName"`     // Josiah MySQL DB
	BestBetsAPIKey   string `json:"bbApiKey"`         // Google API key to access the BestBets document
	BestBetsDocID    string `json:"bbDocID"`          // ID of the Google Sheet with the BestBets data
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
