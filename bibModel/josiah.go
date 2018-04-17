package bibModel

import (
	"bibService/sierra"
)

type SolrDoc struct {
	Id                           string   `json:"id"`
	IsbnT                        []string `json:"isbn_t"`
	IssnT                        []string `json:"issn_t"`
	OclcT                        []string `json:"oclc_t"`
	TitleT                       []string `json:"title_t"`
	TitleDisplay                 string
	TitleVernDisplay             string   `json:"title_vern_display"`
	TitleSeriesT                 []string `json:"title_series_t"`
	TitleSort                    string   `json:"title_sort"`
	UniformTitlesDisplay         []string `json:"uniform_titles_display"`
	NewUniformTitleAuthorDisplay []string `json:"new_uniform_title_author_display"`
	UniformRelatedWorksDisplay   []string `json:"uniform_related_works_display"`
	AuthorDisplay                string   `json:"author_display"`
	AuthorVernDisplay            string   `json:"author_vern_display"`
	AuthorAddlDisplay            []string `json:"author_addl_display"`
	AuthorT                      []string `json:"author_t"`
	AuthorAddlT                  []string `json:"author_addl_t"`
	PublishedDisplay             []string `json:"published_display"`
	PublishedVernDisplay         []string `json:"published_vern_display"`
	PhysicalDisplay              []string `json:"physical_display"`
	// y                            []string `json:"abstract_display"`
	// y                            []string `json:"toc_display"`
	// y                            []string `json:"toc_970_display"`
	// y                            []string `json:"pub_date"`
	// y                            []string `json:"url_fulltext_display"`
	// y                            []string `json:"url_suppl_display"`
	// y                            []string `json:"online_b"`
	// y                            []string `json:"access_facet"`
	// y                            []string `json:"format"`
	// y                            []string `json:"author_facet"`
	// y                            []string `json:"language_facet"`
	// y                            []string `json:"building_facet"`
	// y                            []string `json:"location_code_t"`
	// y                            []string `json:"region_facet"`
	// y                            []string `json:"topic_facet"`
	// y                            []string `json:"subject_t"`
	// y                            []string `json:"callnumber_t"`
	// y                            []string `json:"text"`
	// y                            []string `json:"marc_display"`
	// y                            []string `json:"bookplate_code_facet"`
	// y                            []string `json:"bookplate_code_ss"`
}

func NewSolrDoc(bib sierra.BibResp) (SolrDoc, error) {
	doc := SolrDoc{}
	doc.Id = bib.Id
	doc.IsbnT = bib.MarcValues("020a:020z")
	doc.IssnT = bib.MarcValues("022a:022l:022y:773x:774x:776x") // separator?
	doc.OclcT = bib.MarcValues("001:035a:035z")                 // specific function?

	titleSpec := "100tflnp:110tflnp:111tfklpsv:130adfklmnoprst:210ab:222ab:"
	titleSpec += "240adfklmnoprs:242abnp:246abnp:247abnp:505t:"
	titleSpec += "700fklmnoprstv:710fklmorstv:711fklpt:730adfklmnoprstv:740ap"
	doc.TitleT = bib.MarcValues(titleSpec)
	doc.TitleDisplay = bib.MarcValue("245abfgknp")
	doc.TitleVernDisplay = bib.MarcValue("245abfgknp")
	doc.TitleSeriesT = []string{}

	seriesSpec := "400flnptv:410flnptv:411fklnptv:440ap:490a:800abcdflnpqt:810tflnp:811tfklpsv:830adfklmnoprstv"
	doc.TitleSeriesT = bib.MarcValues(seriesSpec)

	doc.TitleSort = doc.TitleDisplay // TODO
	// doc.UniformTitlesDisplay = ""
	// doc.NewUniformTitleAuthorDisplay = ""
	// doc.UniformRelatedWorksDisplay = ""
	doc.AuthorDisplay = bib.MarcValueTrim("100abcdq:110abcd:111abcd")
	doc.AuthorVernDisplay = doc.AuthorDisplay // TODO: account for alternate_script
	doc.AuthorAddlDisplay = bib.MarcValuesTrim("700abcd:710ab:711ab")
	doc.AuthorT = bib.MarcValuesTrim("100abcdq:110abcd:111abcdeq")
	doc.AuthorAddlT = bib.MarcValues("700abcdq:710abcd:711abcdeq:810abc:811aqdce")
	doc.PublishedDisplay = bib.MarcValuesTrim("260a")
	doc.PublishedVernDisplay = bib.MarcValuesTrim("260a") // TODO: account for alternate script
	doc.PhysicalDisplay = bib.MarcValues("300abcefg:530abcd")
	return doc, nil
}
