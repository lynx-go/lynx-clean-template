package vo

import (
	"regexp"
	"strings"

	"github.com/lynx-go/lynx-clean-template/pkg/errors"
)

var emailRegex = regexp.MustCompile(`^.+@.+\..+$`)

// Email is a validated, normalised email address value object.
type Email struct {
	value string
}

// NewEmail creates a validated Email value object.
// Returns an error if the address does not match a basic email pattern.
func NewEmail(address string) (Email, error) {
	address = strings.TrimSpace(strings.ToLower(address))
	if !emailRegex.MatchString(address) {
		return Email{}, errors.Cause("invalid email address")
	}
	return Email{value: address}, nil
}

// String returns the normalised email string.
func (e Email) String() string { return e.value }

// Equals returns true if both Email values are identical.
func (e Email) Equals(other Email) bool { return e.value == other.value }

// IsZero returns true for an uninitialised Email.
func (e Email) IsZero() bool { return e.value == "" }
