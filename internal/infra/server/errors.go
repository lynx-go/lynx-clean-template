package server

import (
	"github.com/lynx-go/grpcapi/gateway"
	sharedpb "github.com/lynx-go/lynx-clean-template/genproto/shared"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// ErrorType represents the category of error
type ErrorType string

func (et ErrorType) String() string {
	return string(et)
}

const (
	ErrorTypeServerError       ErrorType = "server_error"
	ErrorTypeError             ErrorType = "error"
	ErrorTypeInvalidRequest    ErrorType = "invalid_request_error"
	ErrorTypeNotFound          ErrorType = "not_found_error"
	ErrorTypeUnauthorized      ErrorType = "authentication_error"
	ErrorTypeForbidden         ErrorType = "permission_error"
	ErrorTypeConflict          ErrorType = "conflict_error"
	ErrorTypeRateLimitExceeded ErrorType = "rate_limit_error"
)

// ErrorResponseBuilder 实现 grpcapi gateway.ErrorBodyBuilder：错误体形状沿用
// 模板对外契约（shared.ErrorResponse{error:{type,code,message,params}}）。
// 机制（gRPC code → HTTP 状态映射、Internal/Unknown 脱敏、429 Retry-After、
// error_id 生成）由库的 NewErrorHandler 承担：
//   - code 字段 = gRPC code 字符串（如 "InvalidArgument"，与旧实现一致）；
//   - type 字段 = 下方 grpcCodeToErrorType 映射（模板既有契约）；
//   - error_id 填入库生成的追踪 ID，经 params.error_id 下发（新增字段，
//     不破坏既有字段语义）。
type ErrorResponseBuilder struct{}

// Build 构造最终写出的错误体。
func (ErrorResponseBuilder) Build(code codes.Code, message, errorID string) proto.Message {
	params, err := structpb.NewStruct(map[string]any{"error_id": errorID})
	if err != nil {
		params = &structpb.Struct{}
	}
	return &sharedpb.ErrorResponse{
		Error: &sharedpb.Error{
			Type:    string(grpcCodeToErrorType(code)),
			Code:    code.String(),
			Message: message,
			Params:  params,
		},
	}
}

// MapErrorCode 声明 gRPC code → 项目业务错误码的映射。模板错误体的 code 即
// gRPC code 字符串，在 Build 内直接推导，没有独立映射表——恒返回 nil。
func (ErrorResponseBuilder) MapErrorCode(codes.Code) any { return nil }

var _ gateway.ErrorBodyBuilder = ErrorResponseBuilder{}

// grpcCodeToErrorType maps gRPC codes to error types
func grpcCodeToErrorType(code codes.Code) ErrorType {
	switch code {
	case codes.NotFound:
		return ErrorTypeNotFound
	case codes.InvalidArgument, codes.OutOfRange, codes.FailedPrecondition:
		return ErrorTypeInvalidRequest
	case codes.Unauthenticated:
		return ErrorTypeUnauthorized
	case codes.PermissionDenied:
		return ErrorTypeForbidden
	case codes.AlreadyExists, codes.Aborted:
		return ErrorTypeConflict
	case codes.ResourceExhausted:
		return ErrorTypeRateLimitExceeded
	case codes.Internal, codes.Unknown, codes.DataLoss, codes.Unavailable:
		return ErrorTypeServerError
	default:
		return ErrorTypeServerError
	}
}
