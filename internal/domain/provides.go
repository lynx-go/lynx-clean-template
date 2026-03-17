package domain

import (
	"github.com/google/wire"
	"github.com/lynx-go/lynx-clean-template/internal/domain/files"
	"github.com/lynx-go/lynx-clean-template/internal/domain/users"
)

var ProviderSet = wire.NewSet(
	users.NewUserService,
	files.NewService,
	files.NewURLRenderer,
)
