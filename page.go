package pgrest

// Page structure
type Page struct {
	Slice  interface{} `json:"slice"`
	Offset int         `json:"offset"`
	Limit  int         `json:"limit"`
	Count  int         `json:"count"`
}

// NewPage constructs Page
func NewPage(slice interface{}, count int, restQuery *RestQuery) *Page {
	p := new(Page)
	p.Slice = slice
	p.Offset = restQuery.Offset
	p.Limit = restQuery.Limit
	p.Count = count
	return p
}
