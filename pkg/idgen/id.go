package idgen

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"

	"github.com/lynx-go/lynx-clean-template/pkg/idgen/bigid"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen/ulid"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen/uuid"
	"github.com/samber/lo"
)

func BigID() ID {
	id := bigIdGen.MustNextID()
	return ID(id.String())
}

var bigIdGen = bigid.NewIDGen()

type ID string

func (id ID) String() string {
	return strings.TrimSpace(string(id))
}

func UUID() ID {
	return ID(uuid.NewString())
}

func ULID() ID {
	id := ulid.NewID()
	return ID(id.String())
}

//goland:noinspection GoMixedReceiverTypes
func (id ID) IsValid() bool {
	return id.String() != ""
}

func IDsToStrings(ids []ID) []string {
	return lo.Map(ids, func(id ID, _ int) string {
		return id.String()
	})
}

// MarshalJSON implements json.Marshaler
//
//goland:noinspection GoMixedReceiverTypes
func (id ID) MarshalJSON() ([]byte, error) {
	if id == "" {
		return []byte("null"), nil
	}
	return json.Marshal(string(id))
}

// UnmarshalJSON implements json.Unmarshaler
//
//goland:noinspection GoMixedReceiverTypes
func (id *ID) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*id = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
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
//
//goland:noinspection GoMixedReceiverTypes
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
		return errors.New("cannot scan non-string into ID")
	}
	return nil
}
