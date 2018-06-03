package pgrest

// Action type
type Action int

const (
	// Get action
	Get Action = 1 << iota
	// Post action
	Post
	// Put action
	Put
	// Patch action
	Patch
	// Delete action
	Delete
)

// All actions
const All Action = Get + Post + Put + Patch + Delete

// None action
const None Action = 0

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
