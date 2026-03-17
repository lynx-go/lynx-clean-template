package tests

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lynx-go/lynx"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/lynx/contrib/zap"
	"github.com/lynx-go/lynx/pkg/errors"
)

type TestingSuite struct {
	App lynx.Lynx
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

func newFileLogger(app lynx.Lynx) *slog.Logger {
	logLevel := app.Config().GetString("test.log-level")
	if logLevel == "" {
		logLevel = "info"
	}
	logFile := app.Config().GetString("test.log-file")
	if logFile == "" {
		logFile = "test.log"
	}
	zlogger, err := zap.NewZapLoggerToFile(logLevel, logFile)
	errors.Fatal(err)
	slogger, err := zap.NewSLogger(zlogger, logLevel)
	errors.Fatal(err)
	return slogger
}

// RunTestSuite 初始化并运行测试套件
func RunTestSuite(fn func(ctx context.Context, ts *TestingSuite) error, opts ...TestOption) {
	buildTestSuite(fn, opts...).Run()
}

func buildTestSuite(fn func(ctx context.Context, ts *TestingSuite) error, opts ...TestOption) *lynx.CLI {
	return lynx.New(newTestOptions(), func(ctx context.Context, lx lynx.Lynx) error {
		o := &TestOptions{
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

		ts, cleanup, err := wireTestingSuite(lx)
		if err != nil {
			return err
		}
		ts.App = lx

		if err := lx.Hooks(lynx.OnStop(func(ctx context.Context) error {
			cleanup()
			return nil
		})); err != nil {
			return err
		}

		return lx.CLI(func(ctx context.Context) error {
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
	})
}

func newTestOptions() *lynx.Options {
	return lynx.NewOptions(
		lynx.WithName("lynx:test"),
		lynx.WithBindConfigFunc(config.NewBindConfigFunc("./configs", "../configs", "../../configs")),
	)
}
