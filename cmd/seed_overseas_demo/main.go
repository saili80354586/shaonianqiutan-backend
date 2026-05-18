package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

const demoPasswordHash = "$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK"

type overseasPlayerSeed struct {
	Phone           string
	Name            string
	Nickname        string
	Country         string
	City            string
	Club            string
	School          string
	Position        string
	SecondPosition  string
	Age             int
	BirthDate       string
	Height          float64
	Weight          float64
	Foot            string
	JerseyNumber    int
	TechnicalTags   []string
	MentalTags      []string
	ScoutRating     int
	PotentialRating string
	Summary         string
}

func main() {
	config.LoadEnv()
	config.InitDB()

	db := config.GetDB()
	if err := db.Transaction(func(tx *gorm.DB) error {
		scout, err := ensureOverseasDemoScout(tx)
		if err != nil {
			return err
		}
		for _, seed := range overseasPlayerSeeds() {
			if err := upsertOverseasPlayer(tx, scout, seed); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Fatalf("seed overseas demo players failed: %v", err)
	}

	log.Println("seed_overseas_demo completed successfully")
}

func ensureOverseasDemoScout(tx *gorm.DB) (*models.Scout, error) {
	var scout models.Scout
	if err := tx.Where("current_organization = ?", "少年球探海外观察站").First(&scout).Error; err == nil {
		return &scout, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	now := time.Now()
	user := models.User{
		Phone:                "13800000029",
		Password:             demoPasswordHash,
		Nickname:             "海外观察员",
		Avatar:               avatar("海外观察员"),
		Role:                 models.RoleScout,
		CurrentRole:          models.RoleScout,
		Status:               models.StatusActive,
		Name:                 "海外观察员",
		Gender:               "男",
		Country:              "中国",
		Province:             "上海",
		City:                 "上海",
		NotificationSettings: `{"system":true,"order":true,"weekly":true,"social":true,"message":true,"email":false}`,
		PrivacySettings:      `{"profileVisible":true,"phoneVisible":false,"searchable":true}`,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := tx.Where("phone = ?", user.Phone).Assign(user).FirstOrCreate(&user).Error; err != nil {
		return nil, err
	}

	scout = models.Scout{
		UserID:              user.ID,
		ScoutingExperience:  "5-10",
		Specialties:         mustJSON([]string{"海外青训", "高潜球员", "跨区域跟踪"}),
		PreferredAgeGroups:  mustJSON([]string{"U12", "U14", "U16"}),
		ScoutingRegions:     mustJSON([]string{"欧洲", "东亚"}),
		CurrentOrganization: "少年球探海外观察站",
		Bio:                 "负责跟踪海外华裔与国际青少年球员，为平台海外球员专区提供演示观察样本。",
		Verified:            true,
		TotalDiscovered:     16,
		TotalReports:        12,
		TotalAdopted:        4,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err := tx.Where("user_id = ?", user.ID).Assign(scout).FirstOrCreate(&scout).Error; err != nil {
		return nil, err
	}
	return &scout, nil
}

func upsertOverseasPlayer(tx *gorm.DB, scout *models.Scout, seed overseasPlayerSeed) error {
	now := time.Now()
	user := models.User{
		Phone:                seed.Phone,
		Password:             demoPasswordHash,
		Nickname:             seed.Nickname,
		Avatar:               avatar(seed.Name),
		Role:                 models.RoleUser,
		CurrentRole:          models.RoleUser,
		Status:               models.StatusActive,
		Name:                 seed.Name,
		BirthDate:            seed.BirthDate,
		Age:                  seed.Age,
		Gender:               "男",
		Height:               seed.Height,
		Weight:               seed.Weight,
		Foot:                 seed.Foot,
		DominantFoot:         seed.Foot,
		Position:             seed.Position,
		SecondPosition:       seed.SecondPosition,
		Country:              seed.Country,
		Province:             seed.Country,
		City:                 seed.City,
		Club:                 seed.Club,
		CurrentTeam:          seed.Club,
		School:               seed.School,
		StartYear:            2021,
		FARegistered:         true,
		Association:          seed.Country + " 青训体系",
		JerseyColor:          "蓝白",
		JerseyNumber:         seed.JerseyNumber,
		PlayingStyle:         mustJSON([]string{"速度型", "技术型"}),
		Wechat:               "overseas_" + seed.Phone[len(seed.Phone)-4:],
		TechnicalTags:        mustJSON(seed.TechnicalTags),
		MentalTags:           mustJSON(seed.MentalTags),
		Experiences:          mustJSON([]map[string]string{{"period": "2024-2026", "team": seed.Club, "position": seed.Position, "achievement": "海外青训演示样本"}}),
		VideoUrl:             "https://example.com/videos/overseas-demo.mp4",
		Sprint30m:            4.25,
		StandingLongJump:     226,
		Flexibility:          16,
		PushUp:               34,
		SitUps:               55,
		FiveMeterShuttle:     12.1,
		SitAndReach:          16,
		FatherHeight:         178,
		FatherPhone:          "139" + seed.Phone[len(seed.Phone)-8:],
		FatherOccupation:     "海外工作者",
		FatherEdu:            "本科",
		MotherHeight:         165,
		MotherPhone:          "137" + seed.Phone[len(seed.Phone)-8:],
		MotherOccupation:     "教师",
		MotherEdu:            "本科",
		NotificationSettings: `{"system":true,"order":true,"weekly":true,"social":true,"message":true,"email":false}`,
		PrivacySettings:      `{"profileVisible":true,"phoneVisible":false,"searchable":true}`,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := tx.Where("phone = ?", seed.Phone).Assign(user).FirstOrCreate(&user).Error; err != nil {
		return err
	}

	player := models.Player{
		UserID:     user.ID,
		Name:       user.Name,
		Nickname:   user.Nickname,
		Province:   seed.Country,
		City:       seed.City,
		Position:   seed.Position,
		Age:        seed.Age,
		BirthDate:  seed.BirthDate,
		Height:     int(seed.Height),
		Weight:     int(seed.Weight),
		Foot:       seed.Foot,
		Club:       seed.Club,
		School:     seed.School,
		Phone:      seed.Phone,
		Avatar:     user.Avatar,
		VideoURL:   user.VideoUrl,
		Status:     1,
		CreateTime: now,
		UpdateTime: now,
	}
	if err := tx.Where("user_id = ?", user.ID).Assign(player).FirstOrCreate(&player).Error; err != nil {
		return err
	}

	if seed.ScoutRating <= 0 || scout == nil {
		return nil
	}
	report := models.ScoutReport{
		ScoutID:         scout.ID,
		PlayerID:        user.ID,
		OverallRating:   seed.ScoutRating,
		PotentialRating: seed.PotentialRating,
		Status:          "published",
		Strengths:       mustJSON(seed.TechnicalTags),
		Weaknesses:      mustJSON([]string{"需继续跟踪正式比赛样本", "跨地区参照数据仍需补充"}),
		TechnicalSkills: mustJSON(map[string]int{"shooting": seed.ScoutRating - 3, "passing": seed.ScoutRating - 2, "dribbling": seed.ScoutRating, "defending": seed.ScoutRating - 8, "physical": seed.ScoutRating - 1, "mentality": seed.ScoutRating - 4}),
		Summary:         seed.Summary,
		Recommendation:  "建议纳入海外球员专区重点观察池，持续补充比赛视频、体测和球探报告。",
		TargetClub:      "少年球探海外机会池",
		PublishedAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	return tx.Where("scout_id = ? AND player_id = ?", scout.ID, user.ID).Assign(report).FirstOrCreate(&report).Error
}

func overseasPlayerSeeds() []overseasPlayerSeed {
	return []overseasPlayerSeed{
		{"13800002101", "卢卡斯·陈", "卢卡斯", "西班牙", "巴塞罗那", "巴萨 Escola U13", "Barcelona Youth Academy", "边锋", "前锋", 13, "2013-02-16", 160, 49, "right", 7, []string{"速度", "一对一", "右路突破"}, []string{"自信", "创造力", "抗压"}, 94, "S", "边路启动和连续变向能力突出，适合继续跟踪其高强度比赛决策。"},
		{"13800002102", "米卡·林", "米卡", "日本", "横滨", "横滨水手青训 U12", "Yokohama Junior School", "中场", "前腰", 12, "2014-05-28", 153, 43, "left", 10, []string{"控球", "节奏", "短传"}, []string{"冷静", "团队意识", "专注"}, 82, "B", "节奏控制和小范围接应成熟，适合作为东亚青训样本持续观察。"},
		{"13800002103", "诺亚·王", "诺亚", "德国", "慕尼黑", "慕尼黑青训中心 U14", "Munich International School", "中后卫", "后腰", 14, "2012-10-04", 172, 60, "right", 4, []string{"防空", "对抗", "出球"}, []string{"领导力", "纪律性", "稳定"}, 78, "B", "身体对抗和后场出球具备基础优势，建议补充速度和灵敏数据。"},
		{"13800002104", "伊桑·李", "伊桑", "法国", "巴黎", "巴黎青训学院 U11", "Paris Youth School", "前锋", "边锋", 11, "2015-08-19", 148, 39, "both", 9, []string{"射门", "跑位", "双脚"}, []string{"积极", "求胜欲", "学习能力"}, 0, "", "射门感觉和门前跑位有亮点，当前保留为待评估样本。"},
	}
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func avatar(name string) string {
	return "https://ui-avatars.com/api/?name=" + name + "&background=00D4FF&color=000&size=200&bold=true"
}
