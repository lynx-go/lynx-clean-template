package uuid

import (
	"github.com/google/uuid"
	"github.com/samber/lo"
)

func NewString() string {
	return lo.Must1(uuid.NewV7()).String()
}
