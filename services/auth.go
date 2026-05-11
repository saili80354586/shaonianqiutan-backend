package services

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
)

// AuthService 认证服务
type AuthService struct {
	userRepo          *models.UserRepository
	analystRepo       *models.AnalystRepository
	orderRepo         *models.OrderRepository
	assignmentRepo    *models.OrderAssignmentRepository
	statusHistoryRepo *models.OrderStatusHistoryRepository
	smsService        *SmsService
	db                *gorm.DB
}

var ErrAccountNotActive = errors.New("账号未激活或已被禁用")

func NewAuthService(
	userRepo *models.UserRepository,
	analystRepo *models.AnalystRepository,
	orderRepo *models.OrderRepository,
	assignmentRepo *models.OrderAssignmentRepository,
	statusHistoryRepo *models.OrderStatusHistoryRepository,
	smsService *SmsService,
	db *gorm.DB,
) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
		analystRepo:       analystRepo,
		orderRepo:         orderRepo,
		assignmentRepo:    assignmentRepo,
		statusHistoryRepo: statusHistoryRepo,
		smsService:        smsService,
		db:                db,
	}
}

func (s *AuthService) getActiveUserRoles(user *models.User) ([]models.UserRole, error) {
	if user == nil || user.Status != models.StatusActive {
		return nil, nil
	}

	roles := make([]models.UserRole, 0, 4)
	seen := make(map[models.UserRole]bool)
	addRole := func(role models.UserRole) {
		if role == "" || seen[role] {
			return
		}
		seen[role] = true
		roles = append(roles, role)
	}

	addRole(user.Role)

	var records []models.UserRoleRecord
	if err := s.db.Where("user_id = ? AND status IN ?", user.ID, []string{"active", "approved"}).Find(&records).Error; err != nil {
		if !models.IsMissingUserRolesTableError(err) {
			return nil, err
		}
	}
	for _, record := range records {
		addRole(record.Role)
	}

	var count int64
	if err := s.db.Model(&models.Analyst{}).Where("user_id = ? AND status = ?", user.ID, models.AnalystStatusActive).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		addRole(models.RoleAnalyst)
	}

	count = 0
	if err := s.db.Model(&models.Scout{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		addRole(models.RoleScout)
	}

	count = 0
	if err := s.db.Model(&models.Club{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		addRole(models.RoleClub)
	}

	count = 0
	if err := s.db.Model(&models.ClubCoach{}).Where("user_id = ? AND status = ?", user.ID, models.ClubCoachStatusActive).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		addRole(models.RoleCoach)
	}

	count = 0
	if err := s.db.Model(&models.TeamCoach{}).Where("user_id = ? AND status = ?", user.ID, "active").Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		addRole(models.RoleCoach)
	}

	return roles, nil
}

func hasRole(roles []models.UserRole, role models.UserRole) bool {
	for _, activeRole := range roles {
		if activeRole == role {
			return true
		}
	}
	return false
}

