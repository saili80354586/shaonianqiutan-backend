package models

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

const AdminRoleSuper = "super_admin"

// AdminPermission 管理员接口/菜单权限点
type AdminPermission struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Code        string    `json:"code" gorm:"uniqueIndex;size:80;not null"`
	Name        string    `json:"name" gorm:"size:80;not null"`
	Module      string    `json:"module" gorm:"size:40;index"`
	Description string    `json:"description" gorm:"size:255"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AdminRole 管理员子角色
type AdminRole struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Key         string    `json:"key" gorm:"uniqueIndex;size:40;not null"`
	Name        string    `json:"name" gorm:"size:80;not null"`
	Description string    `json:"description" gorm:"size:255"`
	BuiltIn     bool      `json:"built_in" gorm:"default:false"`
	Enabled     bool      `json:"enabled" gorm:"default:true;index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AdminRolePermission 管理员子角色与权限点关系
type AdminRolePermission struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	RoleKey        string    `json:"role_key" gorm:"size:40;index:idx_admin_role_permission,unique;not null"`
	PermissionCode string    `json:"permission_code" gorm:"size:80;index:idx_admin_role_permission,unique;not null"`
	CreatedAt      time.Time `json:"created_at"`
}

// AdminUserRole 管理员账号被分配的子角色；没有记录时按 super_admin 兼容历史管理员。
type AdminUserRole struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"uniqueIndex;not null"`
	RoleKey    string    `json:"role_key" gorm:"size:40;index;not null"`
	AssignedBy uint      `json:"assigned_by" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type DefaultAdminPermission struct {
	Code        string
	Name        string
	Module      string
	Description string
}

type DefaultAdminRole struct {
	Key         string
	Name        string
	Description string
	Permissions []string
}

func DefaultAdminPermissions() []DefaultAdminPermission {
	return []DefaultAdminPermission{
		{Code: "dashboard.view", Name: "数据看板查看", Module: "dashboard"},
		{Code: "operations.view", Name: "运营洞察查看", Module: "dashboard"},
		{Code: "users.manage", Name: "用户管理", Module: "users"},
		{Code: "orders.manage", Name: "订单管理", Module: "orders"},
		{Code: "dispatch.manage", Name: "订单派发", Module: "orders"},
		{Code: "applications.review", Name: "申请审核", Module: "review"},
		{Code: "reports.review", Name: "报告审核", Module: "review"},
		{Code: "content.manage", Name: "内容治理", Module: "content"},
		{Code: "finance.manage", Name: "财务结算", Module: "finance"},
		{Code: "platform.manage", Name: "平台运营配置", Module: "platform"},
		{Code: "settings.manage", Name: "系统参数配置", Module: "settings"},
		{Code: "audit.view", Name: "操作审计查看", Module: "audit"},
		{Code: "login_logs.view", Name: "登录日志查看", Module: "audit"},
		{Code: "role_permissions.view", Name: "角色权限查看", Module: "rbac"},
		{Code: "role_permissions.manage", Name: "角色权限管理", Module: "rbac"},
		{Code: "exceptions.view", Name: "异常管控查看", Module: "risk"},
		{Code: "exceptions.manage", Name: "异常管控处理", Module: "risk"},
	}
}

func DefaultAdminRoles() []DefaultAdminRole {
	all := make([]string, 0, len(DefaultAdminPermissions()))
	for _, permission := range DefaultAdminPermissions() {
		all = append(all, permission.Code)
	}
	return []DefaultAdminRole{
		{Key: AdminRoleSuper, Name: "超级管理员", Description: "拥有平台管理后台全部权限", Permissions: all},
		{Key: "operations_admin", Name: "运营管理员", Description: "负责用户、订单、内容与平台运营配置", Permissions: []string{"dashboard.view", "operations.view", "users.manage", "orders.manage", "dispatch.manage", "content.manage", "platform.manage", "audit.view", "exceptions.view"}},
		{Key: "review_admin", Name: "审核管理员", Description: "负责申请、报告、举报与异常处理", Permissions: []string{"dashboard.view", "applications.review", "reports.review", "content.manage", "audit.view", "exceptions.view", "exceptions.manage"}},
		{Key: "finance_admin", Name: "财务管理员", Description: "负责结算、收支和相关订单查看", Permissions: []string{"dashboard.view", "orders.manage", "finance.manage", "audit.view"}},
		{Key: "config_admin", Name: "配置管理员", Description: "负责系统参数、平台运营配置和权限查看", Permissions: []string{"dashboard.view", "platform.manage", "settings.manage", "role_permissions.view", "audit.view"}},
	}
}

func SeedDefaultAdminRBAC(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}

	for _, item := range DefaultAdminPermissions() {
		permission := AdminPermission{
			Code:        item.Code,
			Name:        item.Name,
			Module:      item.Module,
			Description: item.Description,
		}
		if err := db.Where("code = ?", item.Code).Assign(permission).FirstOrCreate(&permission).Error; err != nil {
			return err
		}
	}

	for _, item := range DefaultAdminRoles() {
		role := AdminRole{
			Key:         item.Key,
			Name:        item.Name,
			Description: item.Description,
			BuiltIn:     true,
			Enabled:     true,
		}
		if err := db.Where("key = ?", item.Key).Assign(role).FirstOrCreate(&role).Error; err != nil {
			return err
		}
		for _, permissionCode := range item.Permissions {
			relation := AdminRolePermission{RoleKey: item.Key, PermissionCode: permissionCode}
			if err := db.Where("role_key = ? AND permission_code = ?", item.Key, permissionCode).FirstOrCreate(&relation).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func IsMissingAdminRBACTableError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "no such table") ||
		strings.Contains(message, "doesn't exist") ||
		strings.Contains(message, "unknown table")
}
