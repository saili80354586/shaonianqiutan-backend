package services

import (
	"encoding/json"
	"errors"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminService 管理后台服务
type AdminService struct {
	userRepo            *models.UserRepository
	reportRepo          *models.ReportRepository
	orderRepo           *models.OrderRepository
	analystRepo         *models.AnalystRepository
	applicationRepo     *models.AnalystApplicationRepository
	contentReportRepo   *models.ContentReportRepository
	sensitiveWordRepo   *models.SensitiveWordRepository
	platformAnnRepo     *models.PlatformAnnouncementRepository
	bannerRepo          *models.BannerRepository
	faqRepo             *models.FAQRepository
	loginLogRepo        *models.LoginLogRepository
	videoAnalysisRepo   *models.VideoAnalysisRepository
	assignmentRepo      *models.OrderAssignmentRepository
	statusHistoryRepo   *models.OrderStatusHistoryRepository
	notificationService *NotificationService
}

type AdminPermissionDTO struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Module      string `json:"module"`
	Description string `json:"description"`
}

type AdminRoleDTO struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	BuiltIn     bool     `json:"built_in"`
	Enabled     bool     `json:"enabled"`
	Permissions []string `json:"permissions"`
}

type AdminRoleAssignmentDTO struct {
	UserID     uint       `json:"user_id"`
	Phone      string     `json:"phone"`
	Nickname   string     `json:"nickname"`
	Status     string     `json:"status"`
	RoleKey    string     `json:"role_key"`
	RoleName   string     `json:"role_name"`
	AssignedBy uint       `json:"assigned_by"`
	AssignedAt *time.Time `json:"assigned_at"`
}

type AdminRoleAssignmentResult struct {
	TargetUserID    uint   `json:"target_user_id"`
	TargetPhone     string `json:"target_phone"`
	TargetNickname  string `json:"target_nickname"`
	OldRoleKey      string `json:"old_role_key"`
	OldRoleName     string `json:"old_role_name"`
	NewRoleKey      string `json:"new_role_key"`
	NewRoleName     string `json:"new_role_name"`
	ChangedByUserID uint   `json:"changed_by_user_id"`
}

type AdminRoleAssignmentHistoryDTO struct {
	ID             uint      `json:"id"`
	AdminID        uint      `json:"admin_id"`
	AdminName      string    `json:"admin_name"`
	TargetUserID   uint      `json:"target_user_id"`
	TargetPhone    string    `json:"target_phone"`
	TargetNickname string    `json:"target_nickname"`
	OldRoleKey     string    `json:"old_role_key"`
	OldRoleName    string    `json:"old_role_name"`
	NewRoleKey     string    `json:"new_role_key"`
	NewRoleName    string    `json:"new_role_name"`
	Detail         string    `json:"detail"`
	IP             string    `json:"ip"`
	CreatedAt      time.Time `json:"created_at"`
}

type AdminRoleAssignmentHistoryResponse struct {
	List     []AdminRoleAssignmentHistoryDTO `json:"list"`
	Total    int64                           `json:"total"`
	Page     int                             `json:"page"`
	PageSize int                             `json:"pageSize"`
}

type AdminRolePermissionsResponse struct {
	Roles          []AdminRoleDTO           `json:"roles"`
	Permissions    []AdminPermissionDTO     `json:"permissions"`
	Admins         []AdminRoleAssignmentDTO `json:"admins"`
	CurrentRoleKey string                   `json:"current_role_key"`
}

type AdminRoleMutationRequest struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type CurrentAdminPermissionsResponse struct {
	RoleKey     string   `json:"role_key"`
	RoleName    string   `json:"role_name"`
	Permissions []string `json:"permissions"`
}

type AdminExceptionItem struct {
	ID         string    `json:"id"`
	Source     string    `json:"source"`
	Severity   string    `json:"severity"`
	Status     string    `json:"status"`
	Title      string    `json:"title"`
	Subject    string    `json:"subject"`
	Detail     string    `json:"detail"`
	RefID      uint      `json:"ref_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

type AdminExceptionsResponse struct {
	List     []AdminExceptionItem `json:"list"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
	Stats    map[string]int       `json:"stats"`
}

// GetCurrentAdminPermissions 获取当前管理员自身权限，供后台菜单裁剪使用。
func (s *AdminService) GetCurrentAdminPermissions(adminID uint) (*CurrentAdminPermissionsResponse, error) {
	db := s.GetDB()
	if db == nil {
		return nil, errors.New("数据库未初始化")
	}
	if err := models.SeedDefaultAdminRBAC(db); err != nil {
		return nil, err
	}

	roleKey := models.AdminRoleSuper
	var assignment models.AdminUserRole
	if err := db.Where("user_id = ?", adminID).First(&assignment).Error; err == nil && assignment.RoleKey != "" {
		roleKey = assignment.RoleKey
	} else if err != nil && err != gorm.ErrRecordNotFound && !models.IsMissingAdminRBACTableError(err) {
		return nil, err
	}

	var role models.AdminRole
	if err := db.Where("key = ? AND enabled = ?", roleKey, true).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound && roleKey != models.AdminRoleSuper {
			return &CurrentAdminPermissionsResponse{RoleKey: roleKey, Permissions: []string{}}, nil
		}
		return nil, err
	}

	permissionCodes := []string{}
	if role.Key == models.AdminRoleSuper {
		var permissions []models.AdminPermission
		if err := db.Order("code ASC").Find(&permissions).Error; err != nil {
			return nil, err
		}
		for _, permission := range permissions {
			permissionCodes = append(permissionCodes, permission.Code)
		}
	} else {
		var relations []models.AdminRolePermission
		if err := db.Where("role_key = ?", role.Key).Order("permission_code ASC").Find(&relations).Error; err != nil {
			return nil, err
		}
		for _, relation := range relations {
			permissionCodes = append(permissionCodes, relation.PermissionCode)
		}
	}

	sort.Strings(permissionCodes)
	return &CurrentAdminPermissionsResponse{
		RoleKey:     role.Key,
		RoleName:    role.Name,
		Permissions: permissionCodes,
	}, nil
}

// NewAdminService 创建管理后台服务
func NewAdminService(
	userRepo *models.UserRepository,
	reportRepo *models.ReportRepository,
	orderRepo *models.OrderRepository,
	analystRepo *models.AnalystRepository,
	applicationRepo *models.AnalystApplicationRepository,
	contentReportRepo *models.ContentReportRepository,
	sensitiveWordRepo *models.SensitiveWordRepository,
	platformAnnRepo *models.PlatformAnnouncementRepository,
	bannerRepo *models.BannerRepository,
	faqRepo *models.FAQRepository,
	loginLogRepo *models.LoginLogRepository,
	videoAnalysisRepo *models.VideoAnalysisRepository,
	assignmentRepo *models.OrderAssignmentRepository,
	statusHistoryRepo *models.OrderStatusHistoryRepository,
) *AdminService {
	return &AdminService{
		userRepo:          userRepo,
		reportRepo:        reportRepo,
		orderRepo:         orderRepo,
		analystRepo:       analystRepo,
		applicationRepo:   applicationRepo,
		contentReportRepo: contentReportRepo,
		sensitiveWordRepo: sensitiveWordRepo,
		platformAnnRepo:   platformAnnRepo,
		bannerRepo:        bannerRepo,
		faqRepo:           faqRepo,
		loginLogRepo:      loginLogRepo,
		videoAnalysisRepo: videoAnalysisRepo,
		assignmentRepo:    assignmentRepo,
		statusHistoryRepo: statusHistoryRepo,
	}
}

