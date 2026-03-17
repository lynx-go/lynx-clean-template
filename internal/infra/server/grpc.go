package server

import (
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
	app lynx.Lynx,
	cfg *config.AppConfig,
	authValidator interceptor.Validator,
	auth *grpcsvc.AuthService,
	groups *grpcsvc.GroupsService,
	users *grpcsvc.UsersService,
) (*lynxgrpc.Server, error) {
	c := cfg.GetServer().GetGrpc()
	addr := c.Addr
	timeout := parseTimeout(c.Timeout)
	authInterceptor, err := interceptor.NewAuthInterceptor(authValidator,
		[]string{
			apipb.AuthService_Token_FullMethodName,
		})
	if err != nil {
		return nil, err
	}
	validateInterceptor, err := interceptor.NewProtoValidateInterceptor()
	if err != nil {
		return nil, err
	}

	// Create API key interceptor for developer API services
	//apiKeyInterceptor := interceptor.NewAPIKeyInterceptor(apiKeyChecker)

	srv := lynxgrpc.NewServer(lynxgrpc.WithAddr(addr), lynxgrpc.WithTimeout(timeout), lynxgrpc.WithLogger(app.Logger()), lynxgrpc.WithInterceptors(
		authInterceptor.UnaryAuthMiddleware,
		validateInterceptor,
		//apiKeyInterceptor.UnaryAPIKeyMiddleware,
	))
	grpcSrv := srv.GetServer()

	// Register public API services
	apipb.RegisterAuthServiceServer(grpcSrv, auth)
	apipb.RegisterGroupsServiceServer(grpcSrv, groups)
	apipb.RegisterUsersServiceServer(grpcSrv, users)

	return srv, nil
}
