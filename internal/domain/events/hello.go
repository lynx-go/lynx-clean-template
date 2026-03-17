package events

import (
	"time"

	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
)

type HelloEvent struct {
	User    idgen.ID  `json:"user"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}
