package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	sharedpb "github.com/lynx-go/lynx-clean-template/genproto/shared"
	"github.com/lynx-go/x/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
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

// HTTPErrorHandler is a custom error handler for grpc-gateway
var HTTPErrorHandler runtime.ErrorHandlerFunc = func(
	ctx context.Context,
	mux *runtime.ServeMux,
	marshaler runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	// Default to internal server error
	httpStatusCode := http.StatusInternalServerError
	errorType := ErrorTypeServerError
	errorCode := ""
	errorMessage := "An internal error occurred"

	if err != nil {
		// Extract service path and endpoint from URL
		// URL path format: /{api}/{resource} or /{api}/{resource}/{id}
		apiPath, endpoint := extractAPIPathAndEndpoint(r.URL.Path)
		log.ErrorContext(ctx, "http response error", err,
			"api", apiPath,
			"endpoint", endpoint)

		// Extract gRPC status and code
		st, ok := status.FromError(err)
		if ok {
			errorCode = st.Code().String()
			errorMessage = st.Message()
			httpStatusCode = grpcCodeToHTTPStatus(st.Code())
			errorType = grpcCodeToErrorType(st.Code())
		} else {
			errorMessage = err.Error()
		}
	}

	// Create error response
	errorResp := &sharedpb.ErrorResponse{
		Error: &sharedpb.Error{
			Type:    string(errorType),
			Code:    errorCode,
			Message: errorMessage,
		},
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Write status code
	w.WriteHeader(httpStatusCode)

	// Marshal and write response
	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		log.ErrorContext(ctx, "write http response error", err)
	}
}

// extractAPIPathAndEndpoint extracts API path and endpoint from URL
// Examples:
//   - /admin/v1/subscription-types -> api="admin/v1", endpoint="subscription-types"
//   - /api/v1/users/123 -> api="api/v1", endpoint="users"
//   - /api/v1/groups/abc/projects -> api="api/v1", endpoint="groups/{id}/projects"
func extractAPIPathAndEndpoint(path string) (apiPath, endpoint string) {
	// Remove leading slash
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	if path == "" {
		return "", ""
	}

	// Split path into segments
	segments := splitPath(path)
	if len(segments) < 2 {
		return path, ""
	}

	// First two segments are typically the API path (e.g., "admin/v1", "api/v1")
	apiPath = segments[0] + "/" + segments[1]

	if len(segments) == 2 {
		return apiPath, ""
	}

	// Build endpoint, replacing UUID/id segments with placeholders
	var endpointParts []string
	for i := 2; i < len(segments); i++ {
		seg := segments[i]
		if isUUIDOrID(seg) {
			endpointParts = append(endpointParts, "{id}")
		} else {
			endpointParts = append(endpointParts, seg)
		}
	}

	// Remove consecutive {id} placeholders, keep only one
	var cleanedParts []string
	for i, part := range endpointParts {
		if i > 0 && part == "{id}" && endpointParts[i-1] == "{id}" {
			continue
		}
		cleanedParts = append(cleanedParts, part)
	}

	endpoint = joinParts(cleanedParts, "/")
	return apiPath, endpoint
}

func splitPath(path string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			if i > start {
				result = append(result, path[start:i])
			}
			start = i + 1
		}
	}
	return result
}

func isUUIDOrID(s string) bool {
	// Check if it looks like a UUID (contains dashes and is long enough)
	if len(s) >= 32 && containsDash(s) {
		return true
	}
	// Check if it's all digits
	if isAllDigits(s) && len(s) > 0 {
		return true
	}
	// Check if it looks like a random ID (alphanumeric, not all lowercase words)
	if len(s) > 10 && isAlphanumeric(s) && !isAllLowercase(s) {
		return true
	}
	return false
}

func containsDash(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '-' {
			return true
		}
	}
	return false
}

func isAllDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isAlphanumeric(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func isAllLowercase(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			return false
		}
	}
	return true
}

func joinParts(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}

// grpcCodeToHTTPStatus maps gRPC codes to HTTP status codes
func grpcCodeToHTTPStatus(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return 499 // Client Closed Request
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

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

// CustomMarshaler is a custom marshaler that uses protojson with proper settings
type CustomMarshaler struct {
	*runtime.JSONPb
}

// NewCustomMarshaler creates a new custom marshaler
func NewCustomMarshaler() runtime.Marshaler {
	return &CustomMarshaler{
		JSONPb: &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: false,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: false,
			},
		},
	}
}

// ContentType returns the Content-Type header for JSON responses
func (m *CustomMarshaler) ContentType(_ interface{}) string {
	return "application/json"
}

// Marshal marshals "v" into JSON
func (m *CustomMarshaler) Marshal(v interface{}) ([]byte, error) {
	return m.JSONPb.Marshal(v)
}

// Unmarshal unmarshals JSON data into "v"
func (m *CustomMarshaler) Unmarshal(data []byte, v interface{}) error {
	return m.JSONPb.Unmarshal(data, v)
}

// NewDecoder returns a JSON decoder for the given reader
func (m *CustomMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return m.JSONPb.NewDecoder(r)
}

// NewEncoder returns a JSON encoder for the given writer
func (m *CustomMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return m.JSONPb.NewEncoder(w)
}
