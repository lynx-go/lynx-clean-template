package contexts

import (
	"context"

	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
)

func UserID(ctx context.Context) (idgen.ID, bool) {
	v := ctx.Value(ContextKeyCurrentUser)
	uid, ok := v.(string)
	if !ok || uid == "" {
		return "", false
	}
	return idgen.ID(uid), true
}

func GroupIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ContextKeyCurrentGroup)
	groupID, ok := v.(string)
	if ok && groupID != "" {
		return groupID, true
	}
	return "", false
}