// GetDB exposes the shared database connection for controller-level audit writes.
func (s *AdminService) GetDB() *gorm.DB {
	if s == nil || s.orderRepo == nil {
		return nil
	}
	return s.orderRepo.GetDB()
}

// CreateAdminOperationLog 写入平台管理员操作审计日志
func (s *AdminService) CreateAdminOperationLog(logItem *models.AdminOperationLog) error {
	db := s.GetDB()
	if db == nil {
		return errors.New("数据库未初始化")
	}
	if logItem.CreatedAt.IsZero() {
		logItem.CreatedAt = time.Now()
	}
	return db.Create(logItem).Error
}

// GetAdminOperationLogs 获取平台管理员操作审计日志
func (s *AdminService) GetAdminOperationLogs(page, pageSize int, action, target, keyword string) ([]models.AdminOperationLog, int64, error) {
	db := s.GetDB()
	if db == nil {
		return nil, 0, errors.New("数据库未初始化")
	}

	var logs []models.AdminOperationLog
	var total int64
	query := db.Model(&models.AdminOperationLog{})
	if action = strings.TrimSpace(action); action != "" {
		query = query.Where("action = ?", action)
	}
	if target = strings.TrimSpace(target); target != "" {
		query = query.Where("target = ?", target)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("admin_name LIKE ? OR action LIKE ? OR target LIKE ? OR detail LIKE ? OR ip LIKE ?", like, like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

// GetAdminRolePermissions 获取管理员子角色与权限矩阵
func (s *AdminService) GetAdminRolePermissions(adminID uint) (*AdminRolePermissionsResponse, error) {
	db := s.GetDB()
	if db == nil {
		return nil, errors.New("数据库未初始化")
	}
	if err := models.SeedDefaultAdminRBAC(db); err != nil {
		return nil, err
	}

	var permissions []models.AdminPermission
	if err := db.Order("module ASC, code ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}

	var roles []models.AdminRole
	if err := db.Order("built_in DESC, id ASC").Find(&roles).Error; err != nil {
		return nil, err
	}

	var relations []models.AdminRolePermission
	if err := db.Find(&relations).Error; err != nil {
		return nil, err
	}
	rolePermissions := map[string][]string{}
	for _, relation := range relations {
		rolePermissions[relation.RoleKey] = append(rolePermissions[relation.RoleKey], relation.PermissionCode)
	}
	for key := range rolePermissions {
		sort.Strings(rolePermissions[key])
	}

	currentRoleKey := models.AdminRoleSuper
	var assignment models.AdminUserRole
	if err := db.Where("user_id = ?", adminID).First(&assignment).Error; err == nil && assignment.RoleKey != "" {
		currentRoleKey = assignment.RoleKey
	} else if err != nil && err != gorm.ErrRecordNotFound && !models.IsMissingAdminRBACTableError(err) {
		return nil, err
	}

	response := &AdminRolePermissionsResponse{
		Roles:          make([]AdminRoleDTO, 0, len(roles)),
		Permissions:    make([]AdminPermissionDTO, 0, len(permissions)),
		Admins:         []AdminRoleAssignmentDTO{},
		CurrentRoleKey: currentRoleKey,
	}
	roleNameMap := map[string]string{}
	for _, permission := range permissions {
		response.Permissions = append(response.Permissions, AdminPermissionDTO{
			Code:        permission.Code,
			Name:        permission.Name,
			Module:      permission.Module,
			Description: permission.Description,
		})
	}
	for _, role := range roles {
		roleNameMap[role.Key] = role.Name
		response.Roles = append(response.Roles, AdminRoleDTO{
			Key:         role.Key,
			Name:        role.Name,
			Description: role.Description,
			BuiltIn:     role.BuiltIn,
			Enabled:     role.Enabled,
			Permissions: rolePermissions[role.Key],
		})
	}

	var adminUsers []models.User
	if err := db.Where("role = ?", models.RoleAdmin).Order("created_at DESC").Find(&adminUsers).Error; err != nil {
		return nil, err
	}
	if len(adminUsers) > 0 {
		userIDs := make([]uint, 0, len(adminUsers))
		for _, user := range adminUsers {
			userIDs = append(userIDs, user.ID)
		}
		var assignments []models.AdminUserRole
		if err := db.Where("user_id IN ?", userIDs).Find(&assignments).Error; err != nil {
			return nil, err
		}
		assignmentMap := map[uint]models.AdminUserRole{}
		for _, assignment := range assignments {
			assignmentMap[assignment.UserID] = assignment
		}
		for _, user := range adminUsers {
			roleKey := models.AdminRoleSuper
			var assignedBy uint
			var assignedAt *time.Time
			if assignment, ok := assignmentMap[user.ID]; ok && assignment.RoleKey != "" {
				roleKey = assignment.RoleKey
				assignedBy = assignment.AssignedBy
				assignedAt = &assignment.UpdatedAt
			}
			response.Admins = append(response.Admins, AdminRoleAssignmentDTO{
				UserID:     user.ID,
				Phone:      user.Phone,
				Nickname:   user.Nickname,
				Status:     string(user.Status),
				RoleKey:    roleKey,
				RoleName:   roleNameMap[roleKey],
				AssignedBy: assignedBy,
				AssignedAt: assignedAt,
			})
		}
	}
	return response, nil
}

var adminRoleKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{2,39}$`)

// CreateAdminRole 创建自定义管理员子角色。
func (s *AdminService) CreateAdminRole(req AdminRoleMutationRequest) (*AdminRoleDTO, error) {
	db := s.GetDB()
	if db == nil {
		return nil, errors.New("数据库未初始化")
	}
	if err := models.SeedDefaultAdminRBAC(db); err != nil {
		return nil, err
	}
	roleKey, name, description, permissions, err := s.normalizeAdminRoleMutation(req, true)
	if err != nil {
		return nil, err
	}

	var count int64
	if err := db.Model(&models.AdminRole{}).Where("key = ?", roleKey).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("子角色标识已存在")
	}

	role := models.AdminRole{Key: roleKey, Name: name, Description: description, BuiltIn: false, Enabled: true}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&role).Error; err != nil {
			return err
		}
		return replaceAdminRolePermissions(tx, roleKey, permissions)
	}); err != nil {
		return nil, err
	}
	return &AdminRoleDTO{Key: role.Key, Name: role.Name, Description: role.Description, BuiltIn: role.BuiltIn, Enabled: role.Enabled, Permissions: permissions}, nil
}

// UpdateAdminRole 更新自定义管理员子角色资料和权限。
func (s *AdminService) UpdateAdminRole(roleKey string, req AdminRoleMutationRequest) (*AdminRoleDTO, error) {
	db := s.GetDB()
	if db == nil {
		return nil, errors.New("数据库未初始化")
	}
	if err := models.SeedDefaultAdminRBAC(db); err != nil {
		return nil, err
	}
	roleKey = strings.TrimSpace(roleKey)
	if roleKey == "" {
		return nil, errors.New("子角色标识不能为空")
	}

	var role models.AdminRole
	if err := db.Where("key = ?", roleKey).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("子角色不存在")
		}
		return nil, err
	}
	if role.BuiltIn {
		return nil, errors.New("内置子角色不允许编辑")
	}

	_, name, description, permissions, err := s.normalizeAdminRoleMutation(req, false)
	if err != nil {
		return nil, err
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.AdminRole{}).Where("key = ?", roleKey).Updates(map[string]interface{}{
			"name":        name,
			"description": description,
		}).Error; err != nil {
			return err
		}
		return replaceAdminRolePermissions(tx, roleKey, permissions)
	}); err != nil {
		return nil, err
	}

	return &AdminRoleDTO{Key: roleKey, Name: name, Description: description, BuiltIn: false, Enabled: role.Enabled, Permissions: permissions}, nil
}

// UpdateAdminRoleStatus 启用/禁用自定义管理员子角色。
func (s *AdminService) UpdateAdminRoleStatus(roleKey string, enabled bool) (*AdminRoleDTO, error) {
	db := s.GetDB()
	if db == nil {
		return nil, errors.New("数据库未初始化")
	}
	if err := models.SeedDefaultAdminRBAC(db); err != nil {
		return nil, err
	}
	roleKey = strings.TrimSpace(roleKey)
	if roleKey == "" {
		return nil, errors.New("子角色标识不能为空")
	}

	var role models.AdminRole
	if err := db.Where("key = ?", roleKey).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("子角色不存在")
		}
		return nil, err
	}
	if role.BuiltIn {
		return nil, errors.New("内置子角色不允许启用或禁用")
	}
	if !enabled {
		var count int64
		if err := db.Model(&models.AdminUserRole{}).Where("role_key = ?", roleKey).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("该子角色已有管理员使用，不能禁用")
		}
	}
	if err := db.Model(&models.AdminRole{}).Where("key = ?", roleKey).Update("enabled", enabled).Error; err != nil {
		return nil, err
	}
	role.Enabled = enabled
	permissions, err := s.getAdminRolePermissions(roleKey)
	if err != nil {
		return nil, err
	}
	return &AdminRoleDTO{Key: role.Key, Name: role.Name, Description: role.Description, BuiltIn: role.BuiltIn, Enabled: role.Enabled, Permissions: permissions}, nil
}

// BatchAssignAdminRole 批量给管理员账号分配子角色。
func (s *AdminService) BatchAssignAdminRole(targetUserIDs []uint, roleKey string, assignedBy uint) ([]AdminRoleAssignmentResult, error) {
	db := s.GetDB()
	if db == nil {
		return nil, errors.New("数据库未初始化")
	}
	if len(targetUserIDs) == 0 {
		return nil, errors.New("请选择要授权的管理员账号")
	}
	if len(targetUserIDs) > 100 {
		return nil, errors.New("单次最多批量授权100个管理员")
	}
	roleKey = strings.TrimSpace(roleKey)
	if roleKey == "" {
		return nil, errors.New("子角色不能为空")
	}
	if err := models.SeedDefaultAdminRBAC(db); err != nil {
		return nil, err
	}
	var role models.AdminRole
	if err := db.Where("key = ? AND enabled = ?", roleKey, true).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("管理员子角色不存在或已禁用")
		}
		return nil, err
	}

	seen := map[uint]bool{}
	cleanIDs := []uint{}
	for _, targetUserID := range targetUserIDs {
		if targetUserID == 0 || seen[targetUserID] {
			continue
		}
		seen[targetUserID] = true
		cleanIDs = append(cleanIDs, targetUserID)
	}
	if len(cleanIDs) == 0 {
		return nil, errors.New("请选择有效的管理员账号")
	}

	var users []models.User
	if err := db.Where("id IN ?", cleanIDs).Find(&users).Error; err != nil {
		return nil, err
	}
	if len(users) != len(cleanIDs) {
		return nil, errors.New("包含不存在的管理员账号")
	}
	for _, user := range users {
		if user.Role != models.RoleAdmin {
			return nil, errors.New("只能给管理员账号分配管理员子角色")
		}
	}

	results := []AdminRoleAssignmentResult{}
	for _, targetUserID := range cleanIDs {
		result, err := s.AssignAdminRole(targetUserID, roleKey, assignedBy)
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	}
	return results, nil
}

func (s *AdminService) normalizeAdminRoleMutation(req AdminRoleMutationRequest, requireKey bool) (string, string, string, []string, error) {
	roleKey := strings.TrimSpace(req.Key)
	if requireKey {
		if roleKey == "" {
			return "", "", "", nil, errors.New("子角色标识不能为空")
		}
		if !adminRoleKeyPattern.MatchString(roleKey) {
			return "", "", "", nil, errors.New("子角色标识只能包含小写字母、数字和下划线，且以字母开头，长度3-40")
		}
		if roleKey == models.AdminRoleSuper {
			return "", "", "", nil, errors.New("不能使用超级管理员内置标识")
		}
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return "", "", "", nil, errors.New("子角色名称不能为空")
	}
	description := strings.TrimSpace(req.Description)
	permissions, err := s.validateAdminPermissionCodes(req.Permissions)
	if err != nil {
		return "", "", "", nil, err
	}
	return roleKey, name, description, permissions, nil
}

func (s *AdminService) validateAdminPermissionCodes(permissionCodes []string) ([]string, error) {
	db := s.GetDB()
	if db == nil {
		return nil, errors.New("数据库未初始化")
	}
	normalized := make([]string, 0, len(permissionCodes))
	seen := map[string]bool{}
	for _, code := range permissionCodes {
		code = strings.TrimSpace(code)
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true
		normalized = append(normalized, code)
	}
	if len(normalized) == 0 {
		return nil, errors.New("至少选择一个权限点")
	}
	var count int64
	if err := db.Model(&models.AdminPermission{}).Where("code IN ?", normalized).Count(&count).Error; err != nil {
		return nil, err
	}
	if int(count) != len(normalized) {
		return nil, errors.New("包含不存在的权限点")
	}
	sort.Strings(normalized)
	return normalized, nil
}

func replaceAdminRolePermissions(tx *gorm.DB, roleKey string, permissionCodes []string) error {
	if err := tx.Where("role_key = ?", roleKey).Delete(&models.AdminRolePermission{}).Error; err != nil {
		return err
	}
	for _, permissionCode := range permissionCodes {
		relation := models.AdminRolePermission{RoleKey: roleKey, PermissionCode: permissionCode}
		if err := tx.Create(&relation).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *AdminService) getAdminRolePermissions(roleKey string) ([]string, error) {
	db := s.GetDB()
	if db == nil {
		return nil, errors.New("数据库未初始化")
	}
	var relations []models.AdminRolePermission
	if err := db.Where("role_key = ?", roleKey).Order("permission_code ASC").Find(&relations).Error; err != nil {
		return nil, err
	}
	permissions := make([]string, 0, len(relations))
	for _, relation := range relations {
		permissions = append(permissions, relation.PermissionCode)
	}
	return permissions, nil
}

// AssignAdminRole 给管理员账号分配子角色
func (s *AdminService) AssignAdminRole(targetUserID uint, roleKey string, assignedBy uint) (*AdminRoleAssignmentResult, error) {
	db := s.GetDB()
	if db == nil {
		return nil, errors.New("数据库未初始化")
	}
	roleKey = strings.TrimSpace(roleKey)
	if roleKey == "" {
		return nil, errors.New("子角色不能为空")
	}
	if err := models.SeedDefaultAdminRBAC(db); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(targetUserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}
	if user.Role != models.RoleAdmin {
		return nil, errors.New("只能给管理员账号分配管理员子角色")
	}

	var role models.AdminRole
	if err := db.Where("key = ? AND enabled = ?", roleKey, true).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("管理员子角色不存在或已禁用")
		}
		return nil, err
	}

	oldRoleKey := models.AdminRoleSuper
	var existing models.AdminUserRole
	if err := db.Where("user_id = ?", targetUserID).First(&existing).Error; err == nil && existing.RoleKey != "" {
		oldRoleKey = existing.RoleKey
	} else if err != nil && err != gorm.ErrRecordNotFound && !models.IsMissingAdminRBACTableError(err) {
		return nil, err
	}

	assignment := models.AdminUserRole{UserID: targetUserID, RoleKey: roleKey, AssignedBy: assignedBy}
	if err := db.Where("user_id = ?", targetUserID).Assign(assignment).FirstOrCreate(&assignment).Error; err != nil {
		return nil, err
	}

	roleNameMap, err := s.getAdminRoleNameMap()
	if err != nil {
		return nil, err
	}
	return &AdminRoleAssignmentResult{
		TargetUserID:    user.ID,
		TargetPhone:     user.Phone,
		TargetNickname:  user.Nickname,
		OldRoleKey:      oldRoleKey,
		OldRoleName:     roleNameMap[oldRoleKey],
		NewRoleKey:      role.Key,
		NewRoleName:     role.Name,
		ChangedByUserID: assignedBy,
	}, nil
}

func (s *AdminService) getAdminRoleNameMap() (map[string]string, error) {
	db := s.GetDB()
	if db == nil {
		return nil, errors.New("数据库未初始化")
	}
	var roles []models.AdminRole
	if err := db.Find(&roles).Error; err != nil {
		return nil, err
	}
	result := map[string]string{}
	for _, role := range roles {
		result[role.Key] = role.Name
	}
	if result[models.AdminRoleSuper] == "" {
		result[models.AdminRoleSuper] = "超级管理员"
	}
	return result, nil
}

// GetAdminRoleAssignmentHistory 获取管理员子角色授权历史。
func (s *AdminService) GetAdminRoleAssignmentHistory(page, pageSize int, targetUserID uint) (*AdminRoleAssignmentHistoryResponse, error) {
	db := s.GetDB()
	if db == nil {
		return nil, errors.New("数据库未初始化")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := db.Model(&models.AdminOperationLog{}).
		Where("action = ? AND target = ?", "assign_admin_role", "admin_user_role")
	if targetUserID > 0 {
		query = query.Where("target_id = ?", targetUserID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var logs []models.AdminOperationLog
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, err
	}

	items := make([]AdminRoleAssignmentHistoryDTO, 0, len(logs))
	for _, logItem := range logs {
		items = append(items, parseAdminRoleAssignmentLog(logItem))
	}
	return &AdminRoleAssignmentHistoryResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func parseAdminRoleAssignmentLog(logItem models.AdminOperationLog) AdminRoleAssignmentHistoryDTO {
	item := AdminRoleAssignmentHistoryDTO{
		ID:           logItem.ID,
		AdminID:      logItem.AdminID,
		AdminName:    logItem.AdminName,
		TargetUserID: logItem.TargetID,
		Detail:       logItem.Detail,
		IP:           logItem.IP,
		CreatedAt:    logItem.CreatedAt,
	}

	var detail struct {
		TargetUserID   uint   `json:"target_user_id"`
		TargetPhone    string `json:"target_phone"`
		TargetNickname string `json:"target_nickname"`
		OldRoleKey     string `json:"old_role_key"`
		OldRoleName    string `json:"old_role_name"`
		NewRoleKey     string `json:"new_role_key"`
		NewRoleName    string `json:"new_role_name"`
	}
	if err := json.Unmarshal([]byte(logItem.Detail), &detail); err == nil {
		if detail.TargetUserID > 0 {
			item.TargetUserID = detail.TargetUserID
		}
		item.TargetPhone = detail.TargetPhone
		item.TargetNickname = detail.TargetNickname
		item.OldRoleKey = detail.OldRoleKey
		item.OldRoleName = detail.OldRoleName
		item.NewRoleKey = detail.NewRoleKey
		item.NewRoleName = detail.NewRoleName
	}
	return item
}

// GetExceptions 获取异常管控列表，当前聚合举报与失败登录两类高风险事件
func (s *AdminService) GetExceptions(page, pageSize int, status, source, severity string) (*AdminExceptionsResponse, error) {
	db := s.GetDB()
	if db == nil {
		return nil, errors.New("数据库未初始化")
	}

	items := []AdminExceptionItem{}

	if source == "" || source == "content_report" {
		var reports []models.ContentReport
		if err := db.Order("created_at DESC").Limit(200).Find(&reports).Error; err != nil {
			return nil, err
		}
		for _, report := range reports {
			itemStatus := "open"
			if report.Status == models.ContentReportStatusResolved || report.Status == models.ContentReportStatusRejected {
				itemStatus = "closed"
			}
			itemSeverity := "medium"
			if report.TargetType == models.ContentReportTypeUser || report.TargetType == models.ContentReportTypeMessage {
				itemSeverity = "high"
			}
			items = append(items, AdminExceptionItem{
				ID:         "content_report:" + uintToStringForAdmin(report.ID),
				Source:     "content_report",
				Severity:   itemSeverity,
				Status:     itemStatus,
				Title:      "内容举报待处理",
				Subject:    string(report.TargetType),
				Detail:     strings.TrimSpace(report.Reason + " " + report.Detail),
				RefID:      report.ID,
				OccurredAt: report.CreatedAt,
			})
		}
	}

	if source == "" || source == "login_failure" {
		var logs []models.LoginLog
		if err := db.Where("status = ?", "failed").Order("created_at DESC").Limit(200).Find(&logs).Error; err != nil {
			return nil, err
		}
		for _, item := range logs {
			items = append(items, AdminExceptionItem{
				ID:         "login_failure:" + uintToStringForAdmin(item.ID),
				Source:     "login_failure",
				Severity:   "medium",
				Status:     "open",
				Title:      "登录失败",
				Subject:    item.Phone,
				Detail:     item.FailReason,
				RefID:      item.ID,
				OccurredAt: item.CreatedAt,
			})
		}
	}

	filtered := make([]AdminExceptionItem, 0, len(items))
	stats := map[string]int{"total": 0, "open": 0, "closed": 0, "high": 0, "medium": 0, "low": 0}
	for _, item := range items {
		stats["total"]++
		stats[item.Status]++
		stats[item.Severity]++
		if status != "" && item.Status != status {
			continue
		}
		if severity != "" && item.Severity != severity {
			continue
		}
		filtered = append(filtered, item)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].OccurredAt.After(filtered[j].OccurredAt)
	})

	total := len(filtered)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return &AdminExceptionsResponse{
		List:     filtered[start:end],
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Stats:    stats,
	}, nil
}

func reportVersionAnalysisIDForService(id uint) *uint {
	if id == 0 {
		return nil
	}
	value := id
	return &value
}

func uintToStringForAdmin(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

// SetNotificationService 注入通知服务
func (s *AdminService) SetNotificationService(notificationService *NotificationService) {
	s.notificationService = notificationService
}

// ========== Dashboard Stats ==========

// DashboardStats 数据看板统计数据
type DashboardStats struct {
	TotalUsers            int64   `json:"total_users"`
	TotalOrders           int64   `json:"total_orders"`
	TotalReports          int64   `json:"total_reports"`
	TotalRevenue          float64 `json:"total_revenue"`
	TodayNewUsers         int64   `json:"today_new_users"`
	TodayOrders           int64   `json:"today_orders"`
	TodayRevenue          float64 `json:"today_revenue"`
	PendingApplications   int64   `json:"pending_applications"`
	PendingReports        int64   `json:"pending_reports"`
	PendingContentReports int64   `json:"pending_content_reports"`
}

// GetDashboardStats 获取数据看板统计数据
func (s *AdminService) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	totalUsers, err := s.userRepo.Count()
	if err != nil {
		return nil, err
	}
	stats.TotalUsers = totalUsers

	today := time.Now().Format("2006-01-02")
	todayUsers, err := s.userRepo.CountByDate(today)
	if err != nil {
		return nil, err
	}
	stats.TodayNewUsers = todayUsers

	orderStats, err := s.orderRepo.GetStatistics()
	if err != nil {
		return nil, err
	}
	stats.TotalOrders = orderStats.TotalCount
	stats.TotalRevenue = orderStats.TotalRevenue
	stats.TodayOrders = orderStats.TodayCount
	stats.TodayRevenue = orderStats.TodayRevenue

	reportStats, err := s.reportRepo.GetStatistics()
	if err != nil {
		return nil, err
	}
	stats.TotalReports = reportStats.TotalCount
	stats.PendingReports = reportStats.PendingCount

	if s.applicationRepo != nil {
		pendingApps, err := s.applicationRepo.CountByStatus(models.ApplicationStatusPending)
		if err == nil {
			stats.PendingApplications = pendingApps
		}
	}

	if s.contentReportRepo != nil {
		pendingCR, _ := s.contentReportRepo.CountByStatus(models.ContentReportStatusPending)
		stats.PendingContentReports = pendingCR
	}

	return stats, nil
}

// GrowthData 增长数据
type GrowthData struct {
	Date    string  `json:"date"`
	Users   int64   `json:"users"`
	Orders  int64   `json:"orders"`
	Revenue float64 `json:"revenue"`
}

// GetGrowthData 获取增长数据
func (s *AdminService) GetGrowthData(days int) ([]GrowthData, error) {
	var result []GrowthData

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")

		userCount, _ := s.userRepo.CountByDate(date)
		orderCount, orderRevenue, _ := s.orderRepo.GetStatisticsByDate(date)

		result = append(result, GrowthData{
			Date:    date,
			Users:   userCount,
			Orders:  orderCount,
			Revenue: orderRevenue,
		})
	}

	return result, nil
}

// GetFunnelData 获取漏斗数据
func (s *AdminService) GetFunnelData() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	var totalVisitors int64 = 0
	result["visitors"] = totalVisitors

	var registrations int64
	s.userRepo.CountByRole("", &registrations)
	result["registrations"] = registrations

	var orders int64
	s.orderRepo.GetTotalCount(&orders)
	result["orders"] = orders

	var payments int64
	s.orderRepo.GetPaidCount(&payments)
	result["payments"] = payments

	var completed int64
	s.orderRepo.GetCompletedCount(&completed)
	result["completed"] = completed

	return result, nil
}

// GetRetentionData 获取留存数据
func (s *AdminService) GetRetentionData(days int) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")

		newUsers, _ := s.userRepo.CountByDate(date)
		activeUsers, _ := s.userRepo.CountActiveByDate(date)

		result = append(result, map[string]interface{}{
			"date":         date,
			"new_users":    newUsers,
			"active_users": activeUsers,
			"retention":    0,
		})
	}

	return result, nil
}

// GetTopData 获取排行榜数据
func (s *AdminService) GetTopData() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	topPlayers, _ := s.userRepo.GetTopByRole("player", 10)
	result["top_players"] = topPlayers

	topAnalysts, _ := s.analystRepo.GetTopByOrders(10)
	result["top_analysts"] = topAnalysts

	topClubs, _ := s.userRepo.GetTopByRole("club", 10)
	result["top_clubs"] = topClubs

	return result, nil
}

// ========== User Management ==========

// GetUserList 获取用户列表
func (s *AdminService) GetUserList(page, pageSize int) ([]models.User, int64, error) {
	return s.userRepo.FindAll(page, pageSize)
}

// UpdateUserStatus 更新用户状态
func (s *AdminService) UpdateUserStatus(userID uint, status string) error {
	if !isValidAdminUserStatus(status) {
		return errors.New("无效的用户状态")
	}
	return s.userRepo.UpdateStatus(userID, status)
}

// UpdateUser 更新用户基础资料
func (s *AdminService) UpdateUser(userID uint, updates map[string]interface{}) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	if roleValue, ok := updates["role"]; ok && !isValidAdminUserRole(roleValue.(string)) {
		return nil, errors.New("无效的用户角色")
	}
	if currentRoleValue, ok := updates["current_role"]; ok && currentRoleValue.(string) != "" && !isValidAdminUserRole(currentRoleValue.(string)) {
		return nil, errors.New("无效的当前角色")
	}
	if statusValue, ok := updates["status"]; ok && !isValidAdminUserStatus(statusValue.(string)) {
		return nil, errors.New("无效的用户状态")
	}

	if phoneValue, ok := updates["phone"]; ok {
		phone := strings.TrimSpace(phoneValue.(string))
		if phone == "" {
			return nil, errors.New("手机号不能为空")
		}
		if phone != user.Phone {
			existing, err := s.userRepo.FindByPhone(phone)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.ID != userID {
				return nil, errors.New("手机号已被其他用户使用")
			}
		}
		updates["phone"] = phone
	}

	if len(updates) == 0 {
		return user, nil
	}
	if err := s.userRepo.Update(userID, updates); err != nil {
		return nil, err
	}
	return s.userRepo.FindByID(userID)
}

// DeleteUser 删除用户
func (s *AdminService) DeleteUser(userID uint) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if user.Role == "admin" {
		return errors.New("不能删除管理员账号")
	}
	return s.userRepo.Delete(userID)
}

func isValidAdminUserRole(role string) bool {
	switch models.UserRole(role) {
	case models.RoleUser, models.RoleAnalyst, models.RoleAdmin, models.RoleClub, models.RoleCoach, models.RoleScout:
		return true
	default:
		return false
	}
}

func isValidAdminUserStatus(status string) bool {
	switch models.UserStatus(status) {
	case models.StatusActive, models.StatusInactive, models.StatusBanned, models.StatusPending:
		return true
	default:
		return false
	}
}

// ========== Report Management ==========

// GetPendingReports 获取待审核报告列表
func (s *AdminService) GetPendingReports(page, pageSize int) ([]models.Report, int64, error) {
	reports, total, err := s.reportRepo.FindByStatus(models.ReportStatusProcessing, page, pageSize)
	if err != nil {
		log.Printf("[AdminService] get pending reports failed: %v", err)
	}
	return reports, total, err
}

// ReviewReport 审核报告
func (s *AdminService) ReviewReport(reportID uint, status models.ReportStatus, remark string, adminID uint) error {
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return err
	}
	if report == nil {
		return errors.New("报告不存在")
	}

	order, err := s.orderRepo.FindByID(report.OrderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	now := time.Now()
	err = s.orderRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"status": status,
		}
		if remark != "" || status == models.ReportStatusFailed {
			updates["review_remark"] = remark
		}
		if err := tx.Model(&models.Report{}).Where("id = ?", reportID).Updates(updates).Error; err != nil {
			return err
		}

		switch status {
		case models.ReportStatusCompleted:
			if err := tx.Model(&models.Order{}).Where("id = ?", report.OrderID).Updates(map[string]interface{}{
				"status":       models.OrderStatusCompleted,
				"report_id":    report.ID,
				"completed_at": &now,
			}).Error; err != nil {
				return err
			}
			if s.videoAnalysisRepo != nil {
				if err := tx.Model(&models.VideoAnalysis{}).Where("order_id = ?", report.OrderID).Updates(map[string]interface{}{
					"status":           models.AnalysisStatusCompleted,
					"ai_report_status": "confirmed",
				}).Error; err != nil {
					return err
				}
			}
			if err := s.createStatusHistory(tx, report.OrderID, order.Status, models.OrderStatusCompleted, adminID, "admin", "管理员审核通过报告"); err != nil {
				return err
			}
		case models.ReportStatusFailed:
			if s.videoAnalysisRepo != nil {
				if err := tx.Model(&models.VideoAnalysis{}).Where("order_id = ?", report.OrderID).Updates(map[string]interface{}{
					"status":           models.AnalysisStatusDraft,
					"ai_report_status": string(models.ReportVersionStatusAdminRejected),
				}).Error; err != nil {
					return err
				}
			}
		}
		versionStatus := models.ReportVersionStatusApproved
		if status == models.ReportStatusFailed {
			versionStatus = models.ReportVersionStatusAdminRejected
		} else if status == models.ReportStatusProcessing {
			versionStatus = models.ReportVersionStatusAnalystSubmitted
		}
		var analysis models.VideoAnalysis
		if tx.Migrator().HasTable(&models.VideoAnalysis{}) {
			_ = tx.Where("order_id = ?", report.OrderID).First(&analysis).Error
		}
		if status == models.ReportStatusFailed && analysis.ID != 0 {
			if err := models.NewAnalysisOperationEventRepository(s.GetDB()).CreateWithTx(tx, &models.AnalysisOperationEvent{
				OrderID:      report.OrderID,
				AnalysisID:   analysis.ID,
				AnalystID:    analysis.AnalystID,
				EventType:    "revision_received",
				Section:      "submit",
				AfterSummary: firstNonEmptyProgress(remark, "管理员驳回报告，需返工"),
				Metadata: operationMetadata(map[string]interface{}{
					"report_id": report.ID,
					"remark":    remark,
				}),
				CreatedAt: now,
			}); err != nil {
				log.Printf("[AnalysisOperationEvent] revision_received create failed: %v", err)
			}
		}
		adminUserID := adminID
		if err := models.CreateReportVersion(tx, &models.ReportVersion{
			ReportID:                report.ID,
			OrderID:                 report.OrderID,
			AnalysisID:              reportVersionAnalysisIDForService(analysis.ID),
			SourceType:              models.ReportVersionSourceAdminReview,
			Status:                  versionStatus,
			Content:                 report.Content,
			WordURL:                 report.AIReportURL,
			PDFURL:                  report.PdfURL,
			InputSnapshot:           analysis.AIReportInputSnapshot,
			TemplateVersion:         analysis.AIReportTemplateVersion,
			DocumentTemplateVersion: VideoAnalysisDocumentTemplateVersion,
			ReviewRemark:            remark,
			CreatedByUserID:         &adminUserID,
			CreatedByRole:           "admin",
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	if status == models.ReportStatusCompleted {
		s.notifyReportCompleted(report)
	} else if status == models.ReportStatusFailed {
		s.notifyAnalystReportRejected(report, remark)
	}
	return nil
}

// GetReportByID 获取报告详情
func (s *AdminService) GetReportByID(reportID uint) (*models.Report, error) {
	return s.reportRepo.FindByID(reportID)
}

// UpdateReportAIURL 更新报告的 AI 报告/视频 URL
func (s *AdminService) UpdateReportAIURL(reportID uint, reportURL, videoURL string) error {
	updates := map[string]interface{}{}
	if reportURL != "" {
		updates["ai_report_url"] = reportURL
	}
	if videoURL != "" {
		updates["ai_video_url"] = videoURL
	}
	if len(updates) == 0 {
		return nil
	}
	return s.reportRepo.Update(reportID, updates)
}

// ========== Order Management ==========

// GetAllOrders 获取所有订单
func (s *AdminService) GetAllOrders(page, pageSize int, status string) ([]models.Order, int64, error) {
	return s.orderRepo.FindAll(page, pageSize, status)
}

// CancelOrder 取消订单
func (s *AdminService) CancelOrder(orderID, adminID uint) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	allowedStatuses := []models.OrderStatus{
		models.OrderStatusPending,
		models.OrderStatusPaid,
		models.OrderStatusUploaded,
		models.OrderStatusAssigned,
	}
	canCancel := false
	for _, status := range allowedStatuses {
		if order.Status == status {
			canCancel = true
			break
		}
	}
	if !canCancel {
		return errors.New("该订单状态不允许取消")
	}

	return s.orderRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
			"status": models.OrderStatusCancelled,
		}).Error; err != nil {
			return err
		}
		if s.assignmentRepo != nil && order.AnalystID != nil {
			now := time.Now()
			if err := s.assignmentRepo.MarkLatestPendingWithTx(tx, orderID, *order.AnalystID, models.OrderAssignmentStatusExpired, "管理员取消订单", now); err != nil {
				return err
			}
		}
		return s.createStatusHistory(tx, orderID, order.Status, models.OrderStatusCancelled, adminID, "admin", "管理员取消订单")
	})
}

// AssignOrder 管理员派单给分析师
func (s *AdminService) AssignOrder(orderID, analystID, adminID uint) (*models.Order, error) {
	return s.AssignOrderWithRequest(orderID, AssignOrderRequest{AnalystID: analystID}, adminID)
}

// AssignOrderWithRequest 管理员派单给分析师，并保留派单截止时间和备注
func (s *AdminService) AssignOrderWithRequest(orderID uint, req AssignOrderRequest, adminID uint) (*models.Order, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != models.OrderStatusUploaded {
		return nil, errors.New("订单状态不允许派单")
	}

	analyst, err := s.analystRepo.FindByID(req.AnalystID)
	if err != nil {
		return nil, err
	}
	if analyst == nil {
		return nil, errors.New("分析师不存在")
	}
	if analyst.Status != models.AnalystStatusActive {
		return nil, errors.New("分析师暂不可用")
	}

	deadlineHours := 48
	if order.OrderType == "pro" {
		deadlineHours = 72
	}
	assignedAt := time.Now()
	deadline := assignedAt.Add(time.Duration(deadlineHours) * time.Hour)
	if req.Deadline != nil {
		if !req.Deadline.After(assignedAt) {
			return nil, errors.New("截止时间必须晚于当前时间")
		}
		deadline = *req.Deadline
	}
	dispatchReason := "管理员派发订单"
	if note := strings.TrimSpace(req.Note); note != "" {
		dispatchReason = dispatchReason + "；备注：" + note
	}

	updates := map[string]interface{}{
		"analyst_id":  req.AnalystID,
		"status":      models.OrderStatusAssigned,
		"assigned_at": assignedAt,
		"deadline":    deadline,
	}

	if err := s.orderRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Updates(updates).Error; err != nil {
			return err
		}
		if s.assignmentRepo != nil {
			assignment := &models.OrderAssignment{
				OrderID:    orderID,
				AnalystID:  req.AnalystID,
				AssignedBy: optionalActorID(adminID),
				AssignedAt: assignedAt,
				Status:     models.OrderAssignmentStatusPending,
			}
			if err := s.assignmentRepo.CreateWithTx(tx, assignment); err != nil {
				return err
			}
		}
		return s.createStatusHistory(tx, orderID, order.Status, models.OrderStatusAssigned, adminID, "admin", dispatchReason)
	}); err != nil {
		return nil, err
	}

	assignedOrder, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	s.notifyAnalystOrderAssigned(analyst, assignedOrder)
	return assignedOrder, nil
}

// GetAssignmentRecords 获取订单派发历史
func (s *AdminService) GetAssignmentRecords(page, pageSize int, status string) ([]models.OrderAssignment, int64, error) {
	if status != "" && !models.IsValidOrderAssignmentStatus(status) {
		return nil, 0, errors.New("无效的派发状态")
	}
	return s.assignmentRepo.FindAll(page, pageSize, status)
}

// GetOrderStatusHistory 获取订单状态流转历史
func (s *AdminService) GetOrderStatusHistory(orderID uint) ([]models.OrderStatusHistory, error) {
	return s.statusHistoryRepo.FindByOrderID(orderID)
}

func (s *AdminService) createStatusHistory(tx *gorm.DB, orderID uint, fromStatus, toStatus models.OrderStatus, actorID uint, actorRole, reason string) error {
	if s.statusHistoryRepo == nil || fromStatus == toStatus {
		return nil
	}
	return s.statusHistoryRepo.CreateWithTx(tx, &models.OrderStatusHistory{
		OrderID:    orderID,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		ActorID:    optionalActorID(actorID),
		ActorRole:  actorRole,
		Reason:     reason,
	})
}

func (s *AdminService) notifyAnalystOrderAssigned(analyst *models.Analyst, order *models.Order) {
	if s.notificationService == nil || analyst == nil || order == nil {
		return
	}
	if err := s.notificationService.NotifyAnalystOrderAssigned(analyst.UserID, order.ID, order.OrderNo, order.PlayerName); err != nil {
		log.Printf("[AdminService] notify analyst %d for order %d failed: %v", analyst.ID, order.ID, err)
	}
}

func (s *AdminService) notifyReportCompleted(report *models.Report) {
	if s.notificationService == nil || report == nil {
		return
	}
	if err := s.notificationService.NotifyReportCompleted(report.UserID, report.ID, report.PlayerName); err != nil {
		log.Printf("[AdminService] notify player %d for report %d failed: %v", report.UserID, report.ID, err)
	}
}

func (s *AdminService) notifyAnalystReportRejected(report *models.Report, remark string) {
	if s.notificationService == nil || s.analystRepo == nil || report == nil {
		return
	}
	analyst, err := s.analystRepo.FindByID(report.AnalystID)
	if err != nil || analyst == nil {
		if err != nil {
			log.Printf("[AdminService] query analyst %d for report notification failed: %v", report.AnalystID, err)
		}
		return
	}
	if err := s.notificationService.NotifyAnalystReportRejected(analyst.UserID, report.ID, report.PlayerName, remark); err != nil {
		log.Printf("[AdminService] notify analyst %d for rejected report %d failed: %v", analyst.ID, report.ID, err)
	}
}

func optionalActorID(actorID uint) *uint {
	if actorID == 0 {
		return nil
	}
	return &actorID
}

// GetOrderStats 获取订单统计
func (s *AdminService) GetOrderStats() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	stats, _ := s.orderRepo.GetStatistics()
	result["total_count"] = stats.TotalCount
	result["total_revenue"] = stats.TotalRevenue
	result["today_count"] = stats.TodayCount
	result["today_revenue"] = stats.TodayRevenue

	var statusCounts []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	s.orderRepo.GetStatusCounts(&statusCounts)
	result["status_counts"] = statusCounts

	return result, nil
}

// GetRevenueTrend 获取收入趋势
func (s *AdminService) GetRevenueTrend(days int) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		count, revenue, _ := s.orderRepo.GetStatisticsByDate(date)

		result = append(result, map[string]interface{}{
			"date":    date,
			"orders":  count,
			"revenue": revenue,
		})
	}

	return result, nil
}

// ========== Analyst Management ==========

// GetAnalystList 获取分析师列表
func (s *AdminService) GetAnalystList(page, pageSize int, status string) ([]models.User, int64, error) {
	return s.userRepo.FindByRole("analyst", page, pageSize, status)
}

// AuditAnalyst 审核分析师
func (s *AdminService) AuditAnalyst(analystID uint, status string, remark string) error {
	return nil
}

// UpdateAnalystStatus 更新分析师状态
func (s *AdminService) UpdateAnalystStatus(analystID uint, status string) error {
	return s.userRepo.UpdateStatus(analystID, status)
}

// GetAnalystIncomeStats 获取分析师收益统计
func (s *AdminService) GetAnalystIncomeStats(analystID uint) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	totalIncome, _ := s.orderRepo.GetAnalystTotalIncome(analystID)
	result["total_income"] = totalIncome

	monthIncome, _ := s.orderRepo.GetAnalystMonthIncome(analystID)
	result["month_income"] = monthIncome

	orderCount, _ := s.orderRepo.GetAnalystOrderCount(analystID)
	result["order_count"] = orderCount

	return result, nil
}

// GetSettlementList 获取待结算列表
func (s *AdminService) GetSettlementList(page, pageSize int) ([]models.Order, int64, error) {
	return s.orderRepo.FindCompletedUnsettled(page, pageSize)
}

// ProcessSettlement 处理结算
func (s *AdminService) ProcessSettlement(orderIDs []uint, adminID uint) error {
	for _, id := range orderIDs {
		if err := s.orderRepo.UpdateSettlement(id, adminID, time.Now()); err != nil {
			return err
		}
	}
	return nil
}

// ========== Content Report ==========

// GetContentReports 获取举报列表
func (s *AdminService) GetContentReports(page, pageSize int, status string) ([]models.ContentReport, int64, error) {
	return s.contentReportRepo.FindAll(page, pageSize, status)
}

// HandleContentReport 处理举报
func (s *AdminService) HandleContentReport(reportID uint, status models.ContentReportStatus, handlerID uint, handlerName, result string) error {
	return s.contentReportRepo.UpdateStatus(reportID, status, handlerID, handlerName, result)
}

// GetContentReportDetail 获取举报详情
func (s *AdminService) GetContentReportDetail(reportID uint) (*models.ContentReport, error) {
	return s.contentReportRepo.FindByID(reportID)
}

// ========== Sensitive Word ==========

// GetSensitiveWords 获取敏感词列表
func (s *AdminService) GetSensitiveWords(page, pageSize int, category string, enabled *bool) ([]models.SensitiveWord, int64, error) {
	return s.sensitiveWordRepo.FindAll(page, pageSize, category, enabled)
}

// CreateSensitiveWord 创建敏感词
func (s *AdminService) CreateSensitiveWord(word *models.SensitiveWord) error {
	return s.sensitiveWordRepo.Create(word)
}

// UpdateSensitiveWord 更新敏感词
func (s *AdminService) UpdateSensitiveWord(id uint, updates map[string]interface{}) error {
	return s.sensitiveWordRepo.Update(id, updates)
}

// DeleteSensitiveWord 删除敏感词
func (s *AdminService) DeleteSensitiveWord(id uint) error {
	return s.sensitiveWordRepo.Delete(id)
}

// CheckSensitiveWords 检查敏感词
func (s *AdminService) CheckSensitiveWords(text string) ([]string, error) {
	return s.sensitiveWordRepo.CheckText(text)
}

// ========== Platform Announcement ==========

// GetPlatformAnnouncements 获取平台公告列表
func (s *AdminService) GetPlatformAnnouncements(page, pageSize int, annType string, pinned *bool) ([]models.PlatformAnnouncement, int64, error) {
	return s.platformAnnRepo.FindAll(page, pageSize, annType, pinned)
}

// CreatePlatformAnnouncement 创建平台公告
func (s *AdminService) CreatePlatformAnnouncement(ann *models.PlatformAnnouncement) error {
	return s.platformAnnRepo.Create(ann)
}

// UpdatePlatformAnnouncement 更新平台公告
func (s *AdminService) UpdatePlatformAnnouncement(id uint, updates map[string]interface{}) error {
	return s.platformAnnRepo.Update(id, updates)
}

// DeletePlatformAnnouncement 删除平台公告
func (s *AdminService) DeletePlatformAnnouncement(id uint) error {
	return s.platformAnnRepo.Delete(id)
}

// ========== Banner ==========

// GetBanners 获取轮播图列表
func (s *AdminService) GetBanners(page, pageSize int, position string, enabled *bool) ([]models.Banner, int64, error) {
	return s.bannerRepo.FindAll(page, pageSize, position, enabled)
}

// CreateBanner 创建轮播图
func (s *AdminService) CreateBanner(banner *models.Banner) error {
	return s.bannerRepo.Create(banner)
}

// UpdateBanner 更新轮播图
func (s *AdminService) UpdateBanner(id uint, updates map[string]interface{}) error {
	return s.bannerRepo.Update(id, updates)
}

// DeleteBanner 删除轮播图
func (s *AdminService) DeleteBanner(id uint) error {
	return s.bannerRepo.Delete(id)
}

// ========== FAQ ==========

// GetFAQs 获取FAQ列表
func (s *AdminService) GetFAQs(page, pageSize int, category string, enabled *bool) ([]models.FAQ, int64, error) {
	return s.faqRepo.FindAll(page, pageSize, category, enabled)
}

// CreateFAQ 创建FAQ
func (s *AdminService) CreateFAQ(faq *models.FAQ) error {
	return s.faqRepo.Create(faq)
}

// UpdateFAQ 更新FAQ
func (s *AdminService) UpdateFAQ(id uint, updates map[string]interface{}) error {
	return s.faqRepo.Update(id, updates)
}

// DeleteFAQ 删除FAQ
func (s *AdminService) DeleteFAQ(id uint) error {
	return s.faqRepo.Delete(id)
}

// ========== Login Log ==========

// GetLoginLogs 获取登录日志
func (s *AdminService) GetLoginLogs(page, pageSize int, userID uint, status, startDate, endDate string) ([]models.LoginLog, int64, error) {
	return s.loginLogRepo.FindAll(page, pageSize, userID, status, startDate, endDate)
}

// GetLoginLogStats 获取登录日志统计
func (s *AdminService) GetLoginLogStats(days int) (map[string]interface{}, error) {
	return s.loginLogRepo.GetStatistics(days)
}

// CreateLoginLog 创建登录日志
func (s *AdminService) CreateLoginLog(log *models.LoginLog) error {
	return s.loginLogRepo.Create(log)
}

// ========== Admin Login ==========

// ========== Available Analysts for Dispatch ==========

// AvailableAnalyst 可派单的分析师（含工作负载）
type AvailableAnalyst struct {
	AnalystID       uint        `json:"analyst_id"`
	Analyst         models.User `json:"analyst"`
	MaxOrders       int         `json:"max_orders"`
	AcceptedOrders  int         `json:"accepted_orders"`
	CompletedOrders int         `json:"completed_orders"`
	WorkingHours    string      `json:"working_hours"`
	IsAvailable     bool        `json:"is_available"`
	TotalCompleted  int64       `json:"total_completed"`
	AvgRating       float64     `json:"avg_rating"`
	Specialties     []string    `json:"specialties"`
}

// GetAvailableAnalysts 获取可派单的分析师列表（含实时工作负载）
func (s *AdminService) GetAvailableAnalysts() ([]AvailableAnalyst, error) {
	// 查询所有活跃分析师（已预加载 User）
	analysts, _, err := s.analystRepo.FindAll(1, 100)
	if err != nil {
		return nil, err
	}

	var result []AvailableAnalyst
	for _, analyst := range analysts {
		// 统计当前进行中订单数（assigned + processing）
		var activeCount int64
		s.orderRepo.GetDB().Model(&models.Order{}).
			Where("analyst_id = ? AND status IN ?", analyst.ID, []models.OrderStatus{models.OrderStatusAssigned, models.OrderStatusProcessing}).
			Count(&activeCount)

		// 统计历史完成订单数
		var totalCompleted int64
		s.orderRepo.GetDB().Model(&models.Order{}).
			Where("analyst_id = ? AND status = ?", analyst.ID, models.OrderStatusCompleted).
			Count(&totalCompleted)

		maxOrders := 5 // 默认每日最大接单量
		specialties := []string{}
		if analyst.Specialty != "" {
			specialties = strings.Split(analyst.Specialty, ",")
		}

		result = append(result, AvailableAnalyst{
			AnalystID:       analyst.ID,
			Analyst:         analyst.User,
			MaxOrders:       maxOrders,
			AcceptedOrders:  int(activeCount),
			CompletedOrders: int(totalCompleted),
			WorkingHours:    "09:00-18:00",
			IsAvailable:     int(activeCount) < maxOrders,
			TotalCompleted:  totalCompleted,
			AvgRating:       analyst.Rating,
			Specialties:     specialties,
		})
	}

	return result, nil
}

// AdminLogin 管理员登录
func (s *AdminService) AdminLogin(username, password string) (string, *models.User, error) {
	admin, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}
	if admin == nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	if admin.Role != "admin" {
		return "", nil, errors.New("无权限访问")
	}
	if admin.Status != models.StatusActive {
		return "", nil, errors.New("账号未激活或已被禁用")
	}

	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password))
	if err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	token, err := utils.GenerateToken(admin.ID, admin.Phone, string(admin.Role))
	if err != nil {
		return "", nil, err
	}

	admin.Password = ""

	return token, admin, nil
}
