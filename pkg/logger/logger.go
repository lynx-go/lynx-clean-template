// Package logger 提供应用日志构造辅助：基于 zap 的 slog 实例并包装
// lynx/logging 的 trace/attrs handler，使请求级属性（request_id 等）
// 自动进入日志记录。
package logger

import (
	"log/slog"

	"github.com/lynx-go/lynx"
	"github.com/lynx-go/lynx/contrib/zap"
	"github.com/lynx-go/lynx/logging"
)

// NewZap 创建基于 zap 的 slog 实例（含服务标识字段），并包装
// attrs/trace handler：ctx 携带请求级属性（如 WithRequestID 注入的
// request_id）或有效 SpanContext 时，日志自动携带这些字段。
func NewZap(ctx lynx.AppContext) *slog.Logger {
	return slog.New(
		logging.NewAttrsHandler(
			logging.NewTraceHandler(
				zap.MustNewLogger(ctx).Handler(),
			),
		),
	)
}

// NewZapFile 基于 zap 创建输出到文件（或任意 zap output）的 slog 实例，
// 同样包装 attrs/trace handler。logLevel 为空时使用 info。
func NewZapFile(logLevel string, outputs ...string) (*slog.Logger, error) {
	if logLevel == "" {
		logLevel = "info"
	}
	zlogger, err := zap.NewZapLogger(logLevel, outputs...)
	if err != nil {
		return nil, err
	}
	slogger, err := zap.NewSLogger(zlogger, logLevel)
	if err != nil {
		return nil, err
	}
	return slog.New(
		logging.NewAttrsHandler(
			logging.NewTraceHandler(slogger.Handler()),
		),
	), nil
}
