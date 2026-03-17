package interceptor

import (
	"buf.build/go/protovalidate"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"google.golang.org/grpc"
)

func NewProtoValidateInterceptor() (grpc.UnaryServerInterceptor, error) {

	// Create a Protovalidate Validator
	validator, err := protovalidate.New()
	if err != nil {
		return nil, err
	}

	// Use the protovalidate_middleware interceptor provided by grpc-ecosystem
	return protovalidate_middleware.UnaryServerInterceptor(validator), nil
}
