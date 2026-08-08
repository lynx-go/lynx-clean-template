package main

import (
	"context"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/lynx-go/lynx"
	config "github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/lynx-clean-template/pkg/logger"
	"github.com/lynx-go/lynx-clean-template/pkg/timeutil"
	"github.com/spf13/pflag"
)

var (
	version string
)

func main() {
	builder := lynx.NewRunner(func(app lynx.App) error {
		app.SetLogger(logger.NewZap(app))

		boot, cleanup, err := wireBootstrap(app)
		if err != nil {
			log.Fatal(err)
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
		lynx.WithSetFlagsFunc(func(f *pflag.FlagSet) {
			f.String("config-dir", "./configs", "config file path")
			// 默认值为空而非 "info"：BindPFlags 会把未显式传入的 flag
			// 默认值绑进配置，若默认 "info" 会短路配置文件里的
			// logging.level/log_level（lynx v1.0.0 行为）。
			f.String("log-level", "", "log level, default info")
		}),
		lynx.WithBindConfigFunc(config.NewBindConfigFunc()),
		lynx.WithStopTimeout(30*time.Second),
		// 关停排水：readiness 先失败（LB 摘流），窗口结束后才真正关停
		lynx.WithDrainTimeout(5*time.Second),
	)
	builder.Run()
}

func init() {
	// Load .env file if present (dev convenience; silently ignored in production).
	_ = godotenv.Load()

	timeutil.InitCarbon()
}
