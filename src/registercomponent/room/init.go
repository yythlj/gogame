package room

import (
	"context"
	"game/src/common"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/component"
	"github.com/topfreegames/pitaya/v2/logger"
	"github.com/topfreegames/pitaya/v2/timer"
	"time"
)

type (
	Room struct {
		component.Base
		timer *timer.Timer
	}
)

func (r *Room) AfterInit() {
	r.timer = pitaya.NewTimer(time.Minute, func() {
		count, err := common.GetDefaultApp().GroupCountMembers(context.Background(), "room")
		logger.Log.Debugf("UserCount: Time=> %s, Count=> %d, Error=> %q", time.Now().String(), count, err)
	})
}

func (c *Room) Init() {
	logger.Log.Infof("room component init \n")
	err := common.DefaultApp.GroupCreate(context.Background(), "room")
	if err != nil {
		panic(err)
	}
}

func NewComponent() *Room {
	return &Room{}
}
