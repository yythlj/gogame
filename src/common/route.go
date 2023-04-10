package common

import (
	"github.com/topfreegames/pitaya/v2/service"
	"time"
)

type (
	Router interface {
		Close()
		Stop()
		Handle()
		AddTask(task *Functor)
		Type() string
	}
	RouterMgr interface {
		CurRouterID() int
		BindRouter(routerId int) int64
		ListRouter(typ string) []Router
		AddTask(task *Functor, routerID int)
		CloseRouter(typ string)
		StopRouter(typ string)
		AddJobRouter(typ string, job *Functor, ticker *time.Ticker, routerID int) Router
		AddSchedRouter(typ string, routerID int) Router
		AddServiceRouter(typ string, routerID, idSeq int) Router
		RecieveService(routeTyp string, msg *service.UnHandleMsg) error
	}
)

var (
	//routerid大于0为service线程
	FreeRouterId   = 0          //用于表示资源未被锁定
	SchedRouterIDs = [1]int{-1} //全局主逻辑+定时器 只开启一个

	LoginService = map[string]int{
		"login.login": 1,
	}
	RouterMgrInstance RouterMgr
)
