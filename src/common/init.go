package common

import (
	"github.com/topfreegames/pitaya/v2"
)

var (
	DefaultApp *pitaya.App
)

func StaticInit(app *pitaya.App) {
	DefaultApp = app
}

func GetDefaultApp() *pitaya.App {
	return DefaultApp
}
