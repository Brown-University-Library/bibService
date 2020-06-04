package sierra

type UniformTitle struct {
	Display string `json:"display"`
	Query   string `json:"query"`
}

type UniformTitles struct {
	Title []UniformTitle `json:"title"`
}

type UniformRelatedWorks struct {
	Author string         `json:"author"`
	Titles []UniformTitle `json:"title"`
}
