package pgrest

// Action type
type Action int

const (
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
	// None action
	None Action = 0
)

func (a Action) String() string {
	if a == Get {
		return "Get"
	} else if a == Post {
		return "Post"
	} else if a == Put {
		return "Put"
	} else if a == Patch {
		return "Patch"
	} else if a == Delete {
		return "Delete"
	} else {
		return "None"
	}
}