func (s *AuthService) normalizeCurrentRole(user *models.User, requested string) (models.UserRole, error) {
	role := models.UserRole(strings.TrimSpace(requested))
	if role == "" {
		role = user.Role
	}

	roles, err := s.getActiveUserRoles(user)
	if err != nil {
		return "", err
	}
	if !hasRole(roles, role) {
		return "", fmt.Errorf("无权切换到该角色")
	}

	return role, nil
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required"` // player/analyst/club/coach/scout
	// 球员基本信息
	Name           string  `json:"name"`
	Nickname       string  `json:"nickname"`
	BirthDate      string  `json:"birth_date"`
	Age            int     `json:"age"`
	Gender         string  `json:"gender"`
	Height         float64 `json:"height"`
	Weight         float64 `json:"weight"`
	Foot           string  `json:"foot"`
	Position       string  `json:"position"`
	SecondPosition string  `json:"second_position"`
	StartYear      int     `json:"start_year"`
	Country        string  `json:"country"`
	Province       string  `json:"province"`
	City           string  `json:"city"`
	Club           string  `json:"club"`
	FARegistered   bool    `json:"fa_registered"`
	Association    string  `json:"association"`
	JerseyColor    string  `json:"jersey_color"`
	JerseyNumber   int     `json:"jersey_number"`
	// 家庭信息
	FatherHeight     float64 `json:"father_height"`
	FatherPhone      string  `json:"father_phone"`
	FatherOccupation string  `json:"father_occupation"`
	FatherEdu        string  `json:"father_edu"`
	FatherJob        string  `json:"father_job"`
	FatherAthlete    string  `json:"father_athlete"`
	MotherHeight     float64 `json:"mother_height"`
	MotherPhone      string  `json:"mother_phone"`
	MotherOccupation string  `json:"mother_occupation"`
	MotherEdu        string  `json:"mother_edu"`
	MotherJob        string  `json:"mother_job"`
	MotherAthlete    string  `json:"mother_athlete"`
	// 球员扩展字段
	CurrentTeam   string `json:"current_team"`
	PlayingStyle  string `json:"playing_style"` // JSON: ["tech","speed"]
	Wechat        string `json:"wechat"`
	School        string `json:"school"`
	TechnicalTags string `json:"technical_tags"` // JSON: ["盘带","射门"]
	MentalTags    string `json:"mental_tags"`    // JSON: ["领导力","抗压"]
	Experiences   string `json:"experiences"`    // JSON: [{period,team,position,achievement}]
	DominantFoot  string `json:"dominant_foot"`  // 惯用脚：left/right/both
	VideoUrl      string `json:"video_url"`      // 视频链接
	// 注册时填写的体测数据（存到 users 表字段）
	Sprint30m        float64 `json:"sprint_30m"`
	StandingLongJump float64 `json:"standing_long_jump"`
	Flexibility      float64 `json:"flexibility"`        // 坐位体前屈(cm)
	PullUps          int     `json:"pull_ups"`           // 引体向上(个)
	PushUp           int     `json:"push_up"`            // 俯卧撑(个)
	SitUps           int     `json:"sit_ups"`            // 仰卧起坐(个/分钟)
	FiveMeterShuttle float64 `json:"five_meter_shuttle"` // 5×25米折返跑(秒)
	Coordination     float64 `json:"coordination"`       // 协调性测试(秒)
	SitAndReach      float64 `json:"sit_and_reach"`      // 坐位体前屈(cm)
	// 俱乐部专属字段
	ClubName         string `json:"club_name"`
	ClubType         string `json:"club_type"`
	FoundedYear      int    `json:"founded_year"`
	ClubScale        string `json:"club_scale"`
	ClubAddress      string `json:"club_address"`
	ClubWebsite      string `json:"club_website"`
	ContactName      string `json:"contact_name"`
	ContactPosition  string `json:"contact_position"`
	ClubContactPhone string `json:"club_contact_phone"`
	// 教练专属字段
	CoachType       string `json:"coach_type"`
	LicenseLevel    string `json:"license_level"`
	LicenseNumber   string `json:"license_number"`
	CoachExperience string `json:"coach_experience"`
	CoachSpecialty  string `json:"coach_specialty"`
	// 分析师专属字段
	Profession   string `json:"profession"`
	Experience   string `json:"experience"`
	IsProPlayer  bool   `json:"is_pro_player"`
	HasCase      bool   `json:"has_case"`
	CaseDetail   string `json:"case_detail"`
	Certificates string `json:"certificates"`
	ContactPhone string `json:"contact_phone"`
	ContactEmail string `json:"contact_email"`
	// 球探专属字段
	ScoutingExperience  string `json:"scouting_experience"`
	ScoutingSpecialty   string `json:"scouting_specialty"`
	PreferredAgeGroups  string `json:"preferred_age_groups"`
	ScoutingRegions     string `json:"scouting_regions"`
	CurrentOrganization string `json:"current_organization"`
	// 俱乐部专属字段（注册时填写的规模信息）
	TeamCount    int    `json:"team_count"`
	PlayerCount  int    `json:"player_count"`
	CoachCount   int    `json:"coach_count"`
	Achievements string `json:"achievements"`
	// 头像（支持 base64 data URI）
	Avatar string `json:"avatar"`
	// 邀请码（通过邀请链接注册时自动入队）
	InviteCode string `json:"invite_code"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	Nickname       *string  `json:"nickname"`
	Avatar         *string  `json:"avatar"`
	Name           *string  `json:"name"`
	BirthDate      *string  `json:"birth_date"`
	Age            *int     `json:"age"`
	Gender         *string  `json:"gender"`
	Height         *float64 `json:"height"`
	Weight         *float64 `json:"weight"`
	Foot           *string  `json:"foot"`
	Position       *string  `json:"position"`
	SecondPosition *string  `json:"second_position"`
	StartYear      *int     `json:"start_year"`
	Country        *string  `json:"country"`
	Province       *string  `json:"province"`
	City           *string  `json:"city"`
	Club           *string  `json:"club"`
	FARegistered   *bool    `json:"fa_registered"`
	Association    *string  `json:"association"`
	JerseyColor    *string  `json:"jersey_color"`
	JerseyNumber   *int     `json:"jersey_number"`
	FatherHeight   *float64 `json:"father_height"`
	FatherPhone    *string  `json:"father_phone"`
	FatherEdu      *string  `json:"father_edu"`
	FatherJob      *string  `json:"father_job"`
	FatherAthlete  *bool    `json:"father_athlete"`
	MotherHeight   *float64 `json:"mother_height"`
	MotherPhone    *string  `json:"mother_phone"`
	MotherEdu      *string  `json:"mother_edu"`
	MotherJob      *string  `json:"mother_job"`
	MotherAthlete  *bool    `json:"mother_athlete"`
	CurrentRole    *string  `json:"current_role"`
	// 球员扩展字段
	CurrentTeam   *string `json:"current_team"`
	PlayingStyle  *string `json:"playing_style"`
	Wechat        *string `json:"wechat"`
	School        *string `json:"school"`
	TechnicalTags *string `json:"technical_tags"`
	MentalTags    *string `json:"mental_tags"`
	Experiences   *string `json:"experiences"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Message string       `json:"message"`
	Token   string       `json:"token"`
	User    *models.User `json:"user"`
}

