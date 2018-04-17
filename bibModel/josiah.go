package bibModel

import (
	"bibService/sierra"
	"log"
)

type JosiahSolr struct {
	Id     string   `json:"id"`
	IsbnT  []string `json:"isbn_t"`
	TitleT []string `json:"title_t"`
}

func NewJosiahSolr(sierraBib sierra.BibResp) (JosiahSolr, error) {
	doc := JosiahSolr{}
	doc.Id = sierraBib.Id
	// 'isbn_t', extract_marc('020a:020z')
	log.Printf("005: %s", sierraBib.MarcValues("005"))
	log.Printf("003: %s", sierraBib.MarcValues("003"))
	log.Printf("700: %v", sierraBib.MarcValues("700aqde"))
	log.Printf("author: %v", sierraBib.MarcValues("100abcdq:110abcd:111abcdeq"))
	doc.IsbnT = sierraBib.MarcValues("020a:020z")
	/*
	  100tflnp
	  110tflnp
	  111tfklpsv
	  130adfklmnoprst
	  210ab
	  222ab
	  240adfklmnoprs
	  242abnp
	  246abnp
	  247abnp
	  505t
	  700fklmnoprstv
	  710fklmorstv
	  711fklpt
	  730adfklmnoprstv
	  740ap
	*/
	doc.TitleT = sierraBib.MarcValues("090ab")
	//doc.UpdatedDt = 907b
	return doc, nil
}
