package matching

// UpdateType represents the type of update in the order book
type UpdateType uint8

const (
	// UpdateNone indicates no update
	UpdateNone UpdateType = iota
	// UpdateAdd indicates an addition
	UpdateAdd
	// UpdateUpdate indicates a modification
	UpdateUpdate
	// UpdateDelete indicates a deletion
	UpdateDelete
)

// String returns the string representation of an UpdateType
func (u UpdateType) String() string {
	switch u {
	case UpdateNone:
		return "NONE"
	case UpdateAdd:
		return "ADD"
	case UpdateUpdate:
		return "UPDATE"
	case UpdateDelete:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}
