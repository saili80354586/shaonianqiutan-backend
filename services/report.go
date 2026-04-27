package services

import (
	"path"

	"github.com/shaonianqiutan/backend/models"
)

// ReportService 报告服务
type ReportService struct {
	reportRepo *models.ReportRepository
	userRepo   *models.UserRepository
}

func NewReportService(reportRepo *models.ReportRepository, userRepo *models.UserRepository) *ReportService {
	return &ReportService{
		reportRepo: reportRepo,
		userRepo:   userRepo,
	}
}

// CreateReportRequest 创建报告请求
type CreateReportRequest struct {
	OrderID         uint   `json:"order_id" binding:"required"`
	PlayerName      string `json:"player_name" binding:"required"`
	PlayerBirthDate string `json:"player_birth_date"`
	PlayerPosition  string `json:"player_position"`
	PlayerProvince  string `json:"player_province"`
	PlayerCity      string `json:"player_city"`
	Content         string `json:"content" binding:"required"`
}

// CreateReport 创建报告
func (s *ReportService) CreateReport(req *CreateReportRequest, analystID uint) (*models.Report, error) {
	// 创建报告记录
	report := &models.Report{
		OrderID:         req.OrderID,
		UserID:          0, // TODO: 需要从订单中获取实际买家ID，订单模块实现后修复
		AnalystID:       analystID,
		PlayerName:      req.PlayerName,
		PlayerBirthDate: req.PlayerBirthDate,
		PlayerPosition:  req.PlayerPosition,
		PlayerProvince:  req.PlayerProvince,
		PlayerCity:      req.PlayerCity,
		Content:         req.Content,
		Status:          models.ReportStatusProcessing,
	}

	err := s.reportRepo.Create(report)
	if err != nil {
		return nil, err
	}

	return report, nil
}

// GetReportDetail 获取报告详情
func (s *ReportService) GetReportDetail(id uint, userID uint, userRole models.UserRole) (*models.Report, bool, error) {
	report, err := s.reportRepo.FindByID(id)
	if err != nil {
		return nil, false, err
	}
	if report == nil {
		return nil, false, nil // 报告不存在
	}

	// 检查权限：只有购买用户、分析师和管理员可以查看
	if report.UserID != userID && report.AnalystID != userID && userRole != models.RoleAdmin {
		return nil, false, nil // 无权限
	}

	return report, true, nil
}

// CheckDownloadPermission 检查下载权限
func (s *ReportService) CheckDownloadPermission(id uint, userID uint, userRole models.UserRole) (*models.Report, bool, error) {
	report, err := s.reportRepo.FindByID(id)
	if err != nil {
		return nil, false, err
	}
	if report == nil {
		return nil, false, nil // 报告不存在
	}

	// 检查权限：只有购买用户、分析师和管理员可以下载
	if report.UserID != userID && report.AnalystID != userID && userRole != models.RoleAdmin {
		return nil, false, nil // 无权限
	}

	return report, true, nil
}

// GetPdfFilePath 获取PDF文件路径
func (s *ReportService) GetPdfFilePath(report *models.Report) string {
	if report.PdfURL == "" {
		return ""
	}
	// 从项目根目录获取文件路径
	return path.Join(".", report.PdfURL)
}

// GetGlobalStatistics 获取全局报告统计(管理员)
func (s *ReportService) GetGlobalStatistics() (*models.ReportStatistics, error) {
	return s.reportRepo.GetStatistics()
}

// GetUserStatistics 获取用户报告统计
func (s *ReportService) GetUserStatistics(userID uint) (*models.ReportStatistics, error) {
	// 这里暂时返回空统计,用户统计需要在models中添加
	stats := &models.ReportStatistics{
		TotalCount:     0,
		TodayCount:     0,
		PendingCount:   0,
		CompletedCount: 0,
	}
	return stats, nil
}

// GetUserReports 获取用户购买的报告列表
func (s *ReportService) GetUserReports(userID uint, page, pageSize int) (*models.ReportListResult, error) {
	return s.reportRepo.FindByUserID(userID, page, pageSize)
}

// GetAnalystReports 获取分析师发布的报告列表
func (s *ReportService) GetAnalystReports(analystID uint, page, pageSize int) (*models.ReportListResult, error) {
	return s.reportRepo.FindByAnalystID(analystID, page, pageSize)
}

// UpdateReportStatus 更新报告状态
func (s *ReportService) UpdateReportStatus(id uint, status models.ReportStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}
	return s.reportRepo.Update(id, updates)
}

// UpdateReportPdf 更新报告PDF链接
func (s *ReportService) UpdateReportPdf(id uint, pdfUrl string) error {
	updates := map[string]interface{}{
		"pdf_url": pdfUrl,
		"status":  models.ReportStatusCompleted,
	}
	return s.reportRepo.Update(id, updates)
}

// RegeneratePdf 重新生成PDF
func (s *ReportService) RegeneratePdf(id uint, userID uint, userRole models.UserRole) (*models.Report, bool, error) {
	report, err := s.reportRepo.FindByID(id)
	if err != nil {
		return nil, false, err
	}
	if report == nil {
		return nil, false, nil // 报告不存在
	}

	// 检查权限：只有分析师本人和管理员可以重新生成
	if report.AnalystID != userID && userRole != models.RoleAdmin {
		return nil, false, nil // 无权限
	}

	// 更新状态为处理中
	err = s.UpdateReportStatus(id, models.ReportStatusProcessing)
	if err != nil {
		return nil, false, err
	}

	return report, true, nil
}
