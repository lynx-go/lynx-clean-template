package main

import (
	"context"
	"log"
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

	o := lynx.NewOptions(
		lynx.WithName("skyline-api"),
		lynx.WithVersion(version),
		lynx.WithSetFlagsFunc(func(f *pflag.FlagSet) {
			f.String("config-dir", "./configs", "config file path")
			f.String("log-level", "info", "log level, default info")
		}),
		lynx.WithBindConfigFunc(config.NewBindConfigFunc()),
		lynx.WithCloseTimeout(30*time.Second),
	)

	app := lynx.New(o, func(ctx context.Context, app lynx.Lynx) error {
		app.SetLogger(zap.MustNewLogger(app))

		boot, cleanup, err := wireBootstrap(app)
		if err != nil {
			log.Fatal(err)
		}
		if err := app.Hooks(lynx.OnStop(func(ctx context.Context) error {
			cleanup()
			return nil
		})); err != nil {
			return err
		}
		return boot.Bind(app)
	})
	app.Run()
}

func init() {
	// Load .env file if present (dev convenience; silently ignored in production).
	_ = godotenv.Load()

	timeutil.InitCarbon()
}
