package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

type AccountRoleController struct {
	db *gorm.DB
}

func NewAccountRoleController(db *gorm.DB) *AccountRoleController {
	return &AccountRoleController{db: db}
}

func (ctrl *AccountRoleController) writeAdminOperationLog(c *gin.Context, admin *models.User, action, target string, targetID uint, detail string) {
	if ctrl == nil || ctrl.db == nil || admin == nil {
		return
	}
	adminName := admin.Nickname
	if strings.TrimSpace(adminName) == "" {
		adminName = admin.Name
	}
	if strings.TrimSpace(adminName) == "" {
		adminName = admin.Phone
	}
	if strings.TrimSpace(adminName) == "" {
		adminName = "管理员"
	}
	if err := ctrl.db.Create(&models.AdminOperationLog{
		ClubID:    0,
		AdminID:   admin.ID,
		AdminName: adminName,
		Action:    action,
		Target:    target,
		TargetID:  targetID,
		Detail:    detail,
		IP:        c.ClientIP(),
		CreatedAt: time.Now(),
	}).Error; err != nil {
		fmt.Printf("[AdminAudit] write role application audit log failed: %v\n", err)
	}
}

type accountRoleItem struct {
	Type         models.UserRole `json:"type"`
	Status       string          `json:"status"`
	Label        string          `json:"label"`
	CanApply     bool            `json:"can_apply"`
	DashboardURL string          `json:"dashboard_url,omitempty"`
	HomeURL      string          `json:"home_url,omitempty"`
	AppliedAt    *time.Time      `json:"applied_at,omitempty"`
	RejectReason string          `json:"reject_reason,omitempty"`
}

type applyRoleRequest struct {
	Role    models.UserRole        `json:"role" binding:"required"`
	Profile map[string]interface{} `json:"profile"`
	Source  string                 `json:"source"`
}

type reviewRoleApplicationRequest struct {
	Status string `json:"status" binding:"required"`
	Remark string `json:"remark"`
}

