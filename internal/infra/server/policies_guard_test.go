package server

import (
	"path/filepath"
	"testing"

	"github.com/lynx-go/grpcapi/authz"
	"github.com/lynx-go/grpcapi/guard"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// swaggerErrorResponseRef 与各服务 proto 文件级 openapiv2_swagger.responses.default
// 声明的引用经生成器解析后的定义名保持一致（legacy 命名：lynx.shared 包经
// 最短唯一后缀落在 shared 前缀下）。
const swaggerErrorResponseRef = "#/definitions/sharedErrorResponse"

// serviceDefaultAccess 从 proto service_auth 注解解析服务默认档（与启动期
// authz.Build 的回落语义同源）——库导出的 ServiceDefaultAccess + String
// 即规范形态（3 行，无需项目自写扩展解析与字符串映射）。
func serviceDefaultAccess(service string) (string, bool) {
	for _, fd := range authzFileDescriptors() {
		for i := 0; i < fd.Services().Len(); i++ {
			s := fd.Services().Get(i)
			if string(s.FullName()) == service {
				return serviceDefaultAccessOf(s)
			}
		}
	}
	return "", false
}

func serviceDefaultAccessOf(s protoreflect.ServiceDescriptor) (string, bool) {
	lv, ok := authz.ServiceDefaultAccess(s)
	return lv.String(), ok
}

// TestSwaggerAccessMatchesPolicies 断言 genproto/api/v1/*.swagger.json 的
// x-access 扩展与 authz.PolicySet 完全一致（README 第 6 步守卫）：顶层扩展
// == 服务默认档；operation 有效扩展（显式否则继承顶层）== 方法策略档位；
// default 响应引用运行时真实错误体；策略表方法反向全覆盖。前提：仓库根执行
// 过 mise run generate:proto（genproto 已入库）。
func TestSwaggerAccessMatchesPolicies(t *testing.T) {
	policies, err := NewAuthzPolicySet()
	if err != nil {
		t.Fatalf("NewAuthzPolicySet: %v", err)
	}

	guard.SwaggerAccessMatches(t, guard.SwaggerOptions{
		GenprotoRoot:     filepath.Join("..", "..", "..", "genproto"),
		GenprotoSubdirs:  []string{"api/v1"},
		Files:            authzFileDescriptors(),
		Policies:         policies,
		AccessExtension:  "x-access",
		ErrorResponseRef: swaggerErrorResponseRef,
		MinFiles:         4,
		MinOps:           14,
		ServiceAccess:    serviceDefaultAccess,
	})

	// RPC 快照锁：lynx.api.v1 下方法总数（AuthService 4 + UsersService 4 +
	// GroupsService 5 + FilesService 1 = 14）。增删 RPC 请同步更新此常量。
	guard.AssertMethodCount(t, "lynx.api.v1", 14)
}

// TestSwaggerPropertyNamesAreSnakeCase 守卫 buf.gen.yaml 的
// json_names_for_fields=false（与运行时 marshaler UseProtoNames 对齐）。
func TestSwaggerPropertyNamesAreSnakeCase(t *testing.T) {
	guard.SwaggerSnakeCaseProperties(t,
		filepath.Join("..", "..", "..", "genproto"), []string{"api/v1"})
}

// TestSwaggerNoRpcStatus 守卫 buf.gen.yaml 的 disable_default_errors=true
// （生成器默认注入的 rpcStatus 与运行时错误体 lynx.shared.ErrorResponse 不符，
// 配置回退即红）。
func TestSwaggerNoRpcStatus(t *testing.T) {
	guard.SwaggerNoDefinition(t,
		filepath.Join("..", "..", "..", "genproto"), []string{"api/v1"}, "rpcStatus")
}

// TestBuildMatchesContract 等价锁：注解收集结果与鉴权意图一致（服务级回落
// PUBLIC/END_USER、方法级 PERMISSION 提升）。
func TestBuildMatchesContract(t *testing.T) {
	set, err := NewAuthzPolicySet()
	if err != nil {
		t.Fatalf("NewAuthzPolicySet: %v", err)
	}

	want := map[string]authz.AccessLevel{
		"/lynx.api.v1.AuthService/Token":                 authz.AccessPublic,
		"/lynx.api.v1.AuthService/SignUp":                authz.AccessPublic,
		"/lynx.api.v1.AuthService/VerifySignUpEmail":     authz.AccessPublic,
		"/lynx.api.v1.AuthService/ResendSignUpEmailCode": authz.AccessPublic,
		"/lynx.api.v1.UsersService/UpdateMyProfile":      authz.AccessEndUser,
		"/lynx.api.v1.UsersService/GetUserProfile":       authz.AccessEndUser,
		"/lynx.api.v1.UsersService/GrantSuperAdmin":      authz.AccessPermission,
		"/lynx.api.v1.UsersService/RevokeSuperAdmin":     authz.AccessPermission,
		"/lynx.api.v1.GroupsService/CreateGroup":         authz.AccessEndUser,
		"/lynx.api.v1.GroupsService/GetGroup":            authz.AccessEndUser,
		"/lynx.api.v1.GroupsService/ListGroups":          authz.AccessEndUser,
		"/lynx.api.v1.GroupsService/UpdateGroup":         authz.AccessEndUser,
		"/lynx.api.v1.GroupsService/DeleteGroup":         authz.AccessEndUser,
		"/lynx.api.v1.FilesService/GetPresignedURL":      authz.AccessEndUser,
	}
	for method, access := range want {
		p, ok := set.Get(method)
		if !ok {
			t.Fatalf("missing policy for %s", method)
		}
		if p.Access != access {
			t.Errorf("%s: access = %v, want %v", method, p.Access, access)
		}
	}

	// 空 scope 面：END_USER/PERMISSION 方法不得携带 scope（SERVER 面专属）。
	for _, p := range set.Methods() {
		if p.Scope != nil {
			t.Errorf("%s: scope = %+v, want nil（模板无 API key 面）", p.Method, p.Scope)
		}
		if p.IsStreaming {
			t.Errorf("%s: 不应为流式方法", p.Method)
		}
	}
}
