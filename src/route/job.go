package route

import (
	"game/src/common"
	"github.com/topfreegames/pitaya/v2/logger"
	"sync/atomic"
	"time"
)

type (
	JobRouter struct {
		chTask    chan *common.Functor //actor.mailbox
		chStop    chan struct{}
		Id        int
		typ       string
		ticker    *time.Ticker
		job       *common.Functor
		IsRunning atomic.Int32
	}
)

func newJobRouter(typ string, job *common.Functor, ticker *time.Ticker, routerId int) common.Router {
	return &JobRouter{
		Id:     routerId,
		typ:    typ,
		job:    job,
		chTask: make(chan *common.Functor),
		chStop: make(chan struct{}),
		ticker: ticker,
	}
}

func (r *JobRouter) AddTask(task *common.Functor) {
	if r.IsRunning.Load() != 1 {
		return
	}
	r.chTask <- task
}

func (r *JobRouter) Type() string {
	return r.typ
}

func (r *JobRouter) Close() {
	close(r.chStop)
}

func (r *JobRouter) Stop() {
	r.IsRunning.Store(0)
	r.ticker.Stop()
}

func (r *JobRouter) Handle() {
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
			case <-r.ticker.C: // execute cron task
				if r.IsRunning.Load() == 1 {
					r.job.Call()
				}
			default:
				break
			}
		}
	}
}
