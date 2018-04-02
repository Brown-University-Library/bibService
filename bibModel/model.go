package bibModel

import (
	"bibService/sierra"
	"errors"
)

type BibModel struct {
	sierraUrl   string
	keySecret   string
	sessionFile string
}

func New(sierraUrl, keySecret, sessionFile string) BibModel {
	model := BibModel{
		sierraUrl:   sierraUrl,
		keySecret:   keySecret,
		sessionFile: sessionFile,
	}
	return model
}

func (model BibModel) Get(bib string) (string, error) {
	id := idFromBib(bib)
	if id == "" {
		return "", errors.New("No ID was detected on BIB")
	}

	sierra := sierra.NewSierra(model.sierraUrl, model.keySecret, model.sessionFile)
	return sierra.Get(id)
}

func idFromBib(bib string) string {
	if len(bib) < 2 || bib[0] != 'b' {
		return ""
	}
	return bib[1:len(bib)]
}
