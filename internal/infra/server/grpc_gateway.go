package server

import (
	"context"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/lynx-go/grpcapi/gateway"
	"github.com/lynx-go/lynx"
	apipb "github.com/lynx-go/lynx-clean-template/genproto/api/v1"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	lynxhttp "github.com/lynx-go/lynx/server/http"
	"google.golang.org/grpc"
)

type GRPCGatewayServer struct {
	*lynxhttp.Server
}

func NewGRPCGatewayServer(
	app lynx.App,
	config *config.AppConfig,
) (*GRPCGatewayServer, error) {
	c := config.GetServer().GetHttp()
	addr := c.Addr
	timeout := parseTimeout(c.Timeout)

	grpcAddr := config.GetServer().GetGrpc().Addr

	ctx := app.Context()

	// gateway 装配换库（grpcapi）：NewMux = 统一错误处理器
	//（ErrorResponseBuilder 承载模板错误契约）+ 入/出站 header matcher
	//（入站 authorization 恒拒绝防双写 401，出站 set-cookie 直透）+ 三个
	// marshaler 槽位统一 gateway.NewMarshaler（UseProtoNames + DiscardUnknown）。
	mux := gateway.NewMux(gateway.MuxOptions{
		ErrorHandler: gateway.NewErrorHandler(ErrorResponseBuilder{}, gateway.HTTPOptions{}),
	})

	// 拨号换库：gateway.Dial = 端点归一（原 localhost:port 拼接）+ insecure
	// 凭证 + 惰性 grpc.NewClient；整条 gateway 共享一条连接（原先每个
	// Register*HandlerFromEndpoint 各自建连），转发收包上限补齐为 8MiB。
	conn, err := gateway.Dial(ctx, grpcAddr, gateway.DialConfig{})
	if err != nil {
		return nil, err
	}

	register := []gateway.RegisterFunc{
		registerClient(apipb.NewAuthServiceClient, apipb.RegisterAuthServiceHandlerClient),
		registerClient(apipb.NewGroupsServiceClient, apipb.RegisterGroupsServiceHandlerClient),
		registerClient(apipb.NewUsersServiceClient, apipb.RegisterUsersServiceHandlerClient),
	}
	if err := gateway.Register(ctx, mux, conn, register...); err != nil {
		return nil, err
	}

	// Apply CORS middleware if configured
	var handler http.Handler = mux
	if corsConfig := c.GetCors(); corsConfig != nil {
		corsMiddleware := CORSMiddleware(corsConfig)
		handler = corsMiddleware(mux)
	}
	// request_id 透传/生成，注入日志 ctx（建议作为第一个中间件）
	handler = lynxhttp.WithRequestID()(handler)

	return &GRPCGatewayServer{lynxhttp.NewServer(handler, lynxhttp.WithAddr(addr), lynxhttp.WithTimeout(timeout))}, nil
}

// registerClient 以闭包适配 genproto 生成的 New*Client + Register*HandlerClient
// 对为 gateway.RegisterFunc（与原 Register*HandlerFromEndpoint 注册同源，
// 仅建连方式由"每注册一连接"收敛为共享连接）。
func registerClient[T any](
	newClient func(grpc.ClientConnInterface) T,
	register func(context.Context, *runtime.ServeMux, T) error,
) gateway.RegisterFunc {
	return func(ctx context.Context, mux *runtime.ServeMux, conn grpc.ClientConnInterface) error {
		return register(ctx, mux, newClient(conn))
	}
}

func parseTimeout(s string) time.Duration {
	if s == "" {
		s = "60s"
	}
	timeout, _ := time.ParseDuration(s)
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return timeout
}