// Register 用户注册
func (s *AuthService) Register(req *RegisterRequest) (*LoginResponse, error) {
	// 验证验证码
	smsCode, err := s.smsService.VerifyCode(req.Phone, req.Code, models.SmsCodeTypeRegister)
	if err != nil {
		return nil, err
	}
	if smsCode == nil {
		return nil, nil // 验证码无效
	}

	// 检查手机号是否已注册
	existingUser, err := s.userRepo.FindByPhone(req.Phone)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, fmt.Errorf("该手机号已注册，请直接登录") // 手机号已注册
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 根据角色确定用户状态（需要审核的角色设为 pending）
	var userStatus models.UserStatus = models.StatusActive
	var userRole models.UserRole = models.RoleUser

	switch req.Role {
	case "player":
		userRole = models.RoleUser // 球员直接激活
	case "analyst":
		userRole = models.RoleAnalyst
		if !config.IsAnalystRegistrationAutoApproved() {
			userStatus = models.StatusPending
		}
	case "club":
		userRole = models.RoleClub
		userStatus = models.StatusPending // 需要审核
	case "coach":
		userRole = models.RoleCoach
		userStatus = models.StatusPending // 需要审核
	case "scout":
		userRole = models.RoleScout
		userStatus = models.StatusPending // 需要审核
	}

	// 处理头像（base64 data URI -> 本地文件）
	avatarURL := req.Avatar
	if avatarURL != "" && strings.HasPrefix(avatarURL, "data:image") {
		re := regexp.MustCompile(`^data:image/([a-zA-Z]+);base64,`)
		matches := re.FindStringSubmatch(avatarURL)
		if len(matches) > 1 {
			ext := matches[1]
			if ext == "jpeg" {
				ext = "jpg"
			}
			base64Data := re.ReplaceAllString(avatarURL, "")
			decoded, err := base64.StdEncoding.DecodeString(base64Data)
			if err == nil {
				uploadDir := "./uploads/avatars"
				_ = os.MkdirAll(uploadDir, 0755)
				timestamp := time.Now().UnixNano()
				newFilename := fmt.Sprintf("reg_%d_%s.%s", timestamp, req.Phone, ext)
				savePath := filepath.Join(uploadDir, newFilename)
				if err := os.WriteFile(savePath, decoded, 0644); err == nil {
					baseURL := config.GetBaseUrl()
					avatarURL = fmt.Sprintf("%s/uploads/avatars/%s", baseURL, newFilename)
				}
			}
		}
	}

	age := req.Age
	if age <= 0 {
		age = calculateAgeFromBirthDate(req.BirthDate)
	}

	// 创建用户
	user := &models.User{
		Phone:          req.Phone,
		Password:       string(hashedPassword),
		Role:           userRole,
		Status:         userStatus,
		Name:           req.Name,
		Nickname:       req.Nickname,
		BirthDate:      req.BirthDate,
		Age:            age,
		Gender:         req.Gender,
		Height:         req.Height,
		Weight:         req.Weight,
		Foot:           req.Foot,
		Position:       req.Position,
		SecondPosition: req.SecondPosition,
		StartYear:      req.StartYear,
		Country:        req.Country,
		Province:       req.Province,
		City:           req.City,
		Club:           req.Club,
		FARegistered:   req.FARegistered,
		Association:    req.Association,
		JerseyColor:    req.JerseyColor,
		JerseyNumber:   req.JerseyNumber,
		FatherHeight:   req.FatherHeight,
		FatherPhone:    req.FatherPhone,
		FatherEdu:      req.FatherEdu,
		FatherJob:      req.FatherJob,
		FatherAthlete:  req.FatherAthlete == "true" || req.FatherAthlete == "yes",
		MotherHeight:   req.MotherHeight,
		MotherPhone:    req.MotherPhone,
		MotherEdu:      req.MotherEdu,
		MotherJob:      req.MotherJob,
		MotherAthlete:  req.MotherAthlete == "true" || req.MotherAthlete == "yes",
		Avatar:         avatarURL,
		// 球员扩展字段
		CurrentTeam:   req.CurrentTeam,
		PlayingStyle:  req.PlayingStyle,
		Wechat:        req.Wechat,
		School:        req.School,
		TechnicalTags: req.TechnicalTags,
		MentalTags:    req.MentalTags,
		Experiences:   req.Experiences,
		// 注册时填写的体测数据
		Sprint30m:        req.Sprint30m,
		StandingLongJump: req.StandingLongJump,
		PushUp:           req.PushUp,
		SitAndReach:      req.SitAndReach,
		// 俱乐部扩展字段
		TeamCount:    req.TeamCount,
		PlayerCount:  req.PlayerCount,
		CoachCount:   req.CoachCount,
		Achievements: req.Achievements,
		// 新增球员扩展字段
		DominantFoot: req.DominantFoot,
		VideoUrl:     req.VideoUrl,
		// 新增体测字段
		Flexibility:      req.Flexibility,
		PullUps:          req.PullUps,
		SitUps:           req.SitUps,
		FiveMeterShuttle: req.FiveMeterShuttle,
		Coordination:     req.Coordination,
		// 新增家庭信息字段
		FatherOccupation: req.FatherOccupation,
		MotherOccupation: req.MotherOccupation,
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		if req.Role == "analyst" {
			if err := s.createAnalystProfileForRegisteredUser(tx, user, req); err != nil {
				return err
			}
		}

		if req.Role == "club" && req.ClubName != "" {
			club := &models.Club{
				UserID:          user.ID,
				Name:            req.ClubName,
				Logo:            avatarURL,
				Address:         req.ClubAddress,
				ContactName:     req.ContactName,
				ContactPhone:    req.ClubContactPhone,
				EstablishedYear: req.FoundedYear,
				ClubSize:        req.ClubScale,
			}
			if err := tx.Create(club).Error; err != nil {
				return err
			}
		}

		roleRecordStatus := "active"
		if user.Status != models.StatusActive {
			roleRecordStatus = "pending"
		}
		if err := models.UpsertUserRoleRecord(tx, models.UserRoleRecord{
			UserID:        user.ID,
			Role:          user.Role,
			Status:        roleRecordStatus,
			Source:        "register",
			PublicVisible: true,
			ApprovedAt:    timePtrIfActive(user.Status),
		}); err != nil && !models.IsMissingUserRolesTableError(err) {
			return err
		}

		if req.InviteCode != "" {
			_ = s.processInviteOnRegister(tx, user, req.InviteCode)
		}

		if req.Role == "analyst" && config.IsAnalystDefaultDemoOrderEnabled() {
			if err := s.ensureAnalystDefaultDemoOrder(tx, user.ID); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// 标记验证码为已使用
	err = s.smsService.MarkAsUsed(smsCode.ID)
	if err != nil {
		return nil, err
	}

	// 生成JWT
	token, err := middleware.GenerateToken(user.ID, user.Phone)
	if err != nil {
		return nil, err
	}

	// 重新获取完整用户信息（包含 Roles 和 CurrentRole 填充）
	user, err = s.GetUserByID(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Message: "注册成功",
		Token:   token,
		User:    user,
	}, nil
}

func (s *AuthService) createAnalystProfileForRegisteredUser(db *gorm.DB, user *models.User, req *RegisterRequest) error {
	var existingCount int64
	if err := db.Model(&models.Analyst{}).Where("user_id = ?", user.ID).Count(&existingCount).Error; err != nil {
		return err
	}
	if existingCount > 0 {
		return nil
	}

	status := models.AnalystStatusActive
	if !config.IsAnalystRegistrationAutoApproved() {
		status = models.AnalystStatusInactive
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = strings.TrimSpace(req.Nickname)
	}
	if name == "" {
		name = "新分析师"
	}

	specialty := ""
	if req.Profession != "" {
		if encoded, err := json.Marshal([]string{req.Profession}); err == nil {
			specialty = string(encoded)
		}
	}

	analyst := &models.Analyst{
		UserID:       user.ID,
		Name:         name,
		Bio:          req.Experience,
		Specialty:    specialty,
		Experience:   firstInt(req.Experience),
		Profession:   req.Profession,
		IsProPlayer:  req.IsProPlayer,
		HasCase:      req.HasCase,
		CaseDetail:   req.CaseDetail,
		ContactPhone: firstNonEmpty(req.ContactPhone, req.ClubContactPhone),
		ContactEmail: req.ContactEmail,
		Status:       status,
		Rating:       0,
		ReviewCount:  0,
	}

	if err := db.Create(analyst).Error; err != nil {
		return fmt.Errorf("创建分析师资料失败: %w", err)
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstInt(value string) int {
	matches := regexp.MustCompile(`\d+`).FindString(value)
	if matches == "" {
		return 0
	}
	parsed, err := strconv.Atoi(matches)
	if err != nil {
		return 0
	}
	return parsed
}

func timePtrIfActive(status models.UserStatus) *time.Time {
	if status != models.StatusActive {
		return nil
	}
	now := time.Now()
	return &now
}

func (s *AuthService) ensureAnalystDefaultDemoOrder(tx *gorm.DB, userID uint) error {
	if s.analystRepo == nil || s.orderRepo == nil {
		return nil
	}

	templateOrderID, err := config.GetAnalystDefaultDemoOrderTemplateOrderID()
	if err != nil || templateOrderID == 0 {
		return fmt.Errorf("默认样例订单模板配置无效")
	}

	var analyst models.Analyst
	if err := tx.Where("user_id = ?", userID).First(&analyst).Error; err != nil {
		return fmt.Errorf("查询分析师资料失败: %w", err)
	}
	if analyst.Status != models.AnalystStatusActive {
		return nil
	}

	var templateOrder models.Order
	if err := tx.First(&templateOrder, templateOrderID).Error; err != nil {
		return fmt.Errorf("查询默认样例订单模板失败: %w", err)
	}
	if templateOrder.Status != models.OrderStatusUploaded && templateOrder.Status != models.OrderStatusAssigned &&
		templateOrder.Status != models.OrderStatusProcessing && templateOrder.Status != models.OrderStatusCompleted {
		return fmt.Errorf("默认样例订单模板状态不合法: %s", templateOrder.Status)
	}
	if strings.TrimSpace(templateOrder.VideoURL) == "" {
		return fmt.Errorf("默认样例订单模板缺少视频地址")
	}

	remarkPrefix := fmt.Sprintf("[系统样例订单 template=%d analyst=%d]", templateOrder.ID, analyst.ID)
	var existingCount int64
	if err := tx.Model(&models.Order{}).
		Where("analyst_id = ? AND remark LIKE ?", analyst.ID, remarkPrefix+"%").
		Count(&existingCount).Error; err != nil {
		return fmt.Errorf("检查默认样例订单失败: %w", err)
	}
	if existingCount > 0 {
		return nil
	}

	assignedAt := time.Now()
	deadlineHours := 48
	if templateOrder.OrderType == "pro" {
		deadlineHours = 72
	}
	deadline := assignedAt.Add(time.Duration(deadlineHours) * time.Hour)
	orderNo := fmt.Sprintf("DEMO%dA%d%04d", assignedAt.UnixNano(), analyst.ID, templateOrder.ID%10000)
	remark := remarkPrefix
	if trimmed := strings.TrimSpace(templateOrder.Remark); trimmed != "" {
		remark = remark + " " + trimmed
	}

	playerAge := templateOrder.PlayerAge
	if playerAge <= 0 && templateOrder.UserID > 0 {
		var templatePlayer models.User
		if err := tx.Select("id", "birth_date").First(&templatePlayer, templateOrder.UserID).Error; err == nil {
			playerAge = calculateAgeFromBirthDate(templatePlayer.BirthDate)
		}
	}

	clonedOrder := &models.Order{
		UserID:         templateOrder.UserID,
		AnalystID:      &analyst.ID,
		OrderNo:        orderNo,
		Amount:         templateOrder.Amount,
		Status:         models.OrderStatusAssigned,
		PaymentMethod:  templateOrder.PaymentMethod,
		VideoURL:       templateOrder.VideoURL,
		VideoFilename:  templateOrder.VideoFilename,
		PaidAt:         &assignedAt,
		Remark:         remark,
		OrderType:      templateOrder.OrderType,
		PlayerName:     templateOrder.PlayerName,
		PlayerAge:      playerAge,
		PlayerPosition: templateOrder.PlayerPosition,
		JerseyColor:    templateOrder.JerseyColor,
		JerseyNumber:   templateOrder.JerseyNumber,
		MatchName:      templateOrder.MatchName,
		MatchDate:      templateOrder.MatchDate,
		Opponent:       templateOrder.Opponent,
		MatchResult:    templateOrder.MatchResult,
		VideoDuration:  templateOrder.VideoDuration,
		Deadline:       &deadline,
		AssignedAt:     &assignedAt,
	}
	if err := tx.Create(clonedOrder).Error; err != nil {
		return fmt.Errorf("创建默认样例订单失败: %w", err)
	}

	if s.assignmentRepo != nil {
		assignment := &models.OrderAssignment{
			OrderID:    clonedOrder.ID,
			AnalystID:  analyst.ID,
			AssignedAt: assignedAt,
			Status:     models.OrderAssignmentStatusPending,
		}
		if err := s.assignmentRepo.CreateWithTx(tx, assignment); err != nil {
			return fmt.Errorf("创建默认样例派单记录失败: %w", err)
		}
	}

	if s.statusHistoryRepo != nil {
		history := &models.OrderStatusHistory{
			OrderID:    clonedOrder.ID,
			FromStatus: models.OrderStatusUploaded,
			ToStatus:   models.OrderStatusAssigned,
			ActorRole:  "system",
			Reason:     "分析师注册默认样例订单",
		}
		if err := s.statusHistoryRepo.CreateWithTx(tx, history); err != nil {
			return fmt.Errorf("写入默认样例订单状态记录失败: %w", err)
		}
	}

	return nil
}

// processInviteOnRegister 注册时自动处理邀请码入队/入俱乐部
func (s *AuthService) processInviteOnRegister(tx *gorm.DB, user *models.User, inviteCode string) error {
	// 先尝试查找球队邀请
	var teamInv models.TeamInvitation
	teamErr := tx.Where("invite_code = ?", inviteCode).First(&teamInv).Error
	if teamErr == nil {
		return s.processTeamInviteOnRegister(tx, user, &teamInv)
	}

	// 再尝试查找俱乐部邀请
	var clubInv models.ClubInvitation
	clubErr := tx.Where("invite_code = ?", inviteCode).First(&clubInv).Error
	if clubErr == nil {
		return s.processClubInviteOnRegister(tx, user, &clubInv)
	}

	return fmt.Errorf("邀请码无效")
}

// processTeamInviteOnRegister 处理球队邀请
func (s *AuthService) processTeamInviteOnRegister(tx *gorm.DB, user *models.User, inv *models.TeamInvitation) error {
	if inv.Status != models.InvitationStatusPending {
		return fmt.Errorf("邀请已处理")
	}
	if time.Now().After(inv.ExpiresAt) {
		inv.Status = models.InvitationStatusExpired
		_ = tx.Save(inv).Error
		return fmt.Errorf("邀请已过期")
	}

	now := time.Now()

	if inv.Type == models.InvitationTypePlayer {
		var existingCount int64
		tx.Model(&models.TeamPlayer{}).Where("team_id = ? AND user_id = ?", inv.TeamID, user.ID).Count(&existingCount)
		if existingCount == 0 {
			// 获取球队信息以填充 ClubPlayer 字段
			var team models.Team
			teamAgeGroup := ""
			if err := tx.First(&team, inv.TeamID).Error; err == nil {
				teamAgeGroup = team.AgeGroup
			}

			tp := &models.TeamPlayer{
				TeamID:   inv.TeamID,
				UserID:   user.ID,
				Status:   "active",
				JoinedAt: now,
			}
			_ = tx.Create(tp).Error

			// 同步创建 ClubPlayer（如果不存在）
			var clubPlayerCount int64
			tx.Model(&models.ClubPlayer{}).Where("club_id = ? AND user_id = ?", inv.ClubID, user.ID).Count(&clubPlayerCount)
			if clubPlayerCount == 0 {
				cp := &models.ClubPlayer{
					ClubID:   inv.ClubID,
					UserID:   user.ID,
					JoinDate: now,
					AgeGroup: teamAgeGroup,
					Status:   "active",
				}
				_ = tx.Create(cp).Error
			}
		}
	} else if inv.Type == models.InvitationTypeCoach {
		var existingCount int64
		tx.Model(&models.TeamCoach{}).Where("team_id = ? AND user_id = ?", inv.TeamID, user.ID).Count(&existingCount)
		if existingCount == 0 {
			tc := &models.TeamCoach{
				TeamID:   inv.TeamID,
				UserID:   user.ID,
				Role:     "assistant",
				Status:   "active",
				JoinedAt: now,
			}
			_ = tx.Create(tc).Error
		}
	}

	inv.Status = models.InvitationStatusAccepted
	inv.TargetUserID = &user.ID
	inv.AcceptedAt = &now
	return tx.Save(inv).Error
}

// processClubInviteOnRegister 处理俱乐部邀请
func (s *AuthService) processClubInviteOnRegister(tx *gorm.DB, user *models.User, inv *models.ClubInvitation) error {
	if inv.Status != models.InvitationStatusPending {
		return fmt.Errorf("邀请已处理")
	}
	if time.Now().After(inv.ExpiresAt) {
		inv.Status = models.InvitationStatusExpired
		_ = tx.Save(inv).Error
		return fmt.Errorf("邀请已过期")
	}

	now := time.Now()

	// 创建 ClubCoach 关联记录
	var existingCount int64
	tx.Model(&models.ClubCoach{}).Where("club_id = ? AND user_id = ?", inv.ClubID, user.ID).Count(&existingCount)
	if existingCount == 0 {
		cc := &models.ClubCoach{
			ClubID:      inv.ClubID,
			UserID:      user.ID,
			PrimaryRole: inv.TargetRole,
			Status:      models.ClubCoachStatusActive,
			JoinedAt:    now,
		}
		_ = tx.Create(cc).Error
	}

	inv.Status = models.InvitationStatusAccepted
	inv.TargetUserID = &user.ID
	inv.AcceptedAt = &now
	return tx.Save(inv).Error
}

// ResetPassword 重置密码
func (s *AuthService) ResetPassword(req *ResetPasswordRequest) (bool, error) {
	// 验证验证码
	smsCode, err := s.smsService.VerifyCode(req.Phone, req.Code, models.SmsCodeTypeReset)
	if err != nil {
		return false, err
	}
	if smsCode == nil {
		return false, nil // 验证码无效
	}

	// 查找用户
	user, err := s.userRepo.FindByPhone(req.Phone)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil // 用户不存在
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}

	// 更新密码
	updates := map[string]interface{}{
		"password": string(hashedPassword),
	}
	err = s.userRepo.Update(user.ID, updates)
	if err != nil {
		return false, err
	}

	// 标记验证码为已使用
	err = s.smsService.MarkAsUsed(smsCode.ID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(id uint) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil || user == nil {
		return user, err
	}

	activeRoles, err := s.getActiveUserRoles(user)
	if err != nil {
		return nil, err
	}

	if len(activeRoles) == 0 {
		user.CurrentRole = user.Role
		user.Roles = []models.UserRoleInfo{
			{Type: user.Role, Status: string(user.Status)},
		}
		return user, nil
	}

	if user.CurrentRole == "" || !hasRole(activeRoles, user.CurrentRole) {
		user.CurrentRole = activeRoles[0]
	}

	user.Roles = make([]models.UserRoleInfo, 0, len(activeRoles))
	for _, role := range activeRoles {
		user.Roles = append(user.Roles, models.UserRoleInfo{Type: role, Status: "active"})
	}

	return user, nil
}

// UpdateUser 更新用户信息
func (s *AuthService) UpdateUser(id uint, req *UpdateUserRequest) (*models.User, error) {
	existingUser, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if existingUser == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	updates := make(map[string]interface{})

	// 逐个添加非nil字段
	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.BirthDate != nil {
		updates["birth_date"] = *req.BirthDate
		updates["age"] = calculateAgeFromBirthDate(*req.BirthDate)
	} else if req.Age != nil {
		updates["age"] = *req.Age
	}
	if req.Gender != nil {
		updates["gender"] = models.Gender(*req.Gender)
	}
	if req.Height != nil {
		updates["height"] = *req.Height
	}
	if req.Weight != nil {
		updates["weight"] = *req.Weight
	}
	if req.Foot != nil {
		updates["foot"] = models.Foot(*req.Foot)
	}
	if req.Position != nil {
		updates["position"] = *req.Position
	}
	if req.SecondPosition != nil {
		updates["second_position"] = *req.SecondPosition
	}
	if req.StartYear != nil {
		updates["start_year"] = *req.StartYear
	}
	if req.Country != nil {
		updates["country"] = *req.Country
	}
	if req.Province != nil {
		updates["province"] = *req.Province
	}
	if req.City != nil {
		updates["city"] = *req.City
	}
	if req.Club != nil {
		updates["club"] = *req.Club
	}
	if req.FARegistered != nil {
		updates["fa_registered"] = *req.FARegistered
	}
	if req.Association != nil {
		updates["association"] = *req.Association
	}
	if req.JerseyColor != nil {
		updates["jersey_color"] = *req.JerseyColor
	}
	if req.JerseyNumber != nil {
		updates["jersey_number"] = *req.JerseyNumber
	}
	if req.FatherHeight != nil {
		updates["father_height"] = *req.FatherHeight
	}
	if req.FatherPhone != nil {
		updates["father_phone"] = *req.FatherPhone
	}
	if req.FatherEdu != nil {
		updates["father_edu"] = *req.FatherEdu
	}
	if req.FatherJob != nil {
		updates["father_job"] = *req.FatherJob
	}
	if req.FatherAthlete != nil {
		updates["father_athlete"] = *req.FatherAthlete
	}
	if req.MotherHeight != nil {
		updates["mother_height"] = *req.MotherHeight
	}
	if req.MotherPhone != nil {
		updates["mother_phone"] = *req.MotherPhone
	}
	if req.MotherEdu != nil {
		updates["mother_edu"] = *req.MotherEdu
	}
	if req.MotherJob != nil {
		updates["mother_job"] = *req.MotherJob
	}
	if req.MotherAthlete != nil {
		updates["mother_athlete"] = *req.MotherAthlete
	}
	if req.CurrentRole != nil {
		currentRole, err := s.normalizeCurrentRole(existingUser, *req.CurrentRole)
		if err != nil {
			return nil, err
		}
		updates["current_role"] = currentRole
	}

	if len(updates) == 0 {
		return s.userRepo.FindByID(id)
	}

	err = s.userRepo.Update(id, updates)
	if err != nil {
		return nil, err
	}

	return s.userRepo.FindByID(id)
}

// CheckPhoneExists 检查手机号是否已注册
func (s *AuthService) CheckPhoneExists(phone string) (bool, error) {
	user, err := s.userRepo.FindByPhone(phone)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}

// VerifyCodeRequest 验证码校验请求
type VerifyCodeRequest struct {
	Phone string             `json:"phone" binding:"required"`
	Code  string             `json:"code" binding:"required,len=6"`
	Type  models.SmsCodeType `json:"type" binding:"required,oneof=register reset"`
}

// VerifyCodeResponse 验证码校验响应
type VerifyCodeResponse struct {
	Valid bool   `json:"valid"`
	Phone string `json:"phone"`
}

// VerifyCode 验证码校验
func (s *AuthService) VerifyCode(req *VerifyCodeRequest) (*VerifyCodeResponse, error) {
	smsCode, err := s.smsService.VerifyCode(req.Phone, req.Code, req.Type)
	if err != nil {
		return nil, err
	}

	return &VerifyCodeResponse{
		Valid: smsCode != nil,
		Phone: req.Phone,
	}, nil
}

// RefreshTokenResponse 刷新Token响应
type RefreshTokenResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// RefreshToken 刷新Token
func (s *AuthService) RefreshToken(userID uint) (*RefreshTokenResponse, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	// 生成新Token
	token, err := middleware.GenerateToken(user.ID, user.Phone)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenResponse{
		Token: token,
		User:  user,
	}, nil
}

// Login 用户登录
func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.FindByPhone(req.Phone)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, nil
	}

	if user.Status != models.StatusActive {
		return nil, ErrAccountNotActive
	}

	token, err := middleware.GenerateToken(user.ID, user.Phone)
	if err != nil {
		return nil, err
	}

	// 重新获取完整用户信息（填充 current_role 默认值和 Roles 数组）
	user, err = s.GetUserByID(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Message: "登录成功",
		Token:   token,
		User:    user,
	}, nil
}
