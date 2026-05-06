package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

const (
	demoPasswordHash = "$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK"
	expectedDBName   = "shaonianqiutan.db"
)

type seedContext struct {
	admin       models.User
	clubOwner   models.User
	backupClub  models.User
	club        models.Club
	backup      models.Club
	coaches     []models.User
	analysts    []models.User
	scouts      []models.User
	players     []models.User
	playerRows  []models.Player
	teams       []models.Team
	analystRows []models.Analyst
	scoutRows   []models.Scout
}

type playerSeed struct {
	Phone          string
	Name           string
	Nickname       string
	BirthDate      string
	Age            int
	Gender         string
	Height         float64
	Weight         float64
	Foot           string
	Position       string
	SecondPosition string
	Province       string
	City           string
	School         string
	JerseyNumber   int
	TeamIndex      int
	Tags           []string
}

type coachSeed struct {
	Phone        string
	Name         string
	Nickname     string
	License      string
	Role         models.CoachRole
	TeamIndex    int
	Specialties  []string
	AgeGroups    []string
	CoachingYear int
}

type analystSeed struct {
	Phone       string
	Name        string
	Nickname    string
	Specialty   string
	Profession  string
	Experience  int
	Rating      float64
	ReviewCount int
}

type scoutSeed struct {
	Phone        string
	Name         string
	Nickname     string
	Regions      []string
	Specialties  []string
	Organization string
}

func main() {
	config.LoadEnv()
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		log.Fatalf("DB_PATH is required, expected absolute path ending with %s", expectedDBName)
	}
	absDBPath, err := filepath.Abs(dbPath)
	if err != nil {
		log.Fatalf("resolve DB_PATH: %v", err)
	}
	if filepath.Base(absDBPath) != expectedDBName {
		log.Fatalf("refuse to seed unexpected database: %s", absDBPath)
	}
	if _, err := os.Stat(absDBPath); err != nil {
		log.Fatalf("database does not exist: %s: %v", absDBPath, err)
	}

	backupPath, err := backupDatabase(absDBPath)
	if err != nil {
		log.Fatalf("backup database: %v", err)
	}
	log.Printf("database backup created: %s", backupPath)

	if err := os.Setenv("DB_PATH", absDBPath); err != nil {
		log.Fatalf("set DB_PATH: %v", err)
	}
	config.InitDB()
	db := config.GetDB()

	if err := autoMigrate(db); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}
	if err := normalizeLegacySchema(db); err != nil {
		log.Fatalf("normalize legacy schema: %v", err)
	}

	ctx := &seedContext{}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := cleanData(tx); err != nil {
			return err
		}
		if err := seedBaseAccounts(tx, ctx); err != nil {
			return err
		}
		if err := seedClubSystem(tx, ctx); err != nil {
			return err
		}
		if err := seedWeeklyReports(tx, ctx); err != nil {
			return err
		}
		if err := seedMatches(tx, ctx); err != nil {
			return err
		}
		if err := seedPhysicalTests(tx, ctx); err != nil {
			return err
		}
		if err := seedOrdersAndReports(tx, ctx); err != nil {
			return err
		}
		if err := seedScoutData(tx, ctx); err != nil {
			return err
		}
		if err := seedSocialAndAdmin(tx, ctx); err != nil {
			return err
		}
		return validateData(tx)
	}); err != nil {
		log.Fatalf("seed demo data failed, database restored from transaction rollback. backup: %s error: %v", backupPath, err)
	}

	if err := printStats(db); err != nil {
		log.Fatalf("print stats: %v", err)
	}
	log.Println("seed_demo_v2 completed successfully")
}

func backupDatabase(dbPath string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	backupPath := strings.TrimSuffix(dbPath, filepath.Ext(dbPath)) + ".backup-demo-v2-" + timestamp + filepath.Ext(dbPath)
	src, err := os.Open(dbPath)
	if err != nil {
		return "", err
	}
	defer src.Close()
	dst, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}
	return backupPath, dst.Sync()
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{}, &models.Player{}, &models.Analyst{}, &models.AnalystApplication{},
		&models.Club{}, &models.ClubPlayer{}, &models.ClubOrder{}, &models.Coach{}, &models.FootballExperience{}, &models.CoachFollowPlayer{}, &models.TrainingNote{},
		&models.Team{}, &models.TeamPlayer{}, &models.ClubCoach{}, &models.TeamCoach{}, &models.TeamInvitation{}, &models.ClubInvitation{}, &models.TeamApplication{},
		&models.PhysicalTestTemplateCustom{}, &models.PhysicalTestActivity{}, &models.PhysicalTestRecord{}, &models.PhysicalTestReport{},
		&models.ClubHome{}, &models.Achievement{}, &models.ClubHomeTeam{}, &models.ClubHomeCoach{}, &models.ClubHomePlayer{}, &models.ClubActivity{}, &models.ClubActivityRegistration{},
		&models.WeeklyReport{}, &models.WeeklyReportPeriod{}, &models.MatchSummary{}, &models.PlayerReview{}, &models.MatchVideo{},
		&models.Order{}, &models.Report{}, &models.VideoAnalysis{}, &models.AnalysisHighlight{},
		&models.Scout{}, &models.ScoutFollowPlayer{}, &models.ScoutReport{}, &models.ScoutTask{}, &models.PlayerFilterPreset{}, &models.PlayerShortlist{}, &models.TrialInvite{},
		&models.Notification{}, &models.Comment{}, &models.Like{}, &models.Favorite{}, &models.Post{}, &models.SocialAchievement{}, &models.UserSocialAchievement{}, &models.UserStats{}, &models.Follow{}, &models.GrowthRecord{},
		&models.TrainingPlan{}, &models.MatchSchedule{}, &models.TeamSeasonArchive{}, &models.AdminOperationLog{}, &models.Message{}, &models.ContentReport{}, &models.SensitiveWord{}, &models.PlatformAnnouncement{}, &models.Banner{}, &models.FAQ{}, &models.LoginLog{}, &models.Announcement{},
	)
}

func normalizeLegacySchema(db *gorm.DB) error {
	for _, model := range []any{&models.TeamCoach{}, &models.TeamPlayer{}} {
		if err := db.Migrator().DropTable(model); err != nil {
			return err
		}
		if err := db.AutoMigrate(model); err != nil {
			return err
		}
	}
	return nil
}

func cleanData(tx *gorm.DB) error {
	tables := []string{
		"analysis_highlights", "video_analyses", "reports", "orders", "club_orders",
		"match_videos", "player_reviews", "match_summaries", "weekly_report_periods", "weekly_reports",
		"physical_test_reports", "physical_test_records", "physical_test_activities", "physical_test_template_customs",
		"training_notes", "coach_follow_players", "football_experiences", "team_coaches", "club_coaches", "team_players", "team_invitations", "club_invitations", "team_applications", "teams",
		"club_home_players", "club_home_coaches", "club_home_teams", "club_homes", "achievements", "club_activity_registrations", "club_activities", "announcements",
		"scout_reports", "scout_follow_players", "scout_tasks", "scouts", "player_filter_presets", "player_shortlists", "trial_invites",
		"likes", "favorites", "comments", "posts", "follows", "notifications", "user_social_achievements", "social_achievements", "user_social_stats", "growth_records", "messages",
		"training_plans", "match_schedules", "team_season_archives", "admin_operation_logs", "content_reports", "sensitive_words", "platform_announcements", "banners", "faqs", "login_logs",
		"club_players", "clubs", "analyst_applications", "analysts", "coaches", "players", "users",
	}
	if err := tx.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
		return err
	}
	for _, table := range tables {
		if err := tx.Exec("DELETE FROM " + table).Error; err != nil {
			return fmt.Errorf("clean table %s: %w", table, err)
		}
	}
	return tx.Exec("PRAGMA foreign_keys = ON").Error
}

