package tests

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lynx-go/lynx"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/lynx/contrib/zap"
)

type TestingSuite struct {
	App lynx.App
}

func NewTestingSuite() *TestingSuite {
	return &TestingSuite{}
}

type TestOptions struct {
	PreWaitTime  time.Duration
	PostWaitTime time.Duration
	LogToFile    bool
}

type TestOption func(*TestOptions)

// WithPostWaitTime 执行结束后等待的时间，等待其他 Component 关闭
func WithPostWaitTime(waitTime time.Duration) TestOption {
	return func(o *TestOptions) {
		o.PostWaitTime = waitTime
	}
}

func WithLogToFile() TestOption {
	return func(o *TestOptions) {
		o.LogToFile = true
	}
}

func WithPreWaitTime(waitTime time.Duration) TestOption {
	return func(o *TestOptions) {
		o.PreWaitTime = waitTime
	}
}

func newFileLogger(app lynx.App) (*slog.Logger, error) {
	logLevel := app.Config().GetString("test.log-level")
	if logLevel == "" {
		logLevel = "info"
	}
	logFile := app.Config().GetString("test.log-file")
	if logFile == "" {
		logFile = "test.log"
	}
	zlogger, err := zap.NewZapLogger(logLevel, logFile)
	if err != nil {
		return nil, err
	}
	slogger, err := zap.NewSLogger(zlogger, logLevel)
	if err != nil {
		return nil, err
	}
	return slogger, nil
}

// RunTestSuite 初始化并运行测试套件
func RunTestSuite(fn func(ctx context.Context, ts *TestingSuite) error, opts ...TestOption) {
	buildTestSuite(fn, opts...).Run()
}

func buildTestSuite(fn func(ctx context.Context, ts *TestingSuite) error, opts ...TestOption) *lynx.Runner {
	return lynx.NewRunner(func(app lynx.App) error {
		o := &TestOptions{
			PreWaitTime:  10 * time.Millisecond,
			PostWaitTime: 10 * time.Millisecond,
			LogToFile:    false,
		}
		for _, opt := range opts {
			opt(o)
		}

		if o.LogToFile {
			l, err := newFileLogger(app)
			if err != nil {
				return err
			}
			app.SetLogger(l)
		} else {
			app.SetLogger(zap.MustNewLogger(app))
		}

		ts, cleanup, err := wireTestingSuite(app)
		if err != nil {
			return err
		}
		ts.App = app

		app.OnStop(func(ctx context.Context) error {
			cleanup()
			return nil
		})

		app.Command(func(ctx context.Context) error {
			if o.PreWaitTime > 0 {
				slog.InfoContext(ctx, fmt.Sprintf("waiting %s for components startup", o.PreWaitTime.String()))
				time.Sleep(o.PreWaitTime)
			}

			if err := fn(ctx, ts); err != nil {
				slog.ErrorContext(ctx, "test execution error", "error", err)
				return fmt.Errorf("test execution error: %v", err)
			}

			if o.PostWaitTime > 0 {
				slog.InfoContext(ctx, fmt.Sprintf("waiting %s for components shutdown", o.PostWaitTime.String()))
				time.Sleep(o.PostWaitTime)
			}

			return nil
		})
		return nil
	}, newTestOptions()...)
}

func newTestOptions() []lynx.Option {
	return []lynx.Option{
		lynx.WithName("lynx:test"),
		lynx.WithBindConfigFunc(config.NewBindConfigFunc("./configs", "../configs", "../../configs")),
	}
}
