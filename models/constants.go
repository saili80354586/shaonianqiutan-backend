package models

// 球员状态常量
const (
	PlayerStatusActive   = "active"
	PlayerStatusInactive = "inactive"
	PlayerStatusLeft    = "left"
)

// 俱乐部球员位置常量
const (
	PositionForward    = "forward"
	PositionMidfielder = "midfielder"
	PositionDefender   = "defender"
	PositionGoalkeeper = "goalkeeper"
)

// 服务类型常量
const (
	ServiceTypeQuickReport   = "quick_report"
	ServiceTypeFullReport  = "full_report"
	ServiceTypeVideoAnalysis = "video_analysis"
)

// GetPositionName 获取位置中文名
func GetPositionName(position string) string {
	positions := map[string]string{
		PositionForward:    "前锋",
		PositionMidfielder: "中场",
		PositionDefender:   "后卫",
		PositionGoalkeeper: "守门员",
	}
	if name, ok := positions[position]; ok {
		return name
	}
	return position
}

// GetPTStatusName 获取体测状态中文名
func GetPTStatusName(status string) string {
	names := map[string]string{
		"pending":           "待开始",
		"ongoing":           "进行中",
		"completed":         "已完成",
		"report_generated":   "报告已生成",
	}
	if name, ok := names[status]; ok {
		return name
	}
	return status
}

// GetServiceTypeName 获取服务类型中文名
func GetServiceTypeName(serviceType string) string {
	names := map[string]string{
		"quick_report":    "快速分析报告",
		"full_report":     "全方位技术分析报告",
		"video_analysis":  "视频分析报告",
	}
	if name, ok := names[serviceType]; ok {
		return name
	}
	return serviceType
}

// GetPTTemplateName 获取体测模板中文名
func GetPTTemplateName(template string) string {
	names := map[string]string{
		"basic":        "基础版",
		"advanced":     "进阶版",
		"professional": "专业版",
		"custom":      "自定义",
	}
	if name, ok := names[template]; ok {
		return name
	}
	return template
}
