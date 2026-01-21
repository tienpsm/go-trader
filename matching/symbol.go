package matching

import (
	"fmt"
	"strings"
)

// Symbol represents a trading symbol
type Symbol struct {
	// ID is the unique identifier for the symbol
	ID uint32
	// Name is the symbol name (max 8 characters)
	Name string
}

// NewSymbol creates a new Symbol
func NewSymbol(id uint32, name string) Symbol {
	// Truncate name to 8 characters if necessary
	if len(name) > 8 {
		name = name[:8]
	}
	return Symbol{
		ID:   id,
		Name: strings.TrimSpace(name),
	}
}

// String returns the string representation of a Symbol
func (s Symbol) String() string {
	return fmt.Sprintf("Symbol(ID=%d, Name=%s)", s.ID, s.Name)
}