func seedBaseAccounts(tx *gorm.DB, ctx *seedContext) error {
	now := time.Now()
	ctx.admin = newUser("13800000001", "平台管理员", "小球探运营官", models.RoleAdmin, "上海", "上海", now)
	ctx.clubOwner = newUser("13800000010", "上海绿地青训俱乐部", "绿地青训", models.RoleClub, "上海", "上海", now)
	ctx.backupClub = newUser("13800000011", "北京晨星足球学院", "晨星足球", models.RoleClub, "北京", "北京", now)
	users := []*models.User{&ctx.admin, &ctx.clubOwner, &ctx.backupClub}

	coachSeeds := []coachSeed{
		{"13800000020", "王振宇", "王指导", "B级", models.CoachRoleHead, 0, []string{"技术训练", "比赛阅读", "青少年培养"}, []string{"U12", "U14"}, 9},
		{"13800000021", "李明轩", "李教练", "C级", models.CoachRoleFitness, 0, []string{"体能训练", "伤病预防", "速度敏捷"}, []string{"U10", "U12"}, 6},
		{"13800000022", "张凯", "张指导", "B级", models.CoachRoleHead, 1, []string{"战术组织", "防守训练", "心理建设"}, []string{"U14", "U16"}, 11},
	}
	for _, seed := range coachSeeds {
		u := newUser(seed.Phone, seed.Name, seed.Nickname, models.RoleCoach, "上海", "上海", now)
		ctx.coaches = append(ctx.coaches, u)
		users = append(users, &ctx.coaches[len(ctx.coaches)-1])
	}

	analystSeeds := []analystSeed{
		{"13800000030", "陈知远", "知远分析", "技术动作与边路进攻", "前职业梯队分析师", 8, 4.9, 128},
		{"13800000031", "林若然", "若然战术", "中场组织与比赛阅读", "青训比赛数据分析师", 6, 4.8, 93},
		{"13800000032", "周启航", "启航视频", "门将与防守体系", "职业俱乐部视频分析师", 7, 4.7, 76},
		{"13800000033", "吴嘉宁", "嘉宁球探", "潜力评估与发展建议", "校园足球观察员", 5, 4.6, 61},
	}
	for _, seed := range analystSeeds {
		u := newUser(seed.Phone, seed.Name, seed.Nickname, models.RoleAnalyst, "上海", "上海", now)
		ctx.analysts = append(ctx.analysts, u)
		users = append(users, &ctx.analysts[len(ctx.analysts)-1])
	}

	scoutSeeds := []scoutSeed{
		{"13800000024", "赵云帆", "云帆球探", []string{"华东", "上海", "江苏"}, []string{"边锋", "中前卫", "U12"}, "华东青训观察中心"},
		{"13800000025", "陈立青", "立青选材", []string{"华北", "北京", "山东"}, []string{"中后卫", "门将", "U14"}, "全国校园足球人才库"},
	}
	for _, seed := range scoutSeeds {
		u := newUser(seed.Phone, seed.Name, seed.Nickname, models.RoleScout, "上海", "上海", now)
		ctx.scouts = append(ctx.scouts, u)
		users = append(users, &ctx.scouts[len(ctx.scouts)-1])
	}

	playerSeeds := demoPlayers()
	for _, seed := range playerSeeds {
		u := newPlayerUser(seed, now)
		ctx.players = append(ctx.players, u)
		users = append(users, &ctx.players[len(ctx.players)-1])
	}

	for _, user := range users {
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("create user %s: %w", user.Phone, err)
		}
	}
	if err := tx.Where("phone = ?", ctx.admin.Phone).First(&ctx.admin).Error; err != nil {
		return err
	}
	if err := tx.Where("phone = ?", ctx.clubOwner.Phone).First(&ctx.clubOwner).Error; err != nil {
		return err
	}
	if err := tx.Where("phone = ?", ctx.backupClub.Phone).First(&ctx.backupClub).Error; err != nil {
		return err
	}
	for i := range ctx.coaches {
		if err := tx.Where("phone = ?", ctx.coaches[i].Phone).First(&ctx.coaches[i]).Error; err != nil {
			return err
		}
	}
	for i := range ctx.analysts {
		if err := tx.Where("phone = ?", ctx.analysts[i].Phone).First(&ctx.analysts[i]).Error; err != nil {
			return err
		}
	}
	for i := range ctx.scouts {
		if err := tx.Where("phone = ?", ctx.scouts[i].Phone).First(&ctx.scouts[i]).Error; err != nil {
			return err
		}
	}
	for i := range ctx.players {
		if err := tx.Where("phone = ?", ctx.players[i].Phone).First(&ctx.players[i]).Error; err != nil {
			return err
		}
	}

	for i, seed := range coachSeeds {
		coach := models.Coach{UserID: ctx.coaches[i].ID, LicenseType: seed.License, LicenseNumber: fmt.Sprintf("CFA-%04d", 202600+i), Specialties: mustJSON(seed.Specialties), Style: mustJSON([]string{"技术型", "青训专长型"}), AgeGroups: mustJSON(seed.AgeGroups), Bio: seed.Name + "长期深耕青少年足球训练，重视基本功、比赛理解和成长反馈。", CoachingYears: seed.CoachingYear, CurrentClub: "上海绿地青训俱乐部", City: "上海", Verified: true}
		if err := tx.Create(&coach).Error; err != nil {
			return err
		}
	}
	for i, seed := range analystSeeds {
		analyst := models.Analyst{UserID: ctx.analysts[i].ID, Name: seed.Name, Bio: seed.Name + "擅长将比赛视频拆解为可执行训练建议，报告表达清晰，适合家长和教练共同阅读。", Specialty: seed.Specialty, Experience: seed.Experience, Profession: seed.Profession, IsProPlayer: i == 0, HasCase: true, CaseDetail: "已完成多场青少年比赛技术评估与成长追踪案例。", ContactPhone: seed.Phone, ContactEmail: fmt.Sprintf("analyst%d@shaonianqiutan.demo", i+1), Rating: seed.Rating, ReviewCount: seed.ReviewCount, Status: models.AnalystStatusActive}
		if err := tx.Create(&analyst).Error; err != nil {
			return err
		}
		ctx.analystRows = append(ctx.analystRows, analyst)
	}
	for i, seed := range scoutSeeds {
		scout := models.Scout{UserID: ctx.scouts[i].ID, ScoutingExperience: "5-10", Specialties: mustJSON(seed.Specialties), PreferredAgeGroups: mustJSON([]string{"U12", "U14", "U16"}), ScoutingRegions: mustJSON(seed.Regions), CurrentOrganization: seed.Organization, Bio: seed.Name + "关注青少年球员长期潜力与比赛气质，偏好结合体测、视频和现场观察综合判断。", Verified: true, TotalDiscovered: 36 + i*12, TotalReports: 18 + i*9, TotalAdopted: 6 + i*3}
		if err := tx.Create(&scout).Error; err != nil {
			return err
		}
		ctx.scoutRows = append(ctx.scoutRows, scout)
	}
	for _, u := range ctx.players {
		row := models.Player{UserID: u.ID, Name: u.Name, Nickname: u.Nickname, Province: u.Province, City: u.City, District: "浦东新区", Position: u.Position, Age: u.Age, BirthDate: u.BirthDate, Height: int(u.Height), Weight: int(u.Weight), Foot: u.Foot, Club: "上海绿地青训俱乐部", School: u.School, Phone: u.Phone, Avatar: u.Avatar, VideoURL: u.VideoUrl, Status: 1}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		ctx.playerRows = append(ctx.playerRows, row)
	}
	return nil
}

