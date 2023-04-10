package src

import (
	"game/src/common"
	"game/src/registercomponent"
	"game/src/registermodules"
	"game/src/route"
	"github.com/topfreegames/pitaya/v2"
)

func RegisterApp(app *pitaya.App) {
	common.StaticInit(app)
	route.RegisterRouteMgr()
	registermodules.RegisterApp()
	registercomponent.RegisterApp()
}