type adminRoleApplicationItem struct {
	ID           uint                   `json:"id"`
	UserID       uint                   `json:"user_id"`
	Role         models.UserRole        `json:"role"`
	RoleLabel    string                 `json:"role_label"`
	Status       string                 `json:"status"`
	Source       string                 `json:"source"`
	Profile      map[string]interface{} `json:"profile"`
	RejectReason string                 `json:"reject_reason"`
	ReviewedBy   uint                   `json:"reviewed_by"`
	ReviewedAt   *time.Time             `json:"reviewed_at"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	User         gin.H                  `json:"user"`
}

var accountRoleMeta = map[models.UserRole]struct {
	label        string
	dashboardURL string
	homeURL      func(uint) string
	canSelfApply bool
}{
	models.RoleUser:    {label: "球员", dashboardURL: "/user-dashboard", homeURL: func(id uint) string { return "/personal-homepage/" + uintToString(id) }},
	models.RoleAnalyst: {label: "分析师", dashboardURL: "/analyst/dashboard", homeURL: func(id uint) string { return "/analyst/" + uintToString(id) }, canSelfApply: true},
	models.RoleClub:    {label: "俱乐部", dashboardURL: "/club/dashboard", homeURL: func(uint) string { return "/club/dashboard?tab=home-preview" }, canSelfApply: true},
	models.RoleCoach:   {label: "教练", dashboardURL: "/coach/dashboard", homeURL: func(id uint) string { return "/coach/" + uintToString(id) }, canSelfApply: true},
	models.RoleScout:   {label: "球探", dashboardURL: "/scout/dashboard", homeURL: func(id uint) string { return "/scout/" + uintToString(id) }, canSelfApply: true},
	models.RoleAdmin:   {label: "管理员", dashboardURL: "/admin/dashboard", canSelfApply: false},
}

func uintToString(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

func (ctrl *AccountRoleController) ListRoles(c *gin.Context) {
	user := getContextUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": gin.H{"message": "未认证"}})
		return
	}

	activeRoles, err := ctrl.getActiveRoles(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "身份读取失败"}})
		return
	}

	var applications []models.RoleApplication
	if err := ctrl.db.Where("user_id = ?", user.ID).Order("created_at DESC").Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "申请记录读取失败"}})
		return
	}

	latestApplication := make(map[models.UserRole]models.RoleApplication)
	for _, app := range applications {
		if _, exists := latestApplication[app.Role]; !exists {
			latestApplication[app.Role] = app
		}
	}

	orderedRoles := []models.UserRole{models.RoleUser, models.RoleClub, models.RoleCoach, models.RoleAnalyst, models.RoleScout, models.RoleAdmin}
	items := make([]accountRoleItem, 0, len(orderedRoles))
	for _, role := range orderedRoles {
		meta := accountRoleMeta[role]
		status := "none"
		if activeRoles[role] {
			status = "active"
		} else if app, exists := latestApplication[role]; exists {
			status = string(app.Status)
		}

		item := accountRoleItem{
			Type:         role,
			Status:       status,
			Label:        meta.label,
			CanApply:     meta.canSelfApply && status != "active" && status != "pending" && status != "suspended",
			DashboardURL: meta.dashboardURL,
		}
		if meta.homeURL != nil {
			item.HomeURL = meta.homeURL(user.ID)
		}
		if app, exists := latestApplication[role]; exists {
			item.AppliedAt = &app.CreatedAt
			item.RejectReason = app.RejectReason
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"roles":        items,
			"current_role": user.CurrentRole,
		},
	})
}

func (ctrl *AccountRoleController) ApplyRole(c *gin.Context) {
	user := getContextUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": gin.H{"message": "未认证"}})
		return
	}

	var req applyRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}

	meta, ok := accountRoleMeta[req.Role]
	if !ok || !meta.canSelfApply {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "该身份暂不支持用户主动申请"}})
		return
	}

	activeRoles, err := ctrl.getActiveRoles(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "身份校验失败"}})
		return
	}
	if activeRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "您已拥有该身份"}})
		return
	}

	var pendingCount int64
	if err := ctrl.db.Model(&models.RoleApplication{}).
		Where("user_id = ? AND role = ? AND status = ?", user.ID, req.Role, models.RoleApplicationStatusPending).
		Count(&pendingCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "申请状态校验失败"}})
		return
	}
	if pendingCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "该身份已在审核中，请勿重复申请"}})
		return
	}

	profileBytes, err := json.Marshal(req.Profile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "申请资料格式错误"}})
		return
	}

	source := req.Source
	if source == "" {
		source = "self_apply"
	}

	application := models.RoleApplication{
		UserID:      user.ID,
		Role:        req.Role,
		Status:      models.RoleApplicationStatusPending,
		Source:      source,
		ProfileJSON: string(profileBytes),
	}
	if err := ctrl.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&application).Error; err != nil {
			return err
		}
		return models.UpsertUserRoleRecord(tx, models.UserRoleRecord{
			UserID:        user.ID,
			Role:          req.Role,
			Status:        string(models.RoleApplicationStatusPending),
			Source:        source,
			PublicVisible: true,
			RejectReason:  "",
		})
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "申请提交失败"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "申请已提交，等待审核",
		"data": gin.H{
			"application": application,
		},
	})
}

func (ctrl *AccountRoleController) ListAdminApplications(c *gin.Context) {
	status := c.DefaultQuery("status", string(models.RoleApplicationStatusPending))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 100
	}

	query := ctrl.db.Model(&models.RoleApplication{}).Preload("User")
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "申请统计失败"}})
		return
	}

	var applications []models.RoleApplication
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "申请列表读取失败"}})
		return
	}

	items := make([]adminRoleApplicationItem, 0, len(applications))
	for _, application := range applications {
		items = append(items, ctrl.toAdminRoleApplicationItem(application))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"list":     items,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

func (ctrl *AccountRoleController) ReviewRoleApplication(c *gin.Context) {
	admin := getContextUser(c)
	if admin == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": gin.H{"message": "未认证"}})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "申请ID格式错误"}})
		return
	}

	var req reviewRoleApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	if req.Status != string(models.RoleApplicationStatusApproved) && req.Status != string(models.RoleApplicationStatusRejected) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "审核状态不支持"}})
		return
	}

	var application models.RoleApplication
	if err := ctrl.db.Preload("User").First(&application, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "申请不存在"}})
		return
	}
	if application.Status != models.RoleApplicationStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "该申请已处理"}})
		return
	}

	reviewedAt := time.Now()
	remark := strings.TrimSpace(req.Remark)
	if req.Status == string(models.RoleApplicationStatusRejected) && remark == "" {
		remark = "未通过审核"
	}

	if err := ctrl.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"status":        req.Status,
			"reject_reason": "",
			"reviewed_by":   admin.ID,
			"reviewed_at":   &reviewedAt,
		}
		if req.Status == string(models.RoleApplicationStatusRejected) {
			updates["reject_reason"] = remark
		}
		if err := tx.Model(&models.RoleApplication{}).Where("id = ?", application.ID).Updates(updates).Error; err != nil {
			return err
		}

		if req.Status == string(models.RoleApplicationStatusRejected) {
			return models.UpsertUserRoleRecord(tx, models.UserRoleRecord{
				UserID:        application.UserID,
				Role:          application.Role,
				Status:        string(models.RoleApplicationStatusRejected),
				Source:        application.Source,
				PublicVisible: true,
				RejectReason:  remark,
			})
		}

		profileID, err := ctrl.ensureBusinessProfile(tx, &application)
		if err != nil {
			return err
		}
		return models.UpsertUserRoleRecord(tx, models.UserRoleRecord{
			UserID:        application.UserID,
			Role:          application.Role,
			Status:        "active",
			Source:        application.Source,
			ProfileID:     profileID,
			PublicVisible: true,
			RejectReason:  "",
			ApprovedAt:    &reviewedAt,
			ApprovedBy:    admin.ID,
		})
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "审核操作失败"}})
		return
	}

	ctrl.writeAdminOperationLog(c, admin, "review_role_application", "role_application", application.ID, fmt.Sprintf("审核用户ID=%d追加角色=%s，结果=%s", application.UserID, application.Role, req.Status))
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "审核已完成"})
}

func (ctrl *AccountRoleController) getActiveRoles(user *models.User) (map[models.UserRole]bool, error) {
	roles := map[models.UserRole]bool{}
	if user.Role != "" && user.Status == models.StatusActive {
		roles[user.Role] = true
	}

	var records []models.UserRoleRecord
	if err := ctrl.db.Where("user_id = ? AND status IN ?", user.ID, []string{"active", "approved"}).Find(&records).Error; err != nil {
		if !models.IsMissingUserRolesTableError(err) {
			return nil, err
		}
	}
	for _, record := range records {
		roles[record.Role] = true
	}

	var count int64
	if err := ctrl.db.Model(&models.Analyst{}).Where("user_id = ? AND status = ?", user.ID, models.AnalystStatusActive).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		roles[models.RoleAnalyst] = true
	}

	count = 0
	if err := ctrl.db.Model(&models.Scout{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		roles[models.RoleScout] = true
	}

	count = 0
	if err := ctrl.db.Model(&models.Club{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		roles[models.RoleClub] = true
	}

	count = 0
	if err := ctrl.db.Model(&models.ClubCoach{}).Where("user_id = ? AND status = ?", user.ID, models.ClubCoachStatusActive).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		roles[models.RoleCoach] = true
	}

	count = 0
	if err := ctrl.db.Model(&models.TeamCoach{}).Where("user_id = ? AND status = ?", user.ID, "active").Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		roles[models.RoleCoach] = true
	}

	return roles, nil
}

func (ctrl *AccountRoleController) toAdminRoleApplicationItem(application models.RoleApplication) adminRoleApplicationItem {
	profile := map[string]interface{}{}
	if application.ProfileJSON != "" {
		_ = json.Unmarshal([]byte(application.ProfileJSON), &profile)
	}

	meta := accountRoleMeta[application.Role]
	name := application.User.Name
	if name == "" {
		name = application.User.Nickname
	}

	return adminRoleApplicationItem{
		ID:           application.ID,
		UserID:       application.UserID,
		Role:         application.Role,
		RoleLabel:    meta.label,
		Status:       string(application.Status),
		Source:       application.Source,
		Profile:      profile,
		RejectReason: application.RejectReason,
		ReviewedBy:   application.ReviewedBy,
		ReviewedAt:   application.ReviewedAt,
		CreatedAt:    application.CreatedAt,
		UpdatedAt:    application.UpdatedAt,
		User: gin.H{
			"id":       application.User.ID,
			"name":     name,
			"phone":    application.User.Phone,
			"nickname": application.User.Nickname,
			"role":     application.User.Role,
		},
	}
}

func (ctrl *AccountRoleController) ensureBusinessProfile(tx *gorm.DB, application *models.RoleApplication) (uint, error) {
	profile := map[string]interface{}{}
	if application.ProfileJSON != "" {
		_ = json.Unmarshal([]byte(application.ProfileJSON), &profile)
	}

	var user models.User
	if err := tx.First(&user, application.UserID).Error; err != nil {
		return 0, err
	}

	switch application.Role {
	case models.RoleAnalyst:
		return ctrl.ensureAnalystProfile(tx, &user, profile)
	case models.RoleScout:
		return ctrl.ensureScoutProfile(tx, &user, profile)
	case models.RoleClub:
		return ctrl.ensureClubProfile(tx, &user, profile)
	case models.RoleCoach:
		return 0, nil
	default:
		return 0, nil
	}
}

func (ctrl *AccountRoleController) ensureAnalystProfile(tx *gorm.DB, user *models.User, profile map[string]interface{}) (uint, error) {
	name := firstRoleNonEmpty(user.Name, user.Nickname, stringFromProfile(profile, "name"), "用户"+uintToString(user.ID))
	bio := firstRoleNonEmpty(stringFromProfile(profile, "summary"), stringFromProfile(profile, "bio"))
	specialty := firstRoleNonEmpty(stringFromProfile(profile, "specialty"), stringFromProfile(profile, "qualification"))
	profession := stringFromProfile(profile, "profession")
	contactPhone := firstRoleNonEmpty(stringFromProfile(profile, "contact_phone"), user.Phone)
	contactEmail := stringFromProfile(profile, "contact_email")
	experience := intFromProfile(profile, "experience")

	var analyst models.Analyst
	err := tx.Where("user_id = ?", user.ID).First(&analyst).Error
	if err == gorm.ErrRecordNotFound {
		analyst = models.Analyst{
			UserID:       user.ID,
			Name:         name,
			Bio:          bio,
			Specialty:    specialty,
			Experience:   experience,
			Profession:   profession,
			ContactPhone: contactPhone,
			ContactEmail: contactEmail,
			Status:       models.AnalystStatusActive,
		}
		if err := tx.Create(&analyst).Error; err != nil {
			return 0, err
		}
		return analyst.ID, nil
	}
	if err != nil {
		return 0, err
	}

	updates := map[string]interface{}{"status": models.AnalystStatusActive}
	if analyst.Name == "" {
		updates["name"] = name
	}
	if analyst.Bio == "" && bio != "" {
		updates["bio"] = bio
	}
	if analyst.Specialty == "" && specialty != "" {
		updates["specialty"] = specialty
	}
	if analyst.Profession == "" && profession != "" {
		updates["profession"] = profession
	}
	if analyst.ContactPhone == "" && contactPhone != "" {
		updates["contact_phone"] = contactPhone
	}
	if analyst.ContactEmail == "" && contactEmail != "" {
		updates["contact_email"] = contactEmail
	}
	if analyst.Experience == 0 && experience > 0 {
		updates["experience"] = experience
	}
	return analyst.ID, tx.Model(&models.Analyst{}).Where("id = ?", analyst.ID).Updates(updates).Error
}

func (ctrl *AccountRoleController) ensureScoutProfile(tx *gorm.DB, user *models.User, profile map[string]interface{}) (uint, error) {
	var scout models.Scout
	err := tx.Where("user_id = ?", user.ID).First(&scout).Error
	if err == gorm.ErrRecordNotFound {
		scout = models.Scout{
			UserID:              user.ID,
			ScoutingExperience:  stringFromProfile(profile, "experience"),
			Specialties:         firstRoleNonEmpty(stringFromProfile(profile, "specialty"), stringFromProfile(profile, "qualification")),
			CurrentOrganization: stringFromProfile(profile, "organization"),
			Bio:                 firstRoleNonEmpty(stringFromProfile(profile, "summary"), stringFromProfile(profile, "bio")),
			Verified:            true,
		}
		if err := tx.Create(&scout).Error; err != nil {
			return 0, err
		}
		return scout.ID, nil
	}
	if err != nil {
		return 0, err
	}
	if !scout.Verified {
		if err := tx.Model(&models.Scout{}).Where("id = ?", scout.ID).Update("verified", true).Error; err != nil {
			return 0, err
		}
	}
	return scout.ID, nil
}

func (ctrl *AccountRoleController) ensureClubProfile(tx *gorm.DB, user *models.User, profile map[string]interface{}) (uint, error) {
	var club models.Club
	err := tx.Where("user_id = ?", user.ID).First(&club).Error
	if err == gorm.ErrRecordNotFound {
		clubName := firstRoleNonEmpty(stringFromProfile(profile, "club_name"), stringFromProfile(profile, "organization"), user.Club, user.Name, user.Nickname, "用户"+uintToString(user.ID)+"的俱乐部")
		club = models.Club{
			UserID:       user.ID,
			Name:         clubName,
			Description:  firstRoleNonEmpty(stringFromProfile(profile, "summary"), stringFromProfile(profile, "bio")),
			ContactName:  firstRoleNonEmpty(user.Name, user.Nickname),
			ContactPhone: firstRoleNonEmpty(stringFromProfile(profile, "contact_phone"), user.Phone),
			Province:     user.Province,
			City:         user.City,
		}
		if err := tx.Create(&club).Error; err != nil {
			return 0, err
		}
		return club.ID, nil
	}
	if err != nil {
		return 0, err
	}
	return club.ID, nil
}

func stringFromProfile(profile map[string]interface{}, key string) string {
	value, ok := profile[key]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func intFromProfile(profile map[string]interface{}, key string) int {
	value := stringFromProfile(profile, key)
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func firstRoleNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func getContextUser(c *gin.Context) *models.User {
	userValue, exists := c.Get("user")
	if !exists {
		return nil
	}
	user, ok := userValue.(*models.User)
	if !ok || user == nil || middleware.GetUserID(c) == 0 {
		return nil
	}
	return user
}
