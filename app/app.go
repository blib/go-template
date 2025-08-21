package app

import "go.uber.org/fx"

func InitApp() fx.Option {
	return fx.Module(
		"main app",
		fx.Provide(NewHealthz),
	)
}
