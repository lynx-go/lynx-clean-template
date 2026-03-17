package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/lynx-go/lynx"
	apipb "github.com/lynx-go/lynx-clean-template/genproto/api/v1"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	lynxhttp "github.com/lynx-go/lynx/server/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func registerGrpcEndpoints(ctx context.Context, mux *runtime.ServeMux, addr string, opts []grpc.DialOption, registerFuncs ...func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error) error {
	for _, fn := range registerFuncs {
		if fn != nil {
			if err := fn(ctx, mux, addr, opts); err != nil {
				return err

			}
		}
	}
	return nil
}

type GRPCGatewayServer struct {
	*lynxhttp.Server
}

func NewGRPCGatewayServer(
	app lynx.Lynx,
	config *config.AppConfig,
) (*GRPCGatewayServer, error) {
	c := config.GetServer().GetHttp()
	addr := c.Addr
	timeout := parseTimeout(c.Timeout)

	grpcAddr := config.GetServer().GetGrpc().Addr
	grpcEndpoint := fmt.Sprintf("localhost:%s", getPort(grpcAddr))

	ctx := app.Context()

	// Create mux with custom error handler and marshaler
	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(HTTPErrorHandler),
		runtime.WithMarshalerOption("*", NewCustomMarshaler()),
		runtime.WithMarshalerOption("*/*", NewCustomMarshaler()),
		runtime.WithMarshalerOption("application/json", NewCustomMarshaler()),
		runtime.WithMarshalerOption("application/json; charset=utf-8", NewCustomMarshaler()),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := registerGrpcEndpoints(
		ctx,
		mux,
		grpcEndpoint,
		opts,
		apipb.RegisterAuthServiceHandlerFromEndpoint,
		apipb.RegisterGroupsServiceHandlerFromEndpoint,
		apipb.RegisterUsersServiceHandlerFromEndpoint,
	); err != nil {
		return nil, err
	}

	// Apply CORS middleware if configured
	var handler http.Handler = mux
	if corsConfig := c.GetCors(); corsConfig != nil {
		corsMiddleware := CORSMiddleware(corsConfig)
		handler = corsMiddleware(mux)
	}

	return &GRPCGatewayServer{lynxhttp.NewServer(handler, lynxhttp.WithAddr(addr), lynxhttp.WithTimeout(timeout))}, nil
}

func getPort(addr string) string {
	arr := strings.Split(addr, ":")
	return arr[len(arr)-1]
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
