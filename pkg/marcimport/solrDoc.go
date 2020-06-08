package marcimport

// SolrDoc represents a Solr document with bibliographic data
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
	TableOfContents              []string `json:"toc_display"`
	TableOfContents970           []string `json:"toc_970_display"`
	PublicationYear              []int    `json:"pub_date"`
	UrlFullTextDisplay           []string `json:"url_fulltext_display"`
	UrlSupplDisplay              []string `json:"url_suppl_display"`
	Online                       []bool   `json:"online_b"`
	AccessFacet                  []string `json:"access_facet"`
	Format                       []string `json:"format"`
	AuthorFacet                  []string `json:"author_facet"`
	LanguageFacet                []string `json:"language_facet"`
	BuildingFacet                []string `json:"building_facet"`
	LocationCodeT                []string `json:"location_code_t"`
	RegionFacet                  []string `json:"region_facet"`
	TopicFacet                   []string `json:"topic_facet"`
	SubjectsT                    []string `json:"subject_t"`
	CallNumbers                  []string `json:"callnumber_t"`
	Text                         []string `json:"text"`
	MarcDisplay                  []string `json:"marc_display"`
	BookplateCodeFacet           []string `json:"bookplate_code_facet"`
	BookplateCodeSS              []string `json:"bookplate_code_ss"`
	SourceInstitution            string   `json:"source_institution_s"`
}
