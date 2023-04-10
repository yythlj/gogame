package common

import (
	"context"
	"github.com/topfreegames/pitaya/v2/constants"
	"github.com/topfreegames/pitaya/v2/logger"
	"github.com/topfreegames/pitaya/v2/session"
)

func IsOnline(ctx context.Context) bool {
	sessionI := ctx.Value(constants.SessionCtxKey)

	if sessionI == nil {
		logger.Log.Infof("session nil ")
		return false
	}
	sessionVal := sessionI.(session.Session)
	if sessionVal.UID() == "" {
		logger.Log.Infof("session uid nil ")
		return false
	}
	return true
}
