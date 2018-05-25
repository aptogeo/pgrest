package pgrest

// Page structure
type Page struct {
	slice  interface{}
	offset uint64
	limit  uint64
	count  uint64
}

// NewPage constructs Page
func NewPage(slice interface{}, count uint64, restQuery *RestQuery) *Page {
	p := new(Page)
	p.slice = slice
	p.offset = restQuery.Offset
	p.limit = restQuery.Limit
	p.count = count
	return p
}

// Slice returns slice
func (p *Page) Slice() interface{} {
	return p.slice
}

// Offset returns offset
func (p *Page) Offset() uint64 {
	return p.offset
}

// Limit returns limit
func (p *Page) Limit() uint64 {
	return p.limit
}

// Count returns count
func (p *Page) Count() uint64 {
	return p.count
}