func seedClubSystem(tx *gorm.DB, ctx *seedContext) error {
	now := time.Now()
	expire := now.AddDate(1, 0, 0)
	ctx.club = models.Club{UserID: ctx.clubOwner.ID, Name: "上海绿地青训俱乐部", Logo: avatar("上海绿地"), Description: "面向 U8-U16 年龄段的城市青训俱乐部，建立训练、比赛、体测、视频分析和成长档案一体化体系。", Address: "上海市浦东新区张江足球训练中心", ContactName: "周运营", ContactPhone: "13800000010", EstablishedYear: 2016, ClubSize: "large", MemberLevel: models.MemberLevelProfessional, MemberExpireDate: expire, FreeTestQuota: 80, Province: "上海", City: "上海"}
	ctx.backup = models.Club{UserID: ctx.backupClub.ID, Name: "北京晨星足球学院", Logo: avatar("北京晨星"), Description: "用于跨区域对照演示的青训机构。", Address: "北京市朝阳区奥体训练基地", ContactName: "刘老师", ContactPhone: "13800000011", EstablishedYear: 2018, ClubSize: "medium", MemberLevel: models.MemberLevelBasic, MemberExpireDate: expire, FreeTestQuota: 30, Province: "北京", City: "北京"}
	if err := tx.Create(&ctx.club).Error; err != nil {
		return err
	}
	if err := tx.Create(&ctx.backup).Error; err != nil {
		return err
	}

	birthStart12, birthEnd12 := 2014, 2015
	birthStart14, birthEnd14 := 2012, 2013
	ctx.teams = []models.Team{
		{ClubID: ctx.club.ID, Name: "U12 精英队", AgeGroup: "U12", BirthYearStart: &birthStart12, BirthYearEnd: &birthEnd12, Description: "以技术动作、控球推进和小组配合为核心的精英梯队。", Status: models.TeamStatusActive},
		{ClubID: ctx.club.ID, Name: "U14 梯队", AgeGroup: "U14", BirthYearStart: &birthStart14, BirthYearEnd: &birthEnd14, Description: "面向更高强度比赛，强化战术纪律、攻防转换和位置理解。", Status: models.TeamStatusActive},
	}
	for i := range ctx.teams {
		if err := tx.Create(&ctx.teams[i]).Error; err != nil {
			return err
		}
	}

	for i, coach := range ctx.coaches {
		role := []models.CoachRole{models.CoachRoleHead, models.CoachRoleFitness, models.CoachRoleHead}[i]
		if err := tx.Create(&models.ClubCoach{ClubID: ctx.club.ID, UserID: coach.ID, PrimaryRole: role, Status: models.ClubCoachStatusActive, JoinedAt: now.AddDate(-1, -i, 0), Notes: "演示数据：核心教练组成员"}).Error; err != nil {
			return err
		}
		teamIndex := 0
		if i == 2 {
			teamIndex = 1
		}
		invitedBy := ctx.clubOwner.ID
		if err := tx.Create(&models.TeamCoach{TeamID: ctx.teams[teamIndex].ID, UserID: coach.ID, Role: role, Status: "active", InvitedBy: &invitedBy, JoinedAt: now.AddDate(-1, -i, 0)}).Error; err != nil {
			return err
		}
	}

	for i, player := range ctx.players {
		seed := demoPlayers()[i]
		team := ctx.teams[seed.TeamIndex]
		if err := tx.Create(&models.ClubPlayer{ClubID: ctx.club.ID, UserID: player.ID, JoinDate: now.AddDate(-1, -i%6, 0), AgeGroup: team.AgeGroup, Position: player.Position, Tags: mustJSON(seed.Tags), Status: "active", Notes: "演示数据：资料完整，可用于球队、周报、比赛和体测链路。"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.TeamPlayer{TeamID: team.ID, UserID: player.ID, JerseyNumber: fmt.Sprintf("%d", seed.JerseyNumber), Position: player.Position, Status: "active", Source: "invited", JoinedAt: now.AddDate(-1, -i%6, 0), Notes: "主力轮换球员"}).Error; err != nil {
			return err
		}
	}

	home := models.DefaultClubHome(ctx.club.ID)
	home.Hero.Title = "上海绿地青训俱乐部"
	home.Hero.Subtitle = "用数据连接训练、比赛与成长"
	home.About.Content = "我们以科学训练、比赛复盘、体测追踪和视频分析为基础，为青少年球员建立长期成长档案。"
	home.Contact.Phone = ctx.club.ContactPhone
	home.Contact.Address = ctx.club.Address
	home.Facilities.Enabled = true
	home.Facilities.Description = "2 片 11 人制天然草训练场、4 片 8 人制灯光球场、体能评估室和视频复盘教室。"
	home.Facilities.Schedule = []models.ClubHomeScheduleItem{{Day: "周二/周四", TimeRange: "18:30-20:00", Group: "U12 精英队"}, {Day: "周三/周六", TimeRange: "18:30-20:30", Group: "U14 梯队"}}
	home.Recruitment.Enabled = true
	home.Recruitment.Title = "2026 春季精英梯队试训"
	home.Recruitment.Description = "招募 2012-2015 年出生、有稳定训练基础的球员。"
	home.Recruitment.TrialDate = now.AddDate(0, 0, 14).Format("2006-01-02")
	home.Recruitment.ContactPhone = ctx.club.ContactPhone
	if err := tx.Create(home).Error; err != nil {
		return err
	}
	for i, team := range ctx.teams {
		if err := tx.Create(&models.ClubHomeTeam{ClubID: ctx.club.ID, TeamID: team.ID, Sort: i + 1, ShowPlayerCount: true}).Error; err != nil {
			return err
		}
	}
	for i, coach := range ctx.coaches {
		if err := tx.Create(&models.ClubHomeCoach{ClubID: ctx.club.ID, CoachID: coach.ID, Sort: i + 1}).Error; err != nil {
			return err
		}
	}
	for i := 0; i < 4; i++ {
		if err := tx.Create(&models.ClubHomePlayer{ClubID: ctx.club.ID, PlayerID: ctx.players[i].ID, Sort: i + 1, RecommendText: "训练态度稳定，近期成长曲线明显。"}).Error; err != nil {
			return err
		}
	}

	achievements := []models.Achievement{{ClubID: ctx.club.ID, Title: "市青少年联赛冠军", Description: "U12 组别赛季不败夺冠", Icon: "trophy", Count: "1", Sort: 1}, {ClubID: ctx.club.ID, Title: "注册球员", Description: "长期训练球员规模", Icon: "users", Count: "128", Sort: 2}, {ClubID: ctx.club.ID, Title: "专业教练", Description: "持证教练与专项教练", Icon: "badge-check", Count: "12", Sort: 3}}
	if err := tx.Create(&achievements).Error; err != nil {
		return err
	}

	activities := []models.ClubActivity{
		{ClubID: ctx.club.ID, Title: "春季公开试训日", Type: "external", Status: "upcoming", Description: "面向 U10-U14 球员开放，包含基础技术、体能和小场对抗评估。", CoverImage: imageURL("trial"), StartTime: now.AddDate(0, 0, 15), EndTime: now.AddDate(0, 0, 15).Add(3 * time.Hour), Location: ctx.club.Address, MaxParticipants: 48, ContactPhone: ctx.club.ContactPhone, ContactWechat: "SQ-GREENLAND", PublishStatus: "published"},
		{ClubID: ctx.club.ID, Title: "冬训营成长回顾", Type: "internal", Status: "ended", Description: "为期 5 天的冬训营，围绕控球、转换和比赛阅读进行主题训练。", CoverImage: imageURL("camp"), StartTime: now.AddDate(0, -2, 0), EndTime: now.AddDate(0, -2, 5), Location: ctx.club.Address, MaxParticipants: 36, ContactPhone: ctx.club.ContactPhone, PublishStatus: "published", IsReview: true, ReviewContent: "球员在传接球速度和高压下决策方面提升明显。", ReviewImages: mustJSON([]string{imageURL("camp-1"), imageURL("camp-2")})},
	}
	if err := tx.Create(&activities).Error; err != nil {
		return err
	}
	if err := tx.Create(&models.ClubActivityRegistration{ActivityID: activities[0].ID, UserID: &ctx.players[0].ID, Name: ctx.players[0].Name, Phone: ctx.players[0].Phone, Wechat: "player2001", Remark: "希望参加 U12 试训", Status: "confirmed"}).Error; err != nil {
		return err
	}

	if err := tx.Create(&models.Announcement{ClubID: ctx.club.ID, Title: "四月训练重点：高压下第一脚处理", Content: "本月 U12/U14 梯队将围绕接球前观察、第一脚方向和弱侧转移进行专项训练。", IsPinned: true, CreatedBy: ctx.clubOwner.ID}).Error; err != nil {
		return err
	}

	accepted := now.AddDate(0, 0, -2)
	rejected := now.AddDate(0, 0, -1)
	invitations := []models.TeamInvitation{
		{TeamID: ctx.teams[0].ID, ClubID: ctx.club.ID, Type: models.InvitationTypePlayer, InviteCode: "SC-U12-PENDING-001", TargetPhone: "13800002901", Status: models.InvitationStatusPending, CreatedBy: ctx.clubOwner.ID, ExpiresAt: now.AddDate(0, 0, 7)},
		{TeamID: ctx.teams[0].ID, ClubID: ctx.club.ID, Type: models.InvitationTypePlayer, InviteCode: "SC-U12-ACCEPTED-001", TargetUserID: &ctx.players[0].ID, TargetPhone: ctx.players[0].Phone, Status: models.InvitationStatusAccepted, CreatedBy: ctx.clubOwner.ID, ExpiresAt: now.AddDate(0, 0, 7), AcceptedAt: &accepted},
		{TeamID: ctx.teams[1].ID, ClubID: ctx.club.ID, Type: models.InvitationTypeCoach, InviteCode: "SC-U14-REJECTED-001", TargetPhone: "13800002902", Status: models.InvitationStatusRejected, CreatedBy: ctx.clubOwner.ID, ExpiresAt: now.AddDate(0, 0, 7), RejectedAt: &rejected, RejectedReason: "时间安排冲突"},
		{TeamID: ctx.teams[1].ID, ClubID: ctx.club.ID, Type: models.InvitationTypePlayer, InviteCode: "SC-U14-EXPIRED-001", TargetPhone: "13800002903", Status: models.InvitationStatusExpired, CreatedBy: ctx.clubOwner.ID, ExpiresAt: now.AddDate(0, 0, -1)},
	}
	if err := tx.Create(&invitations).Error; err != nil {
		return err
	}
	clubInvites := []models.ClubInvitation{
		{ClubID: ctx.club.ID, Type: models.InvitationTypeCoach, InviteCode: "CLUB-COACH-PENDING-001", TargetPhone: "13800002911", TargetRole: models.CoachRoleAssistant, Status: models.InvitationStatusPending, CreatedBy: ctx.clubOwner.ID, ExpiresAt: now.AddDate(0, 0, 7)},
		{ClubID: ctx.club.ID, Type: models.InvitationTypeCoach, InviteCode: "CLUB-COACH-ACCEPTED-001", TargetUserID: &ctx.coaches[1].ID, TargetPhone: ctx.coaches[1].Phone, TargetRole: models.CoachRoleFitness, Status: models.InvitationStatusAccepted, CreatedBy: ctx.clubOwner.ID, ExpiresAt: now.AddDate(0, 0, 7), AcceptedAt: &accepted},
	}
	return tx.Create(&clubInvites).Error
}

func seedWeeklyReports(tx *gorm.DB, ctx *seedContext) error {
	now := time.Now()
	monday := startOfWeek(now).AddDate(0, 0, -7)
	statuses := []struct{ submit, review string }{{"draft", "pending"}, {"submitted", "pending"}, {"submitted", "approved"}, {"submitted", "rejected"}}
	for i, player := range ctx.players[:8] {
		status := statuses[i%len(statuses)]
		team := ctx.teams[0]
		coach := ctx.coaches[0]
		if i >= 6 {
			team = ctx.teams[1]
			coach = ctx.coaches[2]
		}
		deadline := monday.AddDate(0, 0, 6).Add(20 * time.Hour)
		var submittedAt, reviewedAt *time.Time
		if status.submit == "submitted" {
			t := monday.AddDate(0, 0, 5).Add(time.Duration(i) * time.Hour)
			submittedAt = &t
		}
		if status.review != "pending" {
			t := monday.AddDate(0, 0, 6).Add(time.Duration(i) * time.Hour)
			reviewedAt = &t
		}
		report := models.WeeklyReport{TeamID: team.ID, PlayerID: player.ID, CoachID: coach.ID, WeekStart: monday, WeekEnd: monday.AddDate(0, 0, 6), Deadline: &deadline, TrainingCount: 3 + i%2, TrainingDuration: 270 + i*10, KnowledgeSummary: "本周重点复习接球前观察、第一脚处理和弱侧转移。", TechnicalContent: "传接球、带球摆脱、小组二过一。", TacticalContent: "前场压迫触发点与中场保护距离。", PhysicalCondition: "速度敏捷训练完成度较好。", MatchPerformance: "队内对抗中参与度提升。", SelfAttitudeRating: 4, SelfTechniqueRating: 3 + i%2, SelfTeamworkRating: 4, ImprovementsDetail: "接球前观察次数增加，处理球更果断。", Weaknesses: "高压下逆足传球稳定性还需加强。", FatigueLevel: 2 + i%3, SleepQuality: 4, DietCondition: "饮食规律，训练日前补水充足。", MessageToCoach: "希望下周增加射门前调整训练。", SubmitStatus: status.submit, SubmittedAt: submittedAt, ReviewStatus: status.review, ReviewCoachID: coach.ID, ReviewComment: "态度积极，训练反馈真实，建议继续提升弱侧处理能力。", StrengthsAcknowledgment: "跑动积极，团队配合意识较好。", Suggestions: "增加对抗下的一脚出球训练。", KnowledgeFeedback: "对压迫触发点理解基本准确。", NextWeekFocus: "第一脚方向和无球接应线路。", RecommendAward: i == 2, ReviewedAt: reviewedAt}
		if status.review == "rejected" {
			report.ReviewComment = "内容偏简单，需要补充训练知识点和具体比赛场景。"
		}
		if err := tx.Create(&report).Error; err != nil {
			return err
		}
	}
	periodDeadline := monday.AddDate(0, 0, 6).Add(20 * time.Hour)
	period := models.WeeklyReportPeriod{TeamID: ctx.teams[0].ID, WeekStart: monday, WeekEnd: monday.AddDate(0, 0, 6), Deadline: &periodDeadline, TotalPlayers: 6, SubmittedCount: 5, PendingCount: 1, OverdueCount: 0, ReviewedCount: 3, Status: "active"}
	return tx.Create(&period).Error
}

func seedMatches(tx *gorm.DB, ctx *seedContext) error {
	now := time.Now()
	playerIDs := ids(ctx.players[:6])
	match := models.MatchSummary{TeamID: ctx.teams[0].ID, CoachID: ctx.coaches[0].ID, MatchName: "上海青少年精英邀请赛小组赛", MatchDate: now.AddDate(0, 0, -5).Format("2006-01-02"), Opponent: "浦东联队 U12", Location: "home", MatchFormat: "8人制", OurScore: 3, OppScore: 1, Result: "win", CoverImage: imageURL("match-u12"), CoachOverall: "整体执行了赛前布置的边路推进策略，前场压迫质量较高。", CoachTactic: "上半场通过右路快速推进制造机会，下半场加强中路保护。", CoachKeyMoments: "第18分钟边路连续配合后的倒三角传中是本场最佳团队配合。", Status: "completed"}
	match.SetPlayerIDs(playerIDs)
	if err := tx.Create(&match).Error; err != nil {
		return err
	}
	if err := tx.Create(&models.MatchVideo{MatchID: match.ID, TeamID: ctx.teams[0].ID, UploaderID: ctx.coaches[0].ID, Platform: "demo", URL: "https://example.com/videos/u12-match.mp4", Name: "全场录像", Note: "演示视频链接", SortOrder: 1, Status: "active"}).Error; err != nil {
		return err
	}
	for i, playerID := range playerIDs[:4] {
		review := models.PlayerReview{MatchID: match.ID, PlayerID: playerID, TeamID: ctx.teams[0].ID, Performance: "good", Goals: boolInt(i == 0), Assists: boolInt(i == 1), Tactics: mustJSON([]string{"边路压迫", "快速回防"}), Highlights: "在高压下完成多次有效接应。", Improvements: "需要提升弱脚处理球速度。", NextGoals: "下场比赛争取增加前插跑动。", CoachRating: 4.2 + float64(i)/10, CoachComment: "比赛投入度高，执行力稳定。", CoachReply: "继续保持主动要球，注意观察身后空间。", Status: "reviewed", SubmittedAt: now.AddDate(0, 0, -4)}
		if err := tx.Create(&review).Error; err != nil {
			return err
		}
	}
	pending := models.MatchSummary{TeamID: ctx.teams[1].ID, CoachID: ctx.coaches[2].ID, MatchName: "U14 周末教学赛", MatchDate: now.AddDate(0, 0, 3).Format("2006-01-02"), Opponent: "闵行青训 U14", Location: "away", MatchFormat: "11人制", Status: "pending", CoverImage: imageURL("match-u14")}
	pending.SetPlayerIDs(ids(ctx.players[6:12]))
	return tx.Create(&pending).Error
}

func seedPhysicalTests(tx *gorm.DB, ctx *seedContext) error {
	now := time.Now()
	playerIDs := ids(ctx.players)
	activity := models.PhysicalTestActivity{ClubID: ctx.club.ID, Name: "2026 春季基础体测", Description: "覆盖速度、爆发、柔韧和核心力量，用于建立新赛季成长基线。", StartDate: now.AddDate(0, 0, -10), EndDate: ptrTime(now.AddDate(0, 0, -9)), Location: ctx.club.Address, Template: models.PTTemplateAdvanced, PlayerIDs: mustJSON(playerIDs), Status: models.PTStatusReported, NotifyParents: true, AutoSendReport: true, CreatedBy: ctx.clubOwner.ID}
	if err := tx.Create(&activity).Error; err != nil {
		return err
	}
	for i, player := range ctx.players {
		height := player.Height
		weight := player.Weight
		bmi := weight / ((height / 100) * (height / 100))
		sprint := 4.45 + float64(i%5)*0.08
		sprint50m := 7.65 + float64(i%5)*0.12
		sprint100m := 15.20 + float64(i%5)*0.28
		agilityLadder := 9.75 + float64(i%5)*0.18
		tTest := 11.05 + float64(i%5)*0.20
		shuttleRun := 11.70 + float64(i%5)*0.18
		jump := 188.0 + float64(i%6)*6
		verticalJump := 36.0 + float64(i%6)*2
		reach := 9.0 + float64(i%5)
		push := 22 + i%8
		sit := 36 + i%10
		plank := 80 + i%6*8
		record := models.PhysicalTestRecord{
			ActivityID: activity.ID, PlayerID: player.ID, ClubID: ctx.club.ID,
			TestDate: now.AddDate(0, 0, -9), Height: &height, Weight: &weight, BMI: &bmi,
			Sprint30m: &sprint, Sprint50m: &sprint50m, Sprint100m: &sprint100m,
			AgilityLadder: &agilityLadder, TTest: &tTest, ShuttleRun: &shuttleRun,
			StandingLongJump: &jump, VerticalJump: &verticalJump, SitAndReach: &reach,
			PushUp: &push, SitUp: &sit, Plank: &plank, RecorderID: ctx.coaches[1].ID,
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		reportData := models.PhysicalTestReportData{PlayerName: player.Name, PlayerAge: player.Age, PlayerAgeGroup: ctx.teams[demoPlayers()[i].TeamIndex].AgeGroup, Position: player.Position, TestDate: now.AddDate(0, 0, -9).Format("2006-01-02"), OverallRating: "良好", Percentile: 70 + i%15, Strengths: []string{"速度启动积极", "核心力量稳定"}, Improvements: []string{"柔韧性", "反复冲刺恢复"}, TrainingSuggestions: []string{"每周增加 2 次变向加速训练"}, NutritionSuggestions: []string{"训练后 30 分钟内补充碳水和蛋白"}, RestSuggestions: []string{"保证 8 小时以上睡眠"}, NextTestSuggestion: "6 周后复测速度与爆发指标"}
		if err := tx.Create(&models.PhysicalTestReport{RecordID: record.ID, PlayerID: player.ID, ClubID: ctx.club.ID, ActivityID: activity.ID, ReportData: mustJSON(reportData), PDFURL: fmt.Sprintf("/demo/reports/physical-%d.pdf", player.ID), ShareToken: fmt.Sprintf("demo-physical-%d", player.ID), ShareCount: i % 4, ViewCount: 8 + i}).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedOrdersAndReports(tx *gorm.DB, ctx *seedContext) error {
	now := time.Now()
	statuses := []models.OrderStatus{models.OrderStatusPending, models.OrderStatusPaid, models.OrderStatusUploaded, models.OrderStatusAssigned, models.OrderStatusProcessing, models.OrderStatusCompleted, models.OrderStatusCancelled}
	for i, status := range statuses {
		player := ctx.players[i%len(ctx.players)]
		analyst := ctx.analystRows[i%len(ctx.analystRows)]
		var analystID *uint
		if status == models.OrderStatusAssigned || status == models.OrderStatusProcessing || status == models.OrderStatusCompleted {
			analystID = &analyst.ID
		}
		paidAt := ptrTime(now.AddDate(0, 0, -8+i))
		if status == models.OrderStatusPending || status == models.OrderStatusCancelled {
			paidAt = nil
		}
		order := models.Order{UserID: player.ID, AnalystID: analystID, OrderNo: fmt.Sprintf("SQTD20260424%03d", i+1), Amount: 299 + float64(i%3)*200, Status: status, PaymentMethod: models.PaymentMethodWechat, PaymentTime: paidAt, PaidAt: paidAt, VideoURL: "https://example.com/videos/demo-order.mp4", VideoFilename: "demo-match.mp4", Remark: "演示订单：覆盖不同状态", OrderType: "video", PlayerName: player.Name, PlayerAge: player.Age, PlayerPosition: player.Position, JerseyColor: "绿色", JerseyNumber: fmt.Sprintf("%d", player.JerseyNumber), MatchName: "青少年邀请赛", Opponent: "浦东联队", VideoDuration: 4200, Deadline: ptrTime(now.AddDate(0, 0, 3)), AssignedAt: ptrTime(now.AddDate(0, 0, -2)), AcceptedAt: ptrTime(now.AddDate(0, 0, -1))}
		if status == models.OrderStatusCancelled {
			order.CancelReason = "用户取消演示订单"
		}
		if status == models.OrderStatusCompleted {
			order.CompletedAt = ptrTime(now.AddDate(0, 0, -1))
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		if status == models.OrderStatusCompleted || status == models.OrderStatusProcessing {
			reportStatus := models.ReportStatusProcessing
			if status == models.OrderStatusCompleted {
				reportStatus = models.ReportStatusCompleted
			}
			report := models.Report{OrderID: order.ID, UserID: player.ID, AnalystID: analyst.ID, PlayerName: player.Name, PlayerBirthDate: player.BirthDate, PlayerPosition: player.Position, PlayerProvince: player.Province, PlayerCity: player.City, Content: "演示报告：球员在边路推进中展现出较好的启动速度和观察意识。", PdfURL: fmt.Sprintf("/demo/reports/order-%d.pdf", order.ID), Status: reportStatus, OverallRating: 82.5, OffenseRating: 84, DefenseRating: 76, Summary: "具备较好发展潜力，建议继续强化弱脚处理和对抗下决策。", Strengths: mustJSON([]string{"启动速度快", "接应意识好", "训练态度稳定"}), Weaknesses: mustJSON([]string{"逆足传球稳定性", "高压下抬头观察"}), Suggestions: "每周增加专项弱脚传球和小范围对抗决策训练。", Potential: "high", ClipVideoURL: "https://example.com/videos/demo-clip.mp4", RatingDetails: mustJSON(map[string]int{"speed": 86, "technique": 81, "decision": 78})}
			if err := tx.Create(&report).Error; err != nil {
				return err
			}
			if err := tx.Model(&order).Update("report_id", report.ID).Error; err != nil {
				return err
			}
			analysis := models.VideoAnalysis{OrderID: order.ID, AnalystID: analyst.ID, UserID: player.ID, PlayerName: player.Name, PlayerAge: player.Age, PlayerPosition: player.Position, PlayerFoot: player.Foot, PlayerHeight: player.Height, PlayerWeight: player.Weight, PlayerTeam: player.CurrentTeam, MatchName: order.MatchName, MatchDate: now.AddDate(0, 0, -7).Format("2006-01-02"), MatchType: "正式比赛", OpponentLevel: "中上", Opponent: order.Opponent, PlayTime: 62, Goals: boolInt(i%2 == 0), Assists: boolInt(i%3 == 0), VideoURL: order.VideoURL, OverallScore: 82.5, PotentialLevel: models.PotentialA, Scores: mustJSON(map[string]float64{"firstTouch": 8.1, "passing": 8.0, "speed": 8.6, "decision": 7.8}), Summary: report.Summary, Improvements: report.Suggestions, AnalystNotes: "建议教练结合训练计划跟踪 4 周。", AIReport: "# AI 视频分析报告\n\n该球员具备较好的边路推进能力。", AIReportStatus: "confirmed", AIReportVersion: 1, Status: models.AnalysisStatusCompleted, RatingReportMD: "/demo/md/rating.md", PlayerInfoMD: "/demo/md/player.md"}
			if err := tx.Create(&analysis).Error; err != nil {
				return err
			}
			highlights := []models.AnalysisHighlight{{AnalysisID: analysis.ID, Timestamp: "12:30", TagType: models.HighlightDribble, Description: "边路一对一突破后完成传中。", VideoClipURL: "https://example.com/videos/clip-1.mp4", IncludeInReport: true, SortOrder: 1}, {AnalysisID: analysis.ID, Timestamp: "38:12", TagType: models.HighlightPass, Description: "中场转移找到弱侧空当。", VideoClipURL: "https://example.com/videos/clip-2.mp4", IncludeInReport: true, SortOrder: 2}}
			if err := tx.Create(&highlights).Error; err != nil {
				return err
			}
		}
	}
	for i, player := range ctx.players[:4] {
		order := models.ClubOrder{ClubID: ctx.club.ID, UserID: ctx.clubOwner.ID, OrderNo: fmt.Sprintf("CLUB20260424%03d", i+1), PlayerID: player.ID, AnalystID: ctx.analystRows[i%len(ctx.analystRows)].ID, ServiceType: []string{"quick_report", "full_report", "video_analysis", "quick_report"}[i], Price: 399, Discount: 0.9, FinalPrice: 359.1, Status: []string{"pending", "paid", "processing", "completed"}[i], Remark: "俱乐部批量下单演示"}
		if order.Status != "pending" {
			order.PaidAt = ptrTime(now.AddDate(0, 0, -3))
		}
		if order.Status == "completed" {
			order.CompletedAt = ptrTime(now.AddDate(0, 0, -1))
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedScoutData(tx *gorm.DB, ctx *seedContext) error {
	now := time.Now()
	for i, scout := range ctx.scoutRows {
		if err := tx.Create(&models.ScoutTask{Title: fmt.Sprintf("%s 区域 U%d 潜力边路球员观察", []string{"华东", "华北"}[i], 12+i*2), Description: "观察目标球员在正式比赛中的速度、技术稳定性和比赛气质。", Region: []string{"华东", "华北"}[i], AgeGroup: []string{"U12", "U14"}[i], Reward: 800 + i*300, Status: []string{"accepted", "open"}[i], Deadline: now.AddDate(0, 0, 14), AcceptedBy: &scout.ID}).Error; err != nil {
			return err
		}
	}
	for i, row := range ctx.playerRows[:8] {
		scout := ctx.scoutRows[i%len(ctx.scoutRows)]
		if err := tx.Create(&models.ScoutFollowPlayer{ScoutID: scout.ID, UserID: row.UserID, IsStarred: i%3 == 0, Notes: "演示关注：重点观察比赛气质和位置适应性。", FollowedAt: now.AddDate(0, 0, -i)}).Error; err != nil {
			return err
		}
		published := now.AddDate(0, 0, -i)
		report := models.ScoutReport{ScoutID: scout.ID, PlayerID: row.ID, OverallRating: 78 + i%12, PotentialRating: []string{"A", "B", "S"}[i%3], Status: []string{"draft", "published", "adopted"}[i%3], Strengths: mustJSON([]string{"速度启动快", "对抗意愿强"}), Weaknesses: mustJSON([]string{"逆足稳定性", "防守选位"}), TechnicalSkills: mustJSON(map[string]int{"shooting": 78, "passing": 82, "dribbling": 84, "defending": 70, "physical": 86, "mentality": 80}), Summary: "具备进入更高水平梯队观察名单的潜力。", Recommendation: "建议安排 2 周试训，重点观察高压对抗下的处理球质量。", TargetClub: ctx.club.Name, Views: 20 + i*3, Likes: 3 + i, PublishedAt: &published}
		if report.Status == "draft" {
			report.PublishedAt = nil
		}
		if report.Status == "adopted" {
			report.AdoptedAt = ptrTime(now.AddDate(0, 0, -1))
		}
		if err := tx.Create(&report).Error; err != nil {
			return err
		}
	}
	if err := tx.Create(&models.PlayerShortlist{ClubID: ctx.club.ID, PlayerID: ctx.playerRows[0].ID, Note: "边路推进能力突出，建议试训。", CreatedBy: ctx.clubOwner.ID}).Error; err != nil {
		return err
	}
	return tx.Create(&models.TrialInvite{SenderID: ctx.clubOwner.ID, PlayerID: ctx.players[0].ID, TrialDate: now.AddDate(0, 0, 10).Format("2006-01-02"), TrialTime: "15:00", Location: ctx.club.Address, ContactName: ctx.club.ContactName, ContactPhone: ctx.club.ContactPhone, Note: "请携带训练装备并提前 20 分钟到场。", Status: models.TrialInvitePending}).Error
}

func seedSocialAndAdmin(tx *gorm.DB, ctx *seedContext) error {
	now := time.Now()
	achievements := []models.SocialAchievement{{ID: models.AchievementFirstReport, Name: "首份报告", Description: "完成第一份球探报告", Icon: "file-check", Category: "report", Condition: "report_count", Threshold: 1}, {ID: models.AchievementFirstFollower, Name: "首次关注", Description: "获得第一位关注者", Icon: "user-plus", Category: "social", Condition: "followers", Threshold: 1}, {ID: models.AchievementLikes100, Name: "百赞时刻", Description: "累计获得 100 次点赞", Icon: "thumbs-up", Category: "social", Condition: "likes", Threshold: 100}}
	if err := tx.Create(&achievements).Error; err != nil {
		return err
	}
	posts := []models.Post{{UserID: ctx.clubOwner.ID, Content: "U12 精英队本周完成高压出球专项训练，球员第一脚处理明显更果断。", Images: mustJSON([]string{imageURL("post-training")}), TargetType: "club", TargetID: ctx.club.ID, RoleTag: "club", LikesCount: 8, CommentsCount: 2, IsTop: true}, {UserID: ctx.coaches[0].ID, Content: "比赛不是结果展示，而是训练反馈的放大镜。今天队员们在弱侧转移上进步明显。", RoleTag: "coach", LikesCount: 13, CommentsCount: 3}, {UserID: ctx.players[0].ID, Content: "今天体测 30 米比上次快了 0.12 秒，继续加油！", Images: mustJSON([]string{imageURL("post-player")}), RoleTag: "user", LikesCount: 16, CommentsCount: 4}}
	if err := tx.Create(&posts).Error; err != nil {
		return err
	}
	likes := []models.Like{{UserID: ctx.players[0].ID, TargetType: "post", TargetID: posts[0].ID}, {UserID: ctx.coaches[0].ID, TargetType: "post", TargetID: posts[2].ID}, {UserID: ctx.scouts[0].ID, TargetType: "post", TargetID: posts[2].ID}}
	if err := tx.Create(&likes).Error; err != nil {
		return err
	}
	comments := []models.Comment{{UserID: ctx.coaches[0].ID, TargetType: "post", TargetID: posts[2].ID, Content: "进步明显，继续保持训练记录。"}, {UserID: ctx.clubOwner.ID, TargetType: "post", TargetID: posts[1].ID, Content: "这条可以作为家长会复盘素材。"}}
	if err := tx.Create(&comments).Error; err != nil {
		return err
	}
	follows := []models.Follow{{FollowerID: ctx.scouts[0].ID, FollowingID: ctx.players[0].ID}, {FollowerID: ctx.players[0].ID, FollowingID: ctx.coaches[0].ID}, {FollowerID: ctx.clubOwner.ID, FollowingID: ctx.scouts[0].ID}}
	if err := tx.Create(&follows).Error; err != nil {
		return err
	}
	for _, user := range append([]models.User{ctx.clubOwner, ctx.coaches[0], ctx.players[0]}, ctx.analysts[0]) {
		if err := tx.Create(&models.UserStats{UserID: user.ID, LikesReceived: 20, FavoritesReceived: 6, CommentsReceived: 8, FollowersCount: 3, FollowingCount: 5, LoginStreak: 7, LastLoginDate: &now}).Error; err != nil {
			return err
		}
	}
	if err := tx.Create(&models.UserSocialAchievement{UserID: ctx.players[0].ID, AchievementID: models.AchievementFirstFollower, AchievedAt: &now}).Error; err != nil {
		return err
	}

	notifications := []models.Notification{
		{UserID: ctx.players[0].ID, Type: models.NotificationTypeWeeklyReportApproved, Title: "周报已通过", Content: "王振宇教练已审核你的本周训练周报。", Priority: 2},
		{UserID: ctx.coaches[0].ID, Type: models.NotificationTypeMatchCoachReminder, Title: "比赛点评待完成", Content: "U12 精英队有 2 名球员已提交比赛自评。", Priority: 2},
		{UserID: ctx.clubOwner.ID, Type: models.NotificationTypeActivityRegistration, Title: "新的试训报名", Content: "春季公开试训日收到新的报名。", Priority: 3},
		{UserID: ctx.analysts[0].ID, Type: models.NotificationTypeOrder, Title: "新订单已分配", Content: "你有一条视频分析订单需要处理。", Priority: 2},
	}
	if err := tx.Create(&notifications).Error; err != nil {
		return err
	}
	messages := []models.Message{{SenderID: ctx.coaches[0].ID, ReceiverID: ctx.players[0].ID, Content: "本周周报写得不错，下周重点看弱脚传球。", IsRead: false}, {SenderID: ctx.clubOwner.ID, ReceiverID: ctx.coaches[0].ID, Content: "周末家长沟通会请带上 U12 体测汇总。", IsRead: true}}
	if err := tx.Create(&messages).Error; err != nil {
		return err
	}

	plans := []models.TrainingPlan{{ClubID: ctx.club.ID, TeamID: ctx.teams[0].ID, Title: "U12 高压下第一脚处理", Theme: "技术+决策", Location: ctx.club.Address, StartTime: now.AddDate(0, 0, 1).Add(18 * time.Hour), EndTime: ptrTime(now.AddDate(0, 0, 1).Add(20 * time.Hour)), PlayerIDs: mustJSON(ids(ctx.players[:6])), Content: "热身、接球前观察、三色标志盘转移、小场对抗。", Summary: "用于演示训练计划。", CoachID: ctx.coaches[0].ID, Status: models.TrainingPlanStatusPublished, CreatedBy: ctx.coaches[0].ID}, {ClubID: ctx.club.ID, TeamID: ctx.teams[1].ID, Title: "U14 防守阵型与攻防转换", Theme: "战术", Location: ctx.club.Address, StartTime: now.AddDate(0, 0, 2).Add(18 * time.Hour), EndTime: ptrTime(now.AddDate(0, 0, 2).Add(20 * time.Hour)), PlayerIDs: mustJSON(ids(ctx.players[6:12])), Content: "防守站位、抢断后第一选择、快速反击。", CoachID: ctx.coaches[2].ID, Status: models.TrainingPlanStatusDraft, CreatedBy: ctx.coaches[2].ID}}
	if err := tx.Create(&plans).Error; err != nil {
		return err
	}
	homeScore, awayScore := 3, 1
	schedules := []models.MatchSchedule{{ClubID: ctx.club.ID, TeamID: ctx.teams[0].ID, Name: "上海青少年精英邀请赛", MatchType: models.MatchScheduleTypeLeague, Opponent: "浦东联队 U12", MatchTime: now.AddDate(0, 0, -5), Location: ctx.club.Address, HomeScore: &homeScore, AwayScore: &awayScore, Remark: "已完成比赛", Status: models.MatchScheduleStatusCompleted, CreatedBy: ctx.coaches[0].ID}, {ClubID: ctx.club.ID, TeamID: ctx.teams[1].ID, Name: "U14 周末教学赛", MatchType: models.MatchScheduleTypeFriendly, Opponent: "闵行青训 U14", MatchTime: now.AddDate(0, 0, 3), Location: "闵行体育公园", Status: models.MatchScheduleStatusUpcoming, CreatedBy: ctx.coaches[2].ID}}
	if err := tx.Create(&schedules).Error; err != nil {
		return err
	}

	growth := []models.GrowthRecord{{UserID: ctx.players[0].ID, RecordDate: now.AddDate(0, 0, -9), RecordType: models.GrowthRecordTypePhysical, Title: "春季体测完成", Content: "30 米冲刺成绩刷新个人最好成绩。", StatsJSON: mustJSON(map[string]any{"sprint30m": 4.45, "standingLongJump": 196})}, {UserID: ctx.players[0].ID, RecordDate: now.AddDate(0, 0, -5), RecordType: models.GrowthRecordTypeMatch, Title: "邀请赛贡献 1 球", Content: "边路突破后完成射门得分。", StatsJSON: mustJSON(map[string]any{"goals": 1, "assists": 0})}}
	if err := tx.Create(&growth).Error; err != nil {
		return err
	}

	adminLogs := []models.AdminOperationLog{{ClubID: ctx.club.ID, AdminID: ctx.admin.ID, AdminName: ctx.admin.Name, Action: "seed_demo_v2", Target: "database", Detail: "重建上线演示数据", IP: "127.0.0.1"}, {ClubID: ctx.club.ID, AdminID: ctx.clubOwner.ID, AdminName: ctx.clubOwner.Name, Action: "create_team", Target: "team", TargetID: ctx.teams[0].ID, Detail: "创建 U12 精英队演示数据", IP: "127.0.0.1"}}
	if err := tx.Create(&adminLogs).Error; err != nil {
		return err
	}
	handled := now.AddDate(0, 0, -1)
	contentReports := []models.ContentReport{{ReporterID: ctx.players[1].ID, ReporterName: ctx.players[1].Name, TargetID: posts[0].ID, TargetType: models.ContentReportTypePost, Reason: "演示举报：内容需核查", Detail: "用于管理员内容治理演示", Status: models.ContentReportStatusPending}, {ReporterID: ctx.coaches[0].ID, ReporterName: ctx.coaches[0].Name, TargetID: posts[2].ID, TargetType: models.ContentReportTypePost, Reason: "演示已处理举报", Status: models.ContentReportStatusResolved, HandlerID: ctx.admin.ID, HandlerName: ctx.admin.Name, HandleResult: "确认无违规，已关闭", HandledAt: &handled}}
	if err := tx.Create(&contentReports).Error; err != nil {
		return err
	}
	platform := []models.PlatformAnnouncement{{Title: "少年球探演示环境已更新", Content: "本环境已重建全角色演示数据，覆盖俱乐部、教练、球员、分析师、球探和管理员后台。", Type: "notice", IsPinned: true, IsPublic: true, CreatedBy: ctx.admin.ID, AuthorName: ctx.admin.Name, ViewCount: 128}, {Title: "春季青训数据观察", Content: "建议俱乐部结合体测、周报和比赛复盘进行阶段性培养评估。", Type: "article", IsPublic: true, CreatedBy: ctx.admin.ID, AuthorName: ctx.admin.Name, ViewCount: 86}}
	if err := tx.Create(&platform).Error; err != nil {
		return err
	}
	banners := []models.Banner{{Title: "上线演示主视觉", ImageURL: imageURL("banner-main"), LinkURL: "/", Position: "home", SortOrder: 1, Enabled: true, CreatedBy: ctx.admin.ID}, {Title: "俱乐部试训招募", ImageURL: imageURL("banner-trial"), LinkURL: "/clubs", Position: "club", SortOrder: 2, Enabled: true, CreatedBy: ctx.admin.ID}}
	if err := tx.Create(&banners).Error; err != nil {
		return err
	}
	faqs := []models.FAQ{{Question: "如何查看球员成长档案？", Answer: "球员登录后进入成长中心，可查看体测、周报、比赛和报告记录。", Category: "user", SortOrder: 1, Enabled: true, ViewCount: 42}, {Question: "俱乐部如何邀请教练？", Answer: "俱乐部管理员进入教练管理，可通过手机号或邀请码邀请教练加入。", Category: "club", SortOrder: 2, Enabled: true, ViewCount: 36}}
	if err := tx.Create(&faqs).Error; err != nil {
		return err
	}
	logs := []models.LoginLog{{UserID: ctx.clubOwner.ID, Phone: ctx.clubOwner.Phone, Nickname: ctx.clubOwner.Nickname, Role: string(ctx.clubOwner.Role), IP: "127.0.0.1", Device: "MacBook Pro", Browser: "Chrome", OS: "macOS", Location: "上海", Status: "success"}, {UserID: ctx.players[0].ID, Phone: ctx.players[0].Phone, Nickname: ctx.players[0].Nickname, Role: string(ctx.players[0].Role), IP: "127.0.0.1", Device: "iPhone", Browser: "Safari", OS: "iOS", Location: "上海", Status: "success"}}
	return tx.Create(&logs).Error
}

func validateData(tx *gorm.DB) error {
	checks := map[string]string{
		"teams_without_club":        "SELECT COUNT(*) FROM teams t LEFT JOIN clubs c ON c.id = t.club_id WHERE c.id IS NULL",
		"team_players_without_user": "SELECT COUNT(*) FROM team_players tp LEFT JOIN users u ON u.id = tp.user_id WHERE u.id IS NULL",
		"team_players_without_team": "SELECT COUNT(*) FROM team_players tp LEFT JOIN teams t ON t.id = tp.team_id WHERE t.id IS NULL",
		"team_coaches_without_user": "SELECT COUNT(*) FROM team_coaches tc LEFT JOIN users u ON u.id = tc.user_id WHERE u.id IS NULL",
		"club_players_without_user": "SELECT COUNT(*) FROM club_players cp LEFT JOIN users u ON u.id = cp.user_id WHERE u.id IS NULL",
		"weekly_without_player":     "SELECT COUNT(*) FROM weekly_reports wr LEFT JOIN users u ON u.id = wr.player_id WHERE u.id IS NULL",
		"matches_without_team":      "SELECT COUNT(*) FROM match_summaries ms LEFT JOIN teams t ON t.id = ms.team_id WHERE t.id IS NULL",
		"orders_without_user":       "SELECT COUNT(*) FROM orders o LEFT JOIN users u ON u.id = o.user_id WHERE u.id IS NULL",
		"reports_without_order":     "SELECT COUNT(*) FROM reports r LEFT JOIN orders o ON o.id = r.order_id WHERE o.id IS NULL",
	}
	for name, query := range checks {
		var count int64
		if err := tx.Raw(query).Scan(&count).Error; err != nil {
			return fmt.Errorf("validate %s: %w", name, err)
		}
		if count > 0 {
			return fmt.Errorf("validation failed %s: %d", name, count)
		}
	}
	return nil
}

func printStats(db *gorm.DB) error {
	tables := []string{"users", "clubs", "teams", "club_players", "team_players", "club_coaches", "team_coaches", "weekly_reports", "match_summaries", "physical_test_records", "orders", "reports", "video_analyses", "scout_reports", "posts", "notifications", "platform_announcements"}
	log.Println("demo data stats:")
	for _, table := range tables {
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			return err
		}
		log.Printf("  %-28s %d", table, count)
	}
	return nil
}

func newUser(phone, name, nickname string, role models.UserRole, province, city string, now time.Time) models.User {
	return models.User{Phone: phone, Password: demoPasswordHash, Nickname: nickname, Avatar: avatar(name), Role: role, CurrentRole: role, Status: models.StatusActive, Name: name, Gender: "男", Province: province, City: city, Country: "中国", NotificationSettings: `{"system":true,"order":true,"weekly":true,"social":true,"message":true,"email":false}`, PrivacySettings: `{"profileVisible":true,"phoneVisible":false,"searchable":true}`, CreatedAt: now, UpdatedAt: now}
}

func newPlayerUser(seed playerSeed, now time.Time) models.User {
	u := newUser(seed.Phone, seed.Name, seed.Nickname, models.RoleUser, seed.Province, seed.City, now)
	u.BirthDate = seed.BirthDate
	u.Age = seed.Age
	u.Gender = seed.Gender
	u.Height = seed.Height
	u.Weight = seed.Weight
	u.Foot = seed.Foot
	u.DominantFoot = seed.Foot
	u.Position = seed.Position
	u.SecondPosition = seed.SecondPosition
	u.Club = "上海绿地青训俱乐部"
	u.CurrentTeam = []string{"U12 精英队", "U14 梯队"}[seed.TeamIndex]
	u.School = seed.School
	u.StartYear = 2020 + seed.TeamIndex
	u.FARegistered = true
	u.Association = "上海市足球协会"
	u.JerseyColor = "绿色"
	u.JerseyNumber = seed.JerseyNumber
	u.PlayingStyle = mustJSON([]string{"速度型", "团队型"})
	u.Wechat = "sq_player_" + seed.Phone[len(seed.Phone)-4:]
	u.TechnicalTags = mustJSON(seed.Tags)
	u.MentalTags = mustJSON([]string{"自律", "抗压", "团队意识"})
	u.Experiences = mustJSON([]map[string]string{{"period": "2023-2024", "team": u.CurrentTeam, "position": seed.Position, "achievement": "稳定参加市级青少年比赛"}})
	u.VideoUrl = "https://example.com/videos/player-demo.mp4"
	u.Sprint30m = 4.5
	u.StandingLongJump = 195
	u.Flexibility = 10
	u.PushUp = 24
	u.SitUps = 38
	u.FiveMeterShuttle = 29.8
	u.SitAndReach = 10
	u.FatherHeight = 176
	u.FatherPhone = "139" + seed.Phone[len(seed.Phone)-8:]
	u.FatherOccupation = "工程师"
	u.FatherEdu = "本科"
	u.MotherHeight = 164
	u.MotherPhone = "137" + seed.Phone[len(seed.Phone)-8:]
	u.MotherOccupation = "教师"
	u.MotherEdu = "本科"
	return u
}

func demoPlayers() []playerSeed {
	return []playerSeed{
		{"13800002001", "林子墨", "子墨", "2014-03-12", 12, "男", 151, 42, "right", "边锋", "前锋", "上海", "上海", "浦东新区实验小学", 7, 0, []string{"速度", "突破", "右路推进"}},
		{"13800002002", "周宇航", "宇航", "2014-08-21", 12, "男", 154, 44, "left", "中场", "边锋", "上海", "上海", "建平实验小学", 10, 0, []string{"传球", "视野", "组织"}},
		{"13800002003", "陈奕辰", "奕辰", "2015-01-18", 11, "男", 148, 39, "right", "前锋", "边锋", "上海", "上海", "福山外国语小学", 9, 0, []string{"射门", "抢点", "冲刺"}},
		{"13800002004", "王浩然", "浩然", "2014-11-06", 12, "男", 156, 47, "right", "后卫", "中场", "上海", "上海", "进才实验小学", 5, 0, []string{"对抗", "防守", "长传"}},
		{"13800002005", "赵一诺", "一诺", "2015-05-09", 11, "女", 146, 38, "left", "中场", "后卫", "上海", "上海", "浦明师范附小", 8, 0, []string{"控球", "节奏", "团队"}},
		{"13800002006", "刘景行", "景行", "2014-06-29", 12, "男", 158, 49, "right", "门将", "后卫", "上海", "上海", "明珠小学", 1, 0, []string{"扑救", "反应", "出击"}},
		{"13800002007", "孙嘉树", "嘉树", "2012-04-15", 14, "男", 166, 54, "right", "中后卫", "后腰", "上海", "上海", "进才中学北校", 4, 1, []string{"防空", "卡位", "领导力"}},
		{"13800002008", "吴承泽", "承泽", "2012-09-02", 14, "男", 163, 52, "left", "边后卫", "边锋", "上海", "上海", "建平西校", 3, 1, []string{"往返", "传中", "速度"}},
		{"13800002009", "黄睿哲", "睿哲", "2013-02-25", 13, "男", 160, 50, "right", "后腰", "中场", "上海", "上海", "洋泾中学", 6, 1, []string{"拦截", "转移", "覆盖"}},
		{"13800002010", "郑思远", "思远", "2012-12-11", 14, "男", 168, 56, "right", "前锋", "边锋", "上海", "上海", "张江集团学校", 11, 1, []string{"射门", "跑位", "对抗"}},
		{"13800002011", "马若琳", "若琳", "2013-07-19", 13, "女", 158, 46, "both", "前腰", "边锋", "上海", "上海", "上海实验学校", 18, 1, []string{"创造力", "直塞", "盘带"}},
		{"13800002012", "何景天", "景天", "2012-10-30", 14, "男", 170, 58, "right", "门将", "中后卫", "上海", "上海", "川沙中学", 12, 1, []string{"扑救", "指挥", "开球"}},
	}
}

func avatar(name string) string {
	return "https://ui-avatars.com/api/?name=" + name + "&background=39FF14&color=000&size=200&bold=true"
}

func imageURL(name string) string {
	return "https://images.unsplash.com/photo-1431324155629-1a6deb1dec8d?auto=format&fit=crop&w=1200&q=80&demo=" + name
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func ids(users []models.User) []uint {
	result := make([]uint, 0, len(users))
	for _, user := range users {
		result = append(result, user.ID)
	}
	return result
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func startOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	date := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return date.AddDate(0, 0, 1-weekday)
}
