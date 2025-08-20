package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/blib/go-template/app"
	"github.com/blib/go-template/services"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	envFlag      = "env"
	logLevelFlag = "log-level"
)

func init() {
	rootCmd.AddCommand(serveCmd)

	addStringFlag(
		serveCmd.Flags(),
		envFlag,
		"dev",
		"Environment to run the application",
	)
	addStringFlag(
		serveCmd.Flags(),
		logLevelFlag,
		"info",
		"Log level",
	)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTPS server",
	Run: func(cmd *cobra.Command, _ []string) {

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		app := fx.New(
			fx.Provide(services.NewConfigProvider),
			fx.Provide(services.NewZapLogger),
			fx.Provide(services.NewHTTPServer),
			fx.Provide(app.NewHealthz),
			fx.Invoke(func(lc fx.Lifecycle, server *services.HTTPServer, logger *zap.Logger) {
				lc.Append(fx.Hook{
					OnStart: func(context.Context) error {
						go func() {
							if err := server.Start(); err != nil {
								logger.Error("Server failed to start", zap.Error(err))
								stop()
							}
						}()
						return nil
					},
					OnStop: func(ctx context.Context) error {
						shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
						defer cancel()
						return server.Stop(shutdownCtx)
					},
				})
			}),
		)

		if err := app.Start(ctx); err != nil {
			return
		}

		<-ctx.Done()

		stop()

		if err := app.Stop(context.Background()); err != nil {
			return
		}
	},
}
