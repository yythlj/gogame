package room

import (
	"context"
	"game/src/common"
	"github.com/topfreegames/pitaya/v2/util"
)

func InGroup(ctx context.Context, groupName string) bool {
	s := common.GetDefaultApp().GetSessionFromCtx(ctx)
	if s.UID() == "" {
		return false
	}
	uids, _ := common.GetDefaultApp().GroupMembers(ctx, groupName)
	return util.SliceContainsString(uids, s.UID())
}
