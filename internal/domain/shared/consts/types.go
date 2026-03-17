package consts

type Role string

func (r Role) String() string {
	return string(r)
}

const (
	// RoleOwner 所有者
	RoleOwner Role = "owner"
	// RoleMaintainer 维护者
	RoleMaintainer Role = "maintainer"
	// RoleEditor 编辑
	RoleEditor Role = "editor"
	// RoleViewer 查看
	RoleViewer Role = "viewer"
)
