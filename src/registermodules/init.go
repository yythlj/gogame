package registermodules

import (
	"game/src/common"
	"game/src/registermodules/memdbctrl"
	"game/src/registermodules/schedctrl"
	"game/src/registermodules/servicectrl"
	"github.com/topfreegames/pitaya/v2/interfaces"
	"github.com/topfreegames/pitaya/v2/service"
)

// servicectrl --- 玩家逻辑线程
// schedctrl --- 主逻辑 + 定时器线程

func RegisterApp() {
	var coms = ModuleInit()
	for name, com := range coms {
		common.GetDefaultApp().RegisterModule(com, name)
	}
	serviceCtrl := servicectrl.NewModule()
	service.MsgCtrlInstance = serviceCtrl
	common.GetDefaultApp().RegisterModule(serviceCtrl, "servicectrl")
}

func ModuleInit() (res map[string]interfaces.Module) {
	res = make(map[string]interfaces.Module)
	res["schedctrl"] = schedctrl.NewModule()
	res["memdbctrl"] = memdbctrl.NewModule()
	return res
}
