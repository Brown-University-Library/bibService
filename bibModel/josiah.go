package bibModel

import (
	"bibService/sierra"
)

type SolrDoc struct {
	Id                           []string `json:"id"`
	UpdatedDt                    []string `json:"updated_dt"`
	IsbnT                        []string `json:"isbn_t"`
	IssnT                        []string `json:"issn_t"`
	OclcT                        []string `json:"oclc_t"`
	TitleT                       []string `json:"title_t"`
	TitleDisplay                 []string `json:"title_display"`
	TitleVernDisplay             []string `json:"title_vern_display"`
	TitleSeriesT                 []string `json:"title_series_t"`
	TitleSort                    []string `json:"title_sort"`
	UniformTitlesDisplay         []string `json:"uniform_titles_display"`
	NewUniformTitleAuthorDisplay []string `json:"new_uniform_title_author_display"`
	UniformRelatedWorksDisplay   []string `json:"uniform_related_works_display"`
	AuthorDisplay                []string `json:"author_display"`
	AuthorVernDisplay            []string `json:"author_vern_display"`
	AuthorAddlDisplay            []string `json:"author_addl_display"`
	AuthorT                      []string `json:"author_t"`
	AuthorAddlT                  []string `json:"author_addl_t"`
	PublishedDisplay             []string `json:"published_display"`
	PublishedVernDisplay         []string `json:"published_vern_display"`
	PhysicalDisplay              []string `json:"physical_display"`
	AbstractDisplay              []string `json:"abstract_display"`
	// y                            []string `json:"toc_display"`
	// y                            []string `json:"toc_970_display"`
	PublicationYear    []int    `json:"pub_date"`
	UrlFullTextDisplay []string `json:"url_fulltext_display"`
	UrlSupplDisplay    []string `json:"url_suppl_display"`
	Online             []bool   `json:"online_b"`
	AccessFacet        []string `json:"access_facet"`
	Format             []string `json:"format"`
	AuthorFacet        []string `json:"author_facet"`
	LanguageFacet      []string `json:"language_facet"`
	BuildingFacet      []string `json:"building_facet"`
	LocationCodeT      []string `json:"location_code_t"`
	RegionFacet        []string `json:"region_facet"`
	TopicFacet         []string `json:"topic_facet"`
	SubjectsT          []string `json:"subject_t"`
	CallNumbers        []string `json:"callnumber_t"`
	// y                            []string `json:"text"`
	// y                            []string `json:"marc_display"`
	BookplateCodeFacet []string `json:"bookplate_code_facet"`
	BookplateCodeSS    []string `json:"bookplate_code_ss"`
}

func NewSolrDoc(bib sierra.Bib) (SolrDoc, error) {
	doc := SolrDoc{}
	doc.Id = []string{"b" + bib.Id}
	doc.UpdatedDt = []string{bib.UpdatedDate() + "T00:00:00Z"}
	doc.IsbnT = bib.Isbn()
	doc.IssnT = bib.Issn()
	doc.OclcT = bib.OclcNum()

	online := bib.IsOnline()
	doc.Online = []bool{online}
	if online {
		doc.AccessFacet = []string{"Online"}
	} else {
		doc.AccessFacet = []string{"At the library"}
	}

	doc.Format = []string{bib.Format()}
	doc.LanguageFacet = bib.Languages()

	if year, ok := bib.PublicationYear(); ok {
		doc.PublicationYear = []int{year}
	} else {
		doc.PublicationYear = []int{}
	}

	titleSpec := "100tflnp:110tflnp:111tfklpsv:130adfklmnoprst:210ab:222ab:"
	titleSpec += "240adfklmnoprs:242abnp:246abnp:247abnp:505t:"
	titleSpec += "700fklmtnoprsv:710fklmorstv:711fklpt:730adfklmnoprstv:740ap"
	doc.TitleT = bib.MarcValuesTrim(titleSpec)
	doc.TitleDisplay = []string{bib.TitleDisplay()}
	doc.TitleVernDisplay = []string{bib.TitleVernacularDisplay()}
	doc.TitleSeriesT = bib.TitleSeries()

	seriesSpec := "400flnptv:410flnptv:411fklnptv:440ap:490a:800abcdflnpqt:810tflnp:811tfklpsv:830adfklmnoprstv"
	doc.TitleSeriesT = bib.MarcValuesTrim(seriesSpec)
	doc.TitleSort = []string{bib.SortableTitle()}
	// doc.UniformTitlesDisplay = ""
	// doc.NewUniformTitleAuthorDisplay = ""
	// doc.UniformRelatedWorksDisplay = ""
	doc.AuthorDisplay = []string{bib.AuthorDisplay()}
	doc.AuthorVernDisplay = []string{bib.AuthorVernacularDisplay()}
	doc.AuthorFacet = bib.AuthorFacet()
	doc.AuthorAddlDisplay = bib.MarcValuesTrim("700abcd:710ab:711ab")
	doc.AuthorT = bib.MarcValuesTrim("100abcdq:110abcd:111abcdeq")
	doc.AuthorAddlT = bib.MarcValues("700aqbcd:710abcd:711aqbcde:810abc:811aqdce")

	doc.PublishedDisplay = bib.MarcValuesTrim("260a")
	doc.PublishedVernDisplay = []string{bib.PublishedVernacularDisplay()}
	doc.PhysicalDisplay = bib.MarcValues("300abcefg:530abcd")
	doc.AbstractDisplay = []string{bib.AbstractDisplay()}
	doc.BuildingFacet = bib.BuildingFacets()
	doc.LocationCodeT = bib.LocationCodes()
	doc.TopicFacet = bib.MarcValuesTrim("650a:690a")
	doc.SubjectsT = bib.Subjects()
	doc.CallNumbers = bib.CallNumbers()
	doc.RegionFacet = bib.RegionFacet()

	doc.UrlFullTextDisplay = bib.MarcValues("856u")
	doc.UrlSupplDisplay = bib.MarcValues("856z")

	doc.BookplateCodeFacet = bib.BookplateCodes()
	doc.BookplateCodeSS = doc.BookplateCodeFacet
	return doc, nil
}
