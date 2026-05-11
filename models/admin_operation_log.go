package models

import "time"

// AdminOperationLog 管理员操作日志
type AdminOperationLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ClubID    uint      `gorm:"index;default:0" json:"clubId"`
	AdminID   uint      `gorm:"index;not null" json:"adminId"`
	AdminName string    `gorm:"size:64" json:"adminName"`
	Action    string    `gorm:"size:64;not null" json:"action"` // 操作类型: delete_team, export_players, remove_coach 等
	Target    string    `gorm:"size:64" json:"target"`          // 操作对象类型: team, player, coach 等
	TargetID  uint      `gorm:"default:0" json:"targetId"`      // 操作对象ID
	Detail    string    `gorm:"type:text" json:"detail"`        // 详细描述
	IP        string    `gorm:"size:64" json:"ip"`              // 操作者IP
	CreatedAt time.Time `json:"createdAt"`
}

// AdminOperationLogResponse 前端响应结构
type AdminOperationLogResponse struct {
	ID        uint   `json:"id"`
	AdminName string `json:"adminName"`
	Action    string `json:"action"`
	Target    string `json:"target"`
	TargetID  uint   `json:"targetId"`
	Detail    string `json:"detail"`
	IP        string `json:"ip"`
	CreatedAt string `json:"createdAt"`
}

// ToResponse 转换为响应结构
func (l *AdminOperationLog) ToResponse() *AdminOperationLogResponse {
	return &AdminOperationLogResponse{
		ID:        l.ID,
		AdminName: l.AdminName,
		Action:    l.Action,
		Target:    l.Target,
		TargetID:  l.TargetID,
		Detail:    l.Detail,
		IP:        l.IP,
		CreatedAt: l.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ActionDisplayName 操作类型显示名称
var ActionDisplayName = map[string]string{
	"delete_team":              "删除球队",
	"restore_team":             "恢复球队",
	"export_players":           "导出球员名单",
	"remove_coach":             "移除教练",
	"remove_player":            "移除球员",
	"create_announcement":      "发布公告",
	"update_announcement":      "编辑公告",
	"delete_announcement":      "删除公告",
	"create_order":             "创建订单",
	"create_match":             "创建比赛",
	"create_weekly":            "发起周报",
	"remind_weekly":            "催办周报",
	"remind_match":             "催办比赛自评",
	"remind_physical":          "催办体测录入",
	"login":                    "登录后台",
	"update_system_settings":   "更新系统设置",
	"update_user":              "编辑用户",
	"update_user_status":       "更新用户状态",
	"delete_user":              "删除用户",
	"create_admin_role":        "创建管理员子角色",
	"update_admin_role":        "更新管理员子角色",
	"update_admin_role_status": "更新管理员子角色状态",
	"assign_admin_role":        "分配管理员子角色",
	"review_role_application":  "审核角色申请",
	"assign_order":             "派发订单",
	"cancel_order":             "取消订单",
	"review_report":            "审核报告",
	"audit_analyst":            "审核分析师",
	"update_analyst_status":    "更新分析师状态",
	"process_settlement":       "处理结算",
	"handle_content_report":    "处理举报",
	"create_sensitive_word":    "创建敏感词",
	"update_sensitive_word":    "更新敏感词",
	"delete_sensitive_word":    "删除敏感词",
	"create_banner":            "创建轮播图",
	"update_banner":            "更新轮播图",
	"delete_banner":            "删除轮播图",
	"create_faq":               "创建FAQ",
	"update_faq":               "更新FAQ",
	"delete_faq":               "删除FAQ",
}
