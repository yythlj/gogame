package route

import (
	"game/src/common"
)

func RegisterRouteMgr() {
	common.RouterMgrInstance = newRouterMgr()
}
