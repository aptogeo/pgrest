package pgrest

// Page structure
type Page struct {
	Slice  interface{}
	Offset uint64
	Limit  uint64
	Count  uint64
}
