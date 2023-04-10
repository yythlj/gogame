package route

import (
	"game/src/common"
	"github.com/topfreegames/pitaya/v2/logger"
	"github.com/topfreegames/pitaya/v2/service"
	"sync/atomic"
)

type (
	ServiceRouter struct {
		chMsg     chan *service.UnHandleMsg //use point keep reflect
		chTask    chan *common.Functor      //actor.mailbox
		chStop    chan struct{}
		Id        int
		IdSeque   int
		typ       string
		IsRunning atomic.Int32
	}
)

func newServiceRouter(typ string, id int, idSed int) common.Router {
	return &ServiceRouter{
		Id:      id,
		IdSeque: idSed,
		chMsg:   make(chan *service.UnHandleMsg),
		chTask:  make(chan *common.Functor),
		chStop:  make(chan struct{}),
		typ:     typ,
	}
}

func (r *ServiceRouter) AddTask(task *common.Functor) {
	if r.IsRunning.Load() != 1 {
		return
	}
	r.chTask <- task
}

func (r *ServiceRouter) Type() string {
	return r.typ
}

func (r *ServiceRouter) Close() {
	close(r.chStop)
}

func (r *ServiceRouter) Stop() {
	r.IsRunning.Store(0)
}

func (r *ServiceRouter) Handle() {
	defer func() {
		logger.Log.Infof("router close!!!! %v.%v", r.Type(), r.Id)
		close(r.chMsg)
		close(r.chTask)
	}()
	r.IsRunning.Store(1)
	goid := common.RouterMgrInstance.BindRouter(r.Id)
	logger.Log.Infof("router start %v.%v goid:%v", r.Type(), r.Id, goid)
	handlerService := common.GetDefaultApp().GetHandleService()
	for {
		select {
		case <-r.chStop:
			r.Stop()
			return
		default:
			select {
			case task := <-r.chTask:
				task.Call()
			case rm := <-r.chMsg:
				if r.IsRunning.Load() == 1 {
					if rm.Typ() == "local" {
						handlerService.ProcessLocalMsg(rm)
					} else {
						handlerService.ProcessRemoteMsg(rm)
					}
				}
			default:
				break
			}
		}
	}
}

func (r *ServiceRouter) AddMsg(rmsg *service.UnHandleMsg) {
	r.chMsg <- rmsg
}

func (r *ServiceRouter) ValidRoute(rmsg *service.UnHandleMsg) bool {
	route := rmsg.Router().Short()
	if val, ok := common.LoginService[route]; !ok || val == 0 {
		ctx := rmsg.Ctx()
		if !common.IsOnline(ctx) {
			return false
		}
	}
	seq := rmsg.Agent().GetSession().ID()%int64(r.IdSeque) + 1
	return seq == int64(r.Id)
}
