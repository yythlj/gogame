package login

import (
	"github.com/topfreegames/pitaya/v2/component"
	"github.com/topfreegames/pitaya/v2/logger"
)

type (
	Login struct {
		component.Base
	}
)

func (l *Login) Init() {
	logger.Log.Infof("login component init \n")
}

func NewComponent() *Login {
	return &Login{}
}
