package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"github.com/shaonianqiutan/backend/routes"
	"github.com/shaonianqiutan/backend/services"
	wshub "github.com/shaonianqiutan/backend/wshub"
)

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}

	return false
}

func main() {
	// 加载环境变量
	config.LoadEnv()
	config.ValidateRuntimeConfig()

	// 初始化数据库
	config.InitDB()

	// 自动迁移表
	db := config.GetDB()
	err := db.AutoMigrate(
		&models.User{},
		&models.Player{},
		&models.SmsCode{},
		&models.Report{},
		&models.ReportVersion{},
		&models.Order{},
		&models.OrderAssignment{},
		&models.OrderStatusHistory{},
		&models.StorageObject{},
		&models.Analyst{},
		&models.AnalystApplication{},
		&models.UserRoleRecord{},
		&models.RoleApplication{},
		&models.Scout{},
		&models.Club{},
		&models.ClubPlayer{},
		&models.Team{},
		&models.TeamInvitation{},
		&models.TeamPlayer{},
		&models.ClubCoach{},
		&models.TeamCoach{},
		&models.ClubInvitation{},
		&models.TeamApplication{},
		&models.PhysicalTestActivity{},
		&models.PhysicalTestRecord{},
		&models.PhysicalTestReport{},
		&models.PhysicalTestTemplateCustom{},
		&models.ClubHome{},
		&models.ClubHomeTeam{},
		&models.ClubHomeCoach{},
		&models.ClubHomePlayer{},
		&models.ClubActivity{},
		&models.ClubActivityRegistration{},
		&models.TeamHome{},
		&models.WeeklyReport{},
		&models.WeeklyReportPeriod{},
		&models.MatchSummary{},
		&models.PlayerReview{},
		&models.MatchVideo{},
		&models.Notification{},
		&models.Comment{},
		&models.Like{},
		&models.Post{},
		&models.SocialAchievement{},
		&models.Follow{},
		&models.GrowthRecord{},
		&models.TrainingPlan{},
		&models.MatchSchedule{},
		&models.VideoAnalysis{},
		&models.AnalysisHighlight{},
		&models.VideoClipExportJob{},
		&models.PlayerFilterPreset{},
		&models.AdminOperationLog{},
		&models.PlayerShortlist{},
		&models.TeamSeasonArchive{},
		&models.TrialInvite{},
		&models.Message{},
		&models.ContentReport{},
		&models.SensitiveWord{},
		&models.PlatformAnnouncement{},
		&models.Banner{},
		&models.FAQ{},
		&models.LoginLog{},
		&models.SystemSetting{},
		&models.AdminPermission{},
		&models.AdminRole{},
		&models.AdminRolePermission{},
		&models.AdminUserRole{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库迁移完成")
	if err := models.BackfillUserRoleRecords(db); err != nil {
		log.Printf("用户多身份记录回填失败: %v", err)
	}
	if err := models.BackfillOrderAssignmentsFromOrders(db); err != nil {
		log.Printf("订单派发记录回填失败: %v", err)
	}
	if err := models.SeedDefaultAdminRBAC(db); err != nil {
		log.Printf("管理员权限默认数据初始化失败: %v", err)
	}

	// 初始化 WebSocket Hub
	hub := wshub.NewHub()
	go hub.Run()
	log.Println("WebSocket Hub 已启动")

	// 初始化通知服务（WebSocket推送）
	notifyService := wshub.GetNotifyService()
	notifyService.SetHub(hub)

	// 创建 Gin 引擎
	if os.Getenv("GIN_MODE") == "" {
		if config.IsDevMode() {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}
	}
	r := gin.Default()

	// CORS 中间件 — 只允许特定前端域名，生产环境绝不使用 "*"
	allowedOrigins := config.GetCORSOrigins()
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if isAllowedOrigin(origin, allowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
	r.Use(middleware.MaintenanceModeMiddleware())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"ws_users": hub.GetOnlineUsers(),
		})
	})

	// WebSocket 路由（需要认证）
	r.GET("/ws", middleware.QueryTokenAuthMiddleware(), func(c *gin.Context) {
		userID, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			return
		}

		var upgrader = ws.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				return isAllowedOrigin(origin, allowedOrigins)
			},
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket 升级失败: %v", err)
			return
		}

		client := &wshub.Client{
			Hub:    hub,
			Conn:   conn,
			Send:   make(chan []byte, 256),
			UserID: userID.(uint),
		}
		hub.Register <- client

		go client.WritePump()
		go client.ReadPump()
	})

	// ========== API 路由组 ==========
	api := r.Group("/api")
	{
		// ========== Repository 初始化 ==========
		userRepo := models.NewUserRepository(db)
		orderRepo := models.NewOrderRepository(db)
		reportRepo := models.NewReportRepository(db)
		storageObjectRepo := models.NewStorageObjectRepository(db)
		analystRepo := models.NewAnalystRepository(db)
		assignmentRepo := models.NewOrderAssignmentRepository(db)
		statusHistoryRepo := models.NewOrderStatusHistoryRepository(db)
		analystApplicationRepo := models.NewAnalystApplicationRepository(db)
		smsCodeRepo := models.NewSmsCodeRepository(db)

		teamRepo := repositories.NewTeamRepository(db)
		teamHomeRepo := repositories.NewTeamHomeRepository(db)
		weeklyReportRepo := repositories.NewWeeklyReportRepository(db)
		matchSummaryRepo := repositories.NewMatchSummaryRepository(db)
		playerReviewRepo := repositories.NewPlayerReviewRepository(db)
		matchVideoRepo := repositories.NewMatchVideoRepository(db)
		notificationRepo := repositories.NewNotificationRepository(db)
		socialRepo := repositories.NewSocialRepository(db)
		messageRepo := repositories.NewMessageRepository(db)
		adminLogRepo := repositories.NewAdminOperationLogRepository(db)
		contentReportRepo := models.NewContentReportRepository(db)
		sensitiveWordRepo := models.NewSensitiveWordRepository(db)
		platformAnnRepo := models.NewPlatformAnnouncementRepository(db)
		bannerRepo := models.NewBannerRepository(db)
		faqRepo := models.NewFAQRepository(db)
		loginLogRepo := models.NewLoginLogRepository(db)

		// ========== Service 初始化 ==========
		smsService := services.NewSmsService(smsCodeRepo)
		authService := services.NewAuthService(userRepo, analystRepo, orderRepo, assignmentRepo, statusHistoryRepo, smsService, db)
		storageService := services.NewStorageService(storageObjectRepo)
		orderService := services.NewOrderService(orderRepo, analystRepo, reportRepo, userRepo, storageService)
		reportService := services.NewReportService(reportRepo, userRepo)
		analystService := services.NewAnalystService(analystRepo, orderRepo, userRepo, assignmentRepo, statusHistoryRepo)
		clubService := services.NewClubService(db)
		scoutService := services.NewScoutService(db)
		coachService := services.NewCoachService(db, teamRepo, weeklyReportRepo, matchSummaryRepo)
		physicalTestService := services.NewPhysicalTestService(db)
		weeklyReportService := services.NewWeeklyReportService(db, weeklyReportRepo, teamRepo, userRepo)
		matchSummaryService := services.NewMatchSummaryService(db, matchSummaryRepo, playerReviewRepo, matchVideoRepo, teamRepo, userRepo)
		notificationService := services.NewNotificationService(db, notificationRepo, userRepo)
		analystService.SetNotificationService(notificationService)
		weeklyReportService.SetNotificationService(notificationService)
		clubOrderService := services.NewClubOrderService(db)
		socialService := services.NewSocialService(socialRepo, notificationService)
		messageService := services.NewMessageService(messageRepo, userRepo, socialRepo, notificationService)
		videoAnalysisRepo := models.NewVideoAnalysisRepository(db)
		adminService := services.NewAdminService(userRepo, reportRepo, orderRepo, analystRepo, analystApplicationRepo, contentReportRepo, sensitiveWordRepo, platformAnnRepo, bannerRepo, faqRepo, loginLogRepo, videoAnalysisRepo, assignmentRepo, statusHistoryRepo)
		adminService.SetNotificationService(notificationService)

		// ========== Controller 初始化 ==========
		authController := controllers.NewAuthController(authService, smsService)
		userController := controllers.NewUserController(authService, physicalTestService, db)
		accountRoleController := controllers.NewAccountRoleController(db)
		orderController := controllers.NewOrderController(orderService)
		reportController := controllers.NewReportController(reportService, authService, db)
		analystController := controllers.NewAnalystController(analystService, db)
		clubController := controllers.NewClubController(clubService, db, weeklyReportRepo, matchSummaryRepo, orderRepo, physicalTestService, adminLogRepo)
		physicalTestController := controllers.NewPhysicalTestController(physicalTestService)
		weeklyReportController := controllers.NewWeeklyReportController(weeklyReportService, db)
		matchSummaryController := controllers.NewMatchSummaryController(matchSummaryService, db)
		playerReviewController := controllers.NewPlayerReviewController(matchSummaryService, db)
		matchVideoController := controllers.NewMatchVideoController(matchSummaryService, db)
		trainingPlanController := controllers.NewTrainingPlanController(clubService, db)
		matchScheduleController := controllers.NewMatchScheduleController(clubService, db)
		notificationController := controllers.NewNotificationController(notificationService)
		socialController := controllers.NewSocialController(socialService)
		messageController := controllers.NewMessageController(messageService)
		adminController := controllers.NewAdminController(adminService, videoAnalysisRepo)
		scoutController := controllers.NewScoutController(scoutService)
		coachController := controllers.NewCoachController(coachService)
		coachTeamHomeController := controllers.NewCoachTeamHomeController(teamHomeRepo, teamRepo)
		clubHomeController := controllers.NewClubHomeController(repositories.NewClubHomeRepository(db), db)
		clubActivityController := controllers.NewClubActivityController(db, notificationService)
		clubOrderController := controllers.NewClubOrderController(clubOrderService)
		footballExpController := controllers.NewFootballExperienceController(coachService)
		paymentController := controllers.NewPaymentController(orderRepo)
		uploadController := controllers.NewUploadController()
		mapController := controllers.NewMapController()
		trialInviteController := controllers.NewTrialInviteController()
		playerController := controllers.NewPlayerController(db)

		// AI服务 + 视频分析控制器
		aiService := services.NewAIService(services.DefaultAIConfig)
		videoAnalysisController := controllers.NewVideoAnalysisController(db, aiService)
		videoAnalysisController.SetStorageService(storageService)
		videoAnalysisController.SetNotificationService(notificationService)

		// ========== 设置路由 ==========
		routes.SetupAuthRoutes(api, authController)
		routes.SetupUserRoutes(api, userController)
		routes.SetupAccountRoleRoutes(api, accountRoleController)
		routes.SetupOrderRoutes(api, orderController)
		routes.SetupReportRoutes(api, reportController)
		routes.SetupAnalystRoutes(api, analystController)
		routes.SetupClubRoutes(api, clubController, trainingPlanController, matchScheduleController)
		routes.SetupTeamRoutes(api, teamRepo, weeklyReportRepo, matchSummaryRepo, coachTeamHomeController, physicalTestController, db)
		routes.SetupPublicTeamRoutes(api, teamRepo, db)
		routes.SetupUserSearchRoutes(api, teamRepo)
		routes.SetupPhysicalTestRoutes(api, physicalTestController)
		routes.SetupClubHomeRoutes(api, clubHomeController)
		routes.SetupClubActivityRoutes(api, clubActivityController)
		routes.SetupClubOrderRoutes(api, clubOrderController)
		routes.SetupPaymentRoutes(api, paymentController)
		routes.SetupUploadRoutes(api, uploadController)
		routes.SetupVideoAnalysisRoutes(api, videoAnalysisController)
		routes.SetupWeeklyReportRoutes(api, weeklyReportController)
		routes.SetupWeeklyPeriodRoutes(api, weeklyReportController)
		routes.SetupMatchSummaryRoutes(api, matchSummaryController, playerReviewController, matchVideoController)
		routes.SetupNotificationRoutes(api, notificationController)
		socialApiGroup := api.Group("/social")
		routes.SetupSocialRoutes(socialApiGroup, socialController)
		routes.SetupMessageRoutes(api, messageController)
		routes.SetupAdminRoutes(api, adminController)
		routes.SetupSystemRoutes(api, adminController)
		routes.SetupScoutRoutes(api, scoutController)
		routes.SetupScoutMapRoutes(api, mapController)
		routes.SetupTrialInviteRoutes(api, trialInviteController)
		routes.SetupPlayerRoutes(api, playerController)
		routes.SetupPlayerPublicRoutes(api, playerController)
		routes.SetupCoachRoutes(api, coachController, teamRepo, footballExpController, weeklyReportController, db)
		routes.SetupTeamHomeRoutes(api, coachTeamHomeController)

		// 启动周报定时任务
		weeklyReportCron := services.NewWeeklyReportCron(db, weeklyReportService, teamRepo, notificationService)
		weeklyReportCron.Start()
		defer weeklyReportCron.Stop()
	}

	// 本地开发环境：提供上传文件的静态访问（仅限特定类型）
	if config.IsDevMode() {
		// 使用 StaticFS 限制可访问的文件类型，只允许图片和视频
		uploadsFS := http.Dir("./uploads")
		r.GET("/uploads/*filepath", func(c *gin.Context) {
			filepath := strings.TrimPrefix(c.Param("filepath"), "/")
			// 禁止访问以 . 开头的隐藏文件和目录遍历
			if strings.Contains(filepath, "..") || strings.Contains(filepath, "//") {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			// 只允许特定扩展名
			allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".mp4", ".mov", ".webm", ".pdf", ".doc", ".docx"}
			lowerPath := strings.ToLower(filepath)
			valid := false
			for _, ext := range allowedExts {
				if strings.HasSuffix(lowerPath, ext) {
					valid = true
					break
				}
			}
			if !valid {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.FileFromFS(filepath, uploadsFS)
		})
		log.Println("本地文件服务已启用（受限）: /uploads -> ./uploads")
	}

	// 启动服务器
	port := config.GetPort()
	log.Printf("服务器启动在 http://localhost%s", port)
	log.Printf("WebSocket 端点: ws://localhost%s/ws", port)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := r.Run(port); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	<-quit
	log.Println("服务器关闭中...")
}
