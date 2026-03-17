package ulid

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/oklog/ulid/v2"
)

// ID represents a ULID identifier
type ID string

// IDGen generates ULIDs
type IDGen struct {
	entropy io.Reader
}

// NewIDGen creates a new ULID generator
func NewIDGen() *IDGen {
	// Use math/rand entropy source for simplicity
	entropy := ulid.Monotonic(ulid.DefaultEntropy(), 0)
	return &IDGen{
		entropy: entropy,
	}
}

// MustNextID generates a new ULID or panics
func (g *IDGen) MustNextID() ID {
	id, err := ulid.New(ulid.Timestamp(time.Now()), g.entropy)
	if err != nil {
		panic(err)
	}
	return ID(id.String())
}

// NextID generates a new ULID
func (g *IDGen) NextID() (ID, error) {
	id, err := ulid.New(ulid.Timestamp(time.Now()), g.entropy)
	return ID(id.String()), err
}

// ParseID parses a string to ULID ID
func ParseID(s string) ID {
	return ID(s)
}

// String returns the string representation
func (id ID) String() string {
	return string(id)
}

// IsValid checks if ID is valid (non-empty and valid ULID format)
func (id ID) IsValid() bool {
	if id == "" {
		return false
	}
	_, err := ulid.ParseStrict(string(id))
	return err == nil
}

// MarshalJSON implements json.Marshaler
func (id ID) MarshalJSON() ([]byte, error) {
	if id == "" {
		return []byte("null"), nil
	}
	return json.Marshal(string(id))
}

// UnmarshalJSON implements json.Unmarshaler
func (id *ID) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*id = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s != "" {
		if _, err := ulid.ParseStrict(s); err != nil {
			return fmt.Errorf("invalid ULID format: %w", err)
		}
	}
	*id = ID(s)
	return nil
}

// Value implements driver.Valuer for database compatibility
func (id ID) Value() (driver.Value, error) {
	if id == "" {
		return nil, nil
	}
	return string(id), nil
}

// Scan implements sql.Scanner for database compatibility
func (id *ID) Scan(value any) error {
	if value == nil {
		*id = ""
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*id = ID(v)
	case string:
		*id = ID(v)
	default:
		return errors.New("cannot scan non-string into ULID")
	}
	return nil
}

// EmptyID represents an empty ULID
func EmptyID() ID {
	return ID("")
}

// IDsToStrings converts a slice of ID to slice of string
func IDsToStrings(ids []ID) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}
	return result
}

// NewID generates a new ULID
func NewID() ID {
	gen := NewIDGen()
	return gen.MustNextID()
}
