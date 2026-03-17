package bun

import (
	"github.com/google/wire"
	"github.com/lynx-go/lynx-clean-template/internal/infra/bun/bunrepo"
)

// ProviderSet is the Wire provider set for Bun ORM
var ProviderSet = wire.NewSet(
	bunrepo.NewGroupsRepo,
	bunrepo.NewGroupMembersRepo,
	bunrepo.NewUsersRepo,
	bunrepo.NewRefreshTokensRepo,
	bunrepo.NewFilesRepo,
)
