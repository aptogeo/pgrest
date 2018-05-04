package pgrest

// Action type
type Action int

const (
	// None action
	None Action = 1 << iota
	// Get action
	Get Action = 1 << iota
	// Post action
	Post Action = 1 << iota
	// Put action
	Put Action = 1 << iota
	// Patch action
	Patch Action = 1 << iota
	// Delete action
	Delete Action = 1 << iota
	// All actions
	All Action = Get + Post + Put + Patch + Delete
)
