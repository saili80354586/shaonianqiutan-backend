package services

import (
	"fmt"
	"github.com/shaonianqiutan/backend/models"
)

// AnalystApplicationService 分析师申请服务
type AnalystApplicationService struct {
	appRepo  *models.AnalystApplicationRepository
	userRepo *models.UserRepository
}

func NewAnalystApplicationService(
	appRepo *models.AnalystApplicationRepository,
	userRepo *models.UserRepository,
) *AnalystApplicationService {
	return &AnalystApplicationService{
		appRepo:  appRepo,
		userRepo: userRepo,
	}
}

// CreateApplicationRequest 创建申请请求
type CreateApplicationRequest struct {
	Name       string `json:"name" binding:"required"`
	Phone      string `json:"phone" binding:"required"`
	Email      string `json:"email"`
	Experience string `json:"experience" binding:"required"`
	Resume     string `json:"resume"`
}

// CreateApplication 创建分析师申请
func (s *AnalystApplicationService) CreateApplication(
	userID uint,
	req *CreateApplicationRequest,
) (*models.AnalystApplication, error) {
	// 检查是否已经存在申请
	existing, err := s.appRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("您已经提交过申请，请等待审核")
	}

	app := &models.AnalystApplication{
		UserID:     userID,
		Name:       req.Name,
		Phone:      req.Phone,
		Email:      req.Email,
		Experience: req.Experience,
		Resume:     req.Resume,
		Status:     models.ApplicationStatusPending,
	}

	err = s.appRepo.Create(app)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// GetMyApplication 获取我的申请
func (s *AnalystApplicationService) GetMyApplication(userID uint) (*models.AnalystApplication, error) {
	return s.appRepo.FindByUserID(userID)
}

// GetApplicationList 获取申请列表（管理后台）
func (s *AnalystApplicationService) GetApplicationList(
	page, pageSize int,
	status *models.ApplicationStatus,
) ([]models.AnalystApplication, int64, error) {
	return s.appRepo.FindAll(page, pageSize, status)
}

// ReviewApplication 审核申请
func (s *AnalystApplicationService) ReviewApplication(
	id uint,
	status models.ApplicationStatus,
	remark string,
) error {
	app, err := s.appRepo.FindByID(id)
	if err != nil {
		return err
	}
	if app == nil {
		return fmt.Errorf("申请不存在")
	}
	if app.Status != models.ApplicationStatusPending {
		return fmt.Errorf("申请已审核过")
	}

	// 更新申请状态
	err = s.appRepo.UpdateStatus(id, status, remark)
	if err != nil {
		return err
	}

	// 如果批准，将用户角色升级为分析师
	if status == models.ApplicationStatusApproved {
		updates := map[string]interface{}{
			"role": models.RoleAnalyst,
		}
		err = s.userRepo.Update(app.UserID, updates)
		if err != nil {
			return err
		}
	}

	return nil
}
