package login

import (
	"context"
	"game/src/common"
	"game/src/registermodules/memdbctrl"
	"strconv"
)

type (
	LoginMessage struct {
		AccountName string `json:"account_name"`
		PwdMd5      string `json:"pwd_md5"`
	}

	LoginResponse struct {
		Code   int    `json:"code"`
		Result string `json:"result"`
	}
)

func (l *Login) Login(ctx context.Context, msg *LoginMessage) (*LoginResponse, error) {
	sessionVal := common.GetDefaultApp().GetSessionFromCtx(ctx)
	if sessionVal.UID() != "" {
		return &LoginResponse{-99, "login session already"}, nil
	}
	sessionID := sessionVal.ID()
	var functor = &common.Functor{
		CallFunc: LoginSessionCB,
	}
	functor.AddArgs(sessionID, msg.AccountName, msg.PwdMd5)
	router := common.RouterMgrInstance.CurRouterID()
	memDBCtrlV := memdbctrl.GetModuleInstance()
	memDBCtrlV.LoadData(msg.AccountName, functor)
	return &LoginResponse{1, "wait for " + strconv.Itoa(router)}, nil
}

func (l *Login) UpdatePWD(ctx context.Context, msg *LoginMessage) (*LoginResponse, error) {
	sessionVal := common.GetDefaultApp().GetSessionFromCtx(ctx)
	if sessionVal.UID() == "" {
		return &LoginResponse{-99, "no login session "}, nil
	}
	sessionID := sessionVal.ID()
	var functor = &common.Functor{
		CallFunc: UpdatePWDCB,
	}
	functor.AddArgs(sessionID, msg.AccountName, msg.PwdMd5)
	router := common.RouterMgrInstance.CurRouterID()
	memDBCtrlV := memdbctrl.GetModuleInstance()
	memDBCtrlV.LoadData(msg.AccountName, functor)
	return &LoginResponse{1, "wait for " + strconv.Itoa(router)}, nil
}
