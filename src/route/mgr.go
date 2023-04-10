package route

import (
	"errors"
	"game/src/common"
	"github.com/timandy/routine"
	"github.com/topfreegames/pitaya/v2/logger"
	"github.com/topfreegames/pitaya/v2/service"
	"sync"
	"time"
)

// todo 歪歪梯 router热替换 热增加
type (
	GoidRouterMgr struct {
		RouterMap   map[int]common.Router
		routineBind sync.Map
	}
)

func newRouterMgr() common.RouterMgr {
	rMgr := &GoidRouterMgr{
		RouterMap:   map[int]common.Router{},
		routineBind: sync.Map{},
	}
	return rMgr
}

func (r *GoidRouterMgr) AddTask(task *common.Functor, routerID int) {
	router := r.RouterMap[routerID]
	router.AddTask(task)
}

func (r *GoidRouterMgr) Close() {
	defer func() {
		r.RouterMap = nil
	}()
	for _, router := range r.RouterMap {
		router.Close()
	}
}

func (r *GoidRouterMgr) CloseRouter(typ string) {
	for _, router := range r.ListRouter(typ) {
		router.Close()
	}
}

func (r *GoidRouterMgr) StopRouter(typ string) {
	for _, router := range r.ListRouter(typ) {
		router.Stop()
	}
}

func (r *GoidRouterMgr) BindRouter(routerId int) int64 {
	var goid = routine.Goid()
	r.routineBind.Store(goid, routerId)
	return goid
}

func (r *GoidRouterMgr) CurRouterID() int {
	var goid = routine.Goid()
	res, ok := r.routineBind.Load(goid)
	if !ok {
		logger.Log.Errorf("------------goid 0 %v", goid) //主线程退出时，强刷一遍db时，db的回调可能触发获取
		res = 0
	}
	return res.(int)
}

func (r *GoidRouterMgr) ListRouter(typ string) []common.Router {
	var res []common.Router
	for _, router := range r.RouterMap {
		if router.Type() == typ {
			res = append(res, router)
		}
	}
	return res
}

func (r *GoidRouterMgr) AddJobRouter(typ string, job *common.Functor, ticker *time.Ticker, routerID int) common.Router {
	router := newJobRouter(typ, job, ticker, routerID)
	r.RouterMap[routerID] = router
	return router
}
func (r *GoidRouterMgr) AddSchedRouter(typ string, routerID int) common.Router {
	router := newSchedRouter(typ, routerID)
	r.RouterMap[routerID] = router
	return router
}
func (r *GoidRouterMgr) AddServiceRouter(typ string, routerID, idSeq int) common.Router {
	router := newServiceRouter(typ, routerID, idSeq)
	r.RouterMap[routerID] = router
	return router
}

func (r *GoidRouterMgr) RecieveService(routeTyp string, msg *service.UnHandleMsg) error {
	for _, router := range r.ListRouter(routeTyp) {
		sRouter := router.(*ServiceRouter)
		if sRouter.ValidRoute(msg) {
			sRouter.AddMsg(msg)
			return nil
		}
	}
	return errors.New("no suitable route")
}
