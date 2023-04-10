package route

import (
	"game/src/common"
	"github.com/topfreegames/pitaya/v2/logger"
	"github.com/topfreegames/pitaya/v2/timer"
	"sync/atomic"
)

type (
	SchedRouter struct {
		chTask    chan *common.Functor //actor.mailbox
		chStop    chan struct{}
		Id        int
		typ       string
		IsRunning atomic.Int32
	}
)

func newSchedRouter(typ string, id int) common.Router {
	return &SchedRouter{
		Id:     id,
		chTask: make(chan *common.Functor),
		chStop: make(chan struct{}),
		typ:    typ,
	}
}

func (r *SchedRouter) AddTask(task *common.Functor) {
	if r.IsRunning.Load() != 1 {
		return
	}
	r.chTask <- task
}

func (r *SchedRouter) Type() string {
	return r.typ
}

func (r *SchedRouter) Close() {
	close(r.chStop)
}

func (r *SchedRouter) Stop() {
	r.IsRunning.Store(0)
	timer.GlobalTicker.Stop()
}

func (r *SchedRouter) Handle() {
	defer func() {
		logger.Log.Infof("router close!!!! %v.%v", r.Type(), r.Id)
		close(r.chTask)
	}()
	r.IsRunning.Store(1)
	goid := common.RouterMgrInstance.BindRouter(r.Id)
	logger.Log.Infof("router start %v.%v goid:%v", r.Type(), r.Id, goid)
	for {
		select {
		case <-r.chStop:
			r.Stop()
			return
		default:
			select {
			case task := <-r.chTask:
				task.Call()
			case <-timer.GlobalTicker.C: // execute cron task
				if r.IsRunning.Load() == 1 {
					timer.Cron()
				}
			case t := <-timer.Manager.ChCreatedTimer: // new Timers
				if r.IsRunning.Load() == 1 {
					timer.AddTimer(t)
				}
			case id := <-timer.Manager.ChClosingTimer: // closing Timers
				if r.IsRunning.Load() == 1 {
					timer.RemoveTimer(id)
				}
			default:
				break
			}
		}
	}
}
