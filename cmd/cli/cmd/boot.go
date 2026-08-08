package cmd

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/lynx-go/lynx"
	"github.com/lynx-go/lynx-clean-template/internal/app"
	"github.com/lynx-go/lynx-clean-template/internal/domain/users/repo"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/lynx-clean-template/pkg/logger"
	"github.com/lynx-go/lynx-clean-template/pkg/pubsub"
	"github.com/lynx-go/lynx/boot"
	xl "github.com/lynx-go/x/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type CLIArgs struct {
	Cmd  *cobra.Command
	Args []string
}

func (args *CLIArgs) GetString(key string) string {
	v, _ := args.Cmd.Flags().GetString(key)
	return v
}

func (args *CLIArgs) GetInt(key string) int {
	v, _ := args.Cmd.Flags().GetInt(key)
	return v
}

func (args *CLIArgs) GetBool(key string) bool {
	v, _ := args.Cmd.Flags().GetBool(key)
	return v
}

func NewCLIContext(
	app lynx.App,
	pubSub pubsub.Publisher,
	services []lynx.Service,
	serviceFactories []lynx.ServiceFactory,
	userRepo repo.UsersRepo,
	onStarts boot.OnStartHooks,
	onStops boot.OnStopHooks,
) *CLIContext {
	return &CLIContext{
		App:              app,
		PubSub:           pubSub,
		Services:         services,
		ServiceFactories: serviceFactories,
		OnStarts:         onStarts,
		OnStops:          onStops,
		UserRepo:         userRepo,
	}
}

type CLIContext struct {
	App              lynx.App
	PubSub           pubsub.Publisher
	Account          *app.Account
	Services         []lynx.Service
	ServiceFactories []lynx.ServiceFactory
	OnStarts         boot.OnStartHooks
	OnStops          boot.OnStopHooks
	UserRepo         repo.UsersRepo
}

func (cc *CLIContext) Println(v ...interface{}) {
	fmt.Println(v...)
}

func (cc *CLIContext) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

type cliOptions struct {
	PreWaitTime  time.Duration
	PostWaitTime time.Duration
	LogToFile    bool
}

type CLIOption func(*cliOptions)

// WithPostWaitTime 执行结束后等待的时间，等待其他 Service 关闭
func WithPostWaitTime(waitTime time.Duration) CLIOption {
	return func(o *cliOptions) {
		o.PostWaitTime = waitTime
	}
}

func WithLogToFile() CLIOption {
	return func(o *cliOptions) {
		o.LogToFile = true
	}
}

func WithPreWaitTime(waitTime time.Duration) CLIOption {
	return func(o *cliOptions) {
		o.PreWaitTime = waitTime
	}
}

func newFileLogger(app lynx.App) *slog.Logger {
	logLevel := app.Config().GetString("cli.log-level")
	if logLevel == "" {
		logLevel = "info"
	}
	logFile := app.Config().GetString("cli.log-file")
	if logFile == "" {
		logFile = "cli.log"
	}
	slogger, err := logger.NewZapFile(logLevel, logFile)
	if err != nil {
		log.Fatal(err)
	}
	return slogger
}

func runCLI(cmd *cobra.Command, args []string, fn func(ctx context.Context, cc *CLIContext, args *CLIArgs) error, opts ...CLIOption) {
	buildCLI(cmd, args, fn, opts...).Run()
}
func buildCLI(cmd *cobra.Command, args []string, fn func(ctx context.Context, cc *CLIContext, args *CLIArgs) error, opts ...CLIOption) *lynx.Runner {
	return lynx.NewRunner(func(lx lynx.App) error {
		o := &cliOptions{
			PreWaitTime:  10 * time.Millisecond,
			PostWaitTime: 10 * time.Millisecond,
			LogToFile:    false,
		}
		for _, opt := range opts {
			opt(o)
		}
		if o.LogToFile {
			lx.SetLogger(newFileLogger(lx))
		} else {
			lx.SetLogger(logger.NewZap(lx))
		}
		cc, cleanup, err := wireCLIContext(lx)
		if err != nil {
			return err
		}
		lx.OnStop(func(ctx context.Context) error {
			cleanup()
			return nil
		})

		lx.OnStart(cc.OnStarts...)
		lx.OnStop(cc.OnStops...)
		lx.Register(cc.Services...)
		lx.RegisterFactories(cc.ServiceFactories...)

		return lx.Command(func(ctx context.Context) error {
			if o.PreWaitTime > 0 {
				xl.InfoContext(ctx, fmt.Sprintf("waiting %s for services startup", o.PreWaitTime.String()))
				time.Sleep(o.PreWaitTime)
			}
			err, ok := lo.TryWithErrorValue(func() error {
				if err := fn(ctx, cc, &CLIArgs{Cmd: cmd, Args: args}); err != nil {
					return err
				}
				return nil
			})
			if !ok || err != nil {
				slog.ErrorContext(ctx, "cli execution error", "error", err)
				return fmt.Errorf("cli execution error %v", err)
			}
			if o.PostWaitTime > 0 {
				// wait pubsub completed
				xl.InfoContext(ctx, fmt.Sprintf("waiting %s for services shutdown", o.PostWaitTime.String()))
				time.Sleep(o.PostWaitTime)
			}
			return nil
		})
	}, newOptionsFromCmd(cmd)...)
}

func newOptionsFromCmd(cmd *cobra.Command) []lynx.Option {
	return []lynx.Option{
		lynx.WithName(cmd.Root().Name() + ":" + cmd.Name()),
		lynx.WithBindConfigFunc(func(f *pflag.FlagSet, c lynx.ConfigSource) error {
			if cd, _ := cmd.Root().PersistentFlags().GetString("config-dir"); cd != "" {
				return config.ConfigureViper(f, c, cd)
			}

			return config.ConfigureViper(f, c)
		}),
	}
}
