package sierra

type BibsResp struct {
	Total   int       `json:"total"`
	Entries []BibResp `json:"entries"`
}

type BibResp struct {
	Id           string            `json:"id"`
	UpdatedDate  string            `json:"updatedDate,omitempty"`
	CreatedDate  string            `json:"createdDate,omitempty"`
	DeletedDate  string            `json:"deletedDate,omitempty"`
	Deleted      bool              `json:"deleted,omitempty"`
	Suppressed   bool              `json:"suppressed,omitempty"`
	Available    bool              `json:"available,omitempty"`
	Lang         map[string]string `json:"lang,omitempty"`
	Title        string            `json:"title,omitempty"`
	Author       string            `json:"author,omitempty"`
	MaterialType map[string]string `json:"materialType,omitempty"`
	BibLevel     map[string]string `json:"bibLevel,omitempty"`
	PublishYear  int               `json:"publishYear,omitempty"`
	CatalogDate  string            `json:"catalogDate,omitempty"`
	Country      map[string]string `json:"country,omitempty"`
	NormTitle    string            `json:"normTitle,omitempty"`
	NormAuthor   string            `json:"normAuthor,omitempty"`
	// Locations    []map[string]string `json:"locations"`
	VarFields []VarFieldResp `json:"varFields,omitempty"`
}

type VarFieldResp struct {
	FieldTag  string              `json:"fieldTag"`
	MarcTag   string              `json:"marcTag"`
	Ind1      string              `json:"ind1"`
	Ind2      string              `json:"ind2"`
	Subfields []map[string]string `json:"subfields"`
}
