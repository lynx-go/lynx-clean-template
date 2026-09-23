package server

import (
	"github.com/lynx-go/grpcapi/authz"
	grpcapiinterceptor "github.com/lynx-go/grpcapi/interceptor"
	"github.com/lynx-go/lynx"
	apipb "github.com/lynx-go/lynx-clean-template/genproto/api/v1"
	grpcsvc "github.com/lynx-go/lynx-clean-template/internal/api/grpc"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/lynx-clean-template/pkg/grpc/interceptor"
	lynxgrpc "github.com/lynx-go/lynx/server/grpc"
)

func NewAuthValidator(
	config *config.AppConfig,
) interceptor.Validator {
	return interceptor.NewDefaultValidator(config.Security.Jwt.Secret)
}

func NewGRPCServer(
	app lynx.App,
	cfg *config.AppConfig,
	authValidator interceptor.Validator,
	policies *authz.PolicySet,
	auth *grpcsvc.AuthService,
	groups *grpcsvc.GroupsService,
	users *grpcsvc.UsersService,
) (*lynxgrpc.Server, error) {
	c := cfg.GetServer().GetGrpc()
	addr := c.Addr
	timeout := parseTimeout(c.Timeout)
	authInterceptor, err := interceptor.NewAuthInterceptor(authValidator, policies)
	if err != nil {
		return nil, err
	}

	// 拦截器链装配换库（grpcapi）：Assemble 按职能槽位构造期断言顺序
	//（ClientInfo 打头、Validate 收尾，auth 居中——三槽子集是合法形态）。
	// 模板原先没有 clientInfo 拦截器，经库免费获得（客户端 IP/UA 注入 ctx；
	// 模板 config 无 trusted_proxies 配置，默认不信任转发头，一律使用 gRPC
	// peer 地址）。
	chain, err := grpcapiinterceptor.Assemble(
		grpcapiinterceptor.ChainItem{
			Slot:        grpcapiinterceptor.SlotClientInfo,
			Interceptor: grpcapiinterceptor.NewClientInfo(grpcapiinterceptor.ClientInfoConfig{}).Unary(),
		},
		grpcapiinterceptor.ChainItem{
			Slot:        grpcapiinterceptor.SlotAuth,
			Interceptor: authInterceptor.UnaryAuthMiddleware,
		},
		grpcapiinterceptor.ChainItem{
			Slot:        grpcapiinterceptor.SlotValidate,
			Interceptor: grpcapiinterceptor.NewValidate().Unary(),
		},
	)
	if err != nil {
		return nil, err
	}

	srv := lynxgrpc.NewServer(lynxgrpc.WithAddr(addr), lynxgrpc.WithTimeout(timeout), lynxgrpc.WithLogger(app.Logger()), lynxgrpc.WithInterceptors(chain...))
	grpcSrv := srv.GetServer()

	// Register public API services
	apipb.RegisterAuthServiceServer(grpcSrv, auth)
	apipb.RegisterGroupsServiceServer(grpcSrv, groups)
	apipb.RegisterUsersServiceServer(grpcSrv, users)

	// 启动断言（fail-closed）：每个已注册方法必须在策略集中，proto 注解与
	// 注册实现漂移启动即红。
	if err := grpcapiinterceptor.AssertAllRegisteredHavePolicy(grpcSrv, policies); err != nil {
		return nil, err
	}

	return srv, nil
}
