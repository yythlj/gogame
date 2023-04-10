package memdbctrl

import (
	"game/src/common"
	"game/src/registermodules/servicectrl"
	"sync/atomic"
)

var (
	ModelBindTouter = atomic.Int32{}
	ModelRequire    = atomic.Int32{}
	DBData          = "123"
)

func (m *MemDBCtrl) LoadData(account string, loadFunc *common.Functor) {
	curRouter := common.RouterMgrInstance.CurRouterID()
	var curRouterI32 = int32(curRouter)
	holdRouter := ModelBindTouter.Load()
	if ModelRequire.Load() == 0 || holdRouter == curRouterI32 {
		//DBdata占据者线程，可以直接安全使用
		ModelBindTouter.Store(curRouterI32)
		ModelRequire.Add(1)
		loadFunc.CallWithAddArgs(DBData)
		ModelRequire.Add(-1)
	} else {
		//推送事务到DBdata占据者线程，安全调用
		functor := &common.Functor{
			CallFunc: LoadDataFunc,
		}
		functor.AddArgs(m, account, loadFunc)
		ModelRequire.Add(1)
		serviceCtrlV := servicectrl.GetModuleInstance()
		serviceCtrlV.NewTask(functor, int(holdRouter))
	}
}

func LoadDataFunc(args common.CallArgs) (res interface{}, err error) {
	m := args[0].(*MemDBCtrl)
	account := args[1].(string)
	loadFunc := args[2].(*common.Functor)
	ModelRequire.Add(-1)
	m.LoadData(account, loadFunc)
	return nil, nil
}
