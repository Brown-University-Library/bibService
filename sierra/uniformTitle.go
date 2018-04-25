package sierra

// import (
// 	"encoding/json"
// 	"strings"
// )

type UniformTitle struct {
	Display string `json:"display"`
	Query   string `json:"query"`
}

type UniformTitles struct {
	Title []UniformTitle `json:"title"`
}

// func titleValue(value string) string {
// 	value = strings.TrimSpace(value)
// 	if !strings.HasSuffix(value, ".") {
// 		return value + "."
// 	}
// 	return value
// }

// func getUniformTitles(bib Bib, specStr string) []UniformTitles {
// 	// TODO: refactor this code, the current implementation is
// 	// ugly as hell.
// 	titlesArray := []UniformTitles{}
// 	spec, _ := NewFieldSpec(specStr)
// 	fields := bib.getFields(spec.MarcTag)
// 	for _, field := range fields {
// 		query := ""
// 		fieldTitles := UniformTitles{}
// 		for _, value := range field.getSubfieldsValues(spec.Subfields) {
// 			query = strings.TrimSpace(query + " " + titleValue(value))
// 			title := UniformTitle{Display: titleValue(value), Query: query}
// 			fieldTitles.Title = append(fieldTitles.Title, title)
// 		}
// 		titlesArray = append(titlesArray, fieldTitles)
// 	}
//
// 	vernValues := bib.VernacularValues(specStr)
// 	if len(vernValues) > 0 {
// 		query := ""
// 		titleWrapper := UniformTitles{}
// 		for _, value := range vernValues {
// 			query = strings.TrimSpace(query + " " + titleValue(value))
// 			title := UniformTitle{Display: titleValue(value), Query: query}
// 			titleWrapper.Title = append(titleWrapper.Title, title)
// 		}
// 		titlesArray = append(titlesArray, titleWrapper)
// 	}
// 	return titlesArray
// }
//
// func NewUniformTitles(bib Bib, specStr string) []UniformTitles {
// 	return getUniformTitles(bib, specStr)
// }
//
// func NewUniformTitlesString(bib Bib, specStr string) (string, error) {
// 	titles := getUniformTitles(bib, specStr)
// 	bytes, err := json.Marshal(titles)
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(bytes), nil
//}
