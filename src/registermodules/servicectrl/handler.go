package servicectrl

import (
	"game/src/common"
	"github.com/topfreegames/pitaya/v2/logger"
	"github.com/topfreegames/pitaya/v2/service"
)

func (l *ServiceCtrl) Dispatch(concurrency int) {
	for i := 0; i < concurrency; i++ {
		router := common.RouterMgrInstance.AddServiceRouter("service", i+1, concurrency)
		go router.Handle()
	}
}

func (l *ServiceCtrl) Handle(msg *service.UnHandleMsg) {
	err := common.RouterMgrInstance.RecieveService("service", msg)
	if err != nil {
		logger.Log.Infof("err msg lose!!!! %v %v", msg.Router().Short(), err)
	}
}

func (l *ServiceCtrl) NewTask(functor *common.Functor, routerID int) {
	common.RouterMgrInstance.AddTask(functor, routerID)
}
