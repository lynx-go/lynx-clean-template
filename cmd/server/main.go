package main

import (
	"context"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/lynx-go/lynx"
	config "github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/lynx-clean-template/pkg/timeutil"
	"github.com/lynx-go/lynx/contrib/zap"
	"github.com/spf13/pflag"
)

var (
	version string
)

func main() {
	runner := lynx.NewRunner(func(app lynx.App) error {
		app.SetLogger(zap.MustNewLogger(app))

		boot, cleanup, err := wireBootstrap(app)
		if err != nil {
			return err
		}
		app.OnStop(func(ctx context.Context) error {
			cleanup()
			return nil
		})
		boot.Bind(app)
		return nil
	},
		lynx.WithName("lynx-api"),
		lynx.WithVersion(version),
		lynx.WithBindFlagsFunc(func(f *pflag.FlagSet) {
			f.String("config-dir", "./configs", "config file path")
			f.String("log-level", "info", "log level, default info")
		}),
		lynx.WithBindConfigFunc(config.NewBindConfigFunc()),
		lynx.WithShutdownTimeout(30*time.Second),
	)
	runner.Run()
}

func init() {
	// Load .env file if present (dev convenience; silently ignored in production).
	_ = godotenv.Load()

	timeutil.InitCarbon()
}
