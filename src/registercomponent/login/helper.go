package login

import (
	"game/src/common"
	"game/src/registermodules/memdbctrl"
)

func LoginSessionCB(args common.CallArgs) (res interface{}, err error) {
	sessionID, accountname, pwdMd5 := (args)[0].(int64), (args)[1].(string), (args)[2].(string)
	dbData := (args)[3].(string)
	sessionVal := common.GetDefaultApp().GetSessionPool().GetSessionByID(sessionID)
	if sessionVal == nil {
		return nil, nil
	}
	if sessionVal.UID() != "" {
		return nil, nil
	}
	if pwdMd5 != dbData {
		sessionVal.Push("login.login", &LoginResponse{Code: -100, Result: "密码错误"})
		return nil, nil
	}
	uid := accountname
	sessionVal.Bind(nil, uid)
	sessionVal.Push("login.login", &LoginResponse{Result: "登录成功+" + uid})
	return nil, nil
}

func UpdatePWDCB(args common.CallArgs) (res interface{}, err error) {
	sessionID, _, pwdMd5 := (args)[0].(int64), (args)[1].(string), (args)[2].(string)
	sessionVal := common.GetDefaultApp().GetSessionPool().GetSessionByID(sessionID)
	if sessionVal == nil {
		return nil, nil
	}
	if sessionVal.UID() == "" {
		return nil, nil
	}
	memdbctrl.DBData = pwdMd5
	sessionVal.Push("login.login", &LoginResponse{Result: "修改dbdata +" + pwdMd5})
	return nil, nil
}
