package server

import (
	"github.com/lynx-go/grpcapi/authz"
	apipb "github.com/lynx-go/lynx-clean-template/genproto/api/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// authzFileDescriptors 是业务 proto 文件描述符的单一事实源：启动期策略收集
// （NewAuthzPolicySet）与守卫测试（swagger access 对齐检查）共用同一清单。
// FilesService 当前未注册到 gRPC server，但策略先行收集——注册后即被
// AssertAllRegisteredHavePolicy 覆盖，无需再补注解。
func authzFileDescriptors() []protoreflect.FileDescriptor {
	return []protoreflect.FileDescriptor{
		apipb.File_api_v1_auth_proto,
		apipb.File_api_v1_users_proto,
		apipb.File_api_v1_groups_proto,
		apipb.File_api_v1_files_proto,
	}
}

// NewAuthzPolicySet 从 proto 注解（grpcapi.v1.method_auth / service_auth）收集
// 全量 API 授权策略，构造期 fail-closed（缺注解、档位约束违反、流式方法等
// 启动即失败）。模板没有 API key scope 面，无 scope 声明时 Vocabulary 传空
// 是合法路径（空词表断言仅在存在 scope 声明时触发）。
func NewAuthzPolicySet() (*authz.PolicySet, error) {
	return authz.Build(authzFileDescriptors(), authz.Options{})
}
