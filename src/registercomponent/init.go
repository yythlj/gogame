package registercomponent

import (
	"game/src/common"
	"game/src/registercomponent/login"
	"game/src/registercomponent/room"
	"github.com/topfreegames/pitaya/v2/component"
	"strings"
)

func RegisterApp() {
	var coms = ComponentInit()
	for name, com := range coms {
		common.GetDefaultApp().Register(com,
			component.WithName(name),
			component.WithNameFunc(strings.ToLower),
		)
	}
}

func ComponentInit() (res map[string]component.Component) {
	res = make(map[string]component.Component)
	res["login"] = login.NewComponent()
	res["room"] = room.NewComponent()
	return
}
