package sierra

type BibsResp struct {
	Total   int       `json:"total"`
	Entries []BibResp `json:"entries"`
}

type BibResp struct {
	Id           string            `json:"id"`
	UpdatedDate  string            `json:"updatedDate"`
	CreatedDate  string            `json:"createdDate"`
	Deleted      bool              `json:"deleted"`
	Suppressed   bool              `json:"suppressed"`
	Available    bool              `json:"available"`
	Lang         map[string]string `json:"lang"`
	Title        string            `json:"title"`
	Author       string            `json:"author"`
	MaterialType map[string]string `json:"materialType"`
	BibLevel     map[string]string `json:"bibLevel"`
	PublishYear  int               `json:"publishYear"`
	CatalogDate  string            `json:"catalogDate"`
	Country      map[string]string `json:"country"`
	NormTitle    string            `json:"normTitle"`
	NormAuthor   string            `json:"normAuthor"`
	// Locations    []map[string]string `json:"locations"`
	VarFields []VarFieldResp `json:"varFields"`
}

type VarFieldResp struct {
	FieldTag  string              `json:"fieldTag"`
	MarcTag   string              `json:"marcTag"`
	Ind1      string              `json:"ind1"`
	Ind2      string              `json:"ind2"`
	Subfields []map[string]string `json:"subfields"`
}
