package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lynx-go/lynx"
	"github.com/lynx-go/lynx-clean-template/internal/app"
	"github.com/lynx-go/lynx-clean-template/internal/domain/users/repo"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/lynx-clean-template/pkg/pubsub"
	"github.com/lynx-go/lynx/contrib/zap"
	"github.com/lynx-go/lynx/pkg/errors"
	"github.com/lynx-go/x/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
	app lynx.Lynx,
	pubSub pubsub.Publisher,
	components []lynx.Component,
	componentBuilders []lynx.ComponentBuilder,
	componentBuilderSetFunc lynx.ComponentBuilderSetFunc,
	userRepo repo.UsersRepo,
	onStarts lynx.OnStartHooks,
	onStops lynx.OnStopHooks,
) *CLIContext {
	return &CLIContext{
		App:                     app,
		PubSub:                  pubSub,
		Components:              components,
		ComponentBuilders:       componentBuilders,
		ComponentBuilderSetFunc: componentBuilderSetFunc,
		OnStarts:                onStarts,
		OnStops:                 onStops,
		UserRepo:                userRepo,
	}
}

type CLIContext struct {
	App                     lynx.Lynx
	PubSub                  pubsub.Publisher
	Account                 *app.Account
	Components              []lynx.Component
	ComponentBuilders       []lynx.ComponentBuilder
	ComponentBuilderSetFunc lynx.ComponentBuilderSetFunc
	OnStarts                lynx.OnStartHooks
	OnStops                 lynx.OnStopHooks
	UserRepo                repo.UsersRepo
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

// WithPostWaitTime 执行结束后等待的时间，等待其他 Component 关闭
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

func newFileLogger(app lynx.Lynx) *slog.Logger {
	logLevel := app.Config().GetString("cli.log-level")
	if logLevel == "" {
		logLevel = "info"
	}
	logFile := app.Config().GetString("cli.log-file")
	if logFile == "" {
		logFile = "cli.log"
	}
	zlogger, err := zap.NewZapLoggerToFile(logLevel, logFile)
	errors.Fatal(err)
	slogger, err := zap.NewSLogger(zlogger, logLevel)
	errors.Fatal(err)
	return slogger
}

func runCLI(cmd *cobra.Command, args []string, fn func(ctx context.Context, cc *CLIContext, args *CLIArgs) error, opts ...CLIOption) {
	buildCLI(cmd, args, fn, opts...).Run()
}
func buildCLI(cmd *cobra.Command, args []string, fn func(ctx context.Context, cc *CLIContext, args *CLIArgs) error, opts ...CLIOption) *lynx.CLI {
	return lynx.New(newOptionsFromCmd(cmd), func(ctx context.Context, lx lynx.Lynx) error {
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
			lx.SetLogger(zap.MustNewLogger(lx))
		}
		cc, cleanup, err := wireCLIContext(lx)
		if err != nil {
			return err
		}
		if err := lx.Hooks(lynx.OnStop(func(ctx context.Context) error {
			cleanup()
			return nil
		})); err != nil {
			return err
		}

		if err := lx.Hooks(
			lynx.OnStart(cc.OnStarts...),
			lynx.OnStop(cc.OnStops...),
			lynx.Components(cc.Components...),
			lynx.ComponentBuilders(cc.ComponentBuilders...),
			lynx.ComponentBuilders(cc.ComponentBuilderSetFunc()...),
		); err != nil {
			return err
		}

		return lx.CLI(func(ctx context.Context) error {
			if o.PreWaitTime > 0 {
				log.InfoContext(ctx, fmt.Sprintf("waiting %s for components startup", o.PreWaitTime.String()))
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
				log.InfoContext(ctx, fmt.Sprintf("waiting %s for components shutdown", o.PostWaitTime.String()))
				time.Sleep(o.PostWaitTime)
			}
			return nil
		})
	})
}

func newOptionsFromCmd(cmd *cobra.Command) *lynx.Options {
	return lynx.NewOptions(
		lynx.WithName(cmd.Root().Name()+":"+cmd.Name()),
		lynx.WithBindConfigFunc(func(f *pflag.FlagSet, v *viper.Viper) error {
			if cd, _ := cmd.Root().PersistentFlags().GetString("config-dir"); cd != "" {
				return config.ConfigureViper(f, v, cd)
			}

			return config.ConfigureViper(f, v)
		}),
	)
}
