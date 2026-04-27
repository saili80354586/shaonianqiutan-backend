package main

import (
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// SeedSocialAchievements 初始化社交成就数据
func SeedSocialAchievements(db *gorm.DB) error {
	achievements := []models.SocialAchievement{
		// 贡献类成就
		{ID: "first_report", Name: "初出茅庐", Description: "发布第一份报告", Icon: "trophy", Category: "contribution", Condition: "发布1份报告", Threshold: 1},
		{ID: "report_10", Name: "小有名气", Description: "发布10份报告", Icon: "trophy", Category: "contribution", Condition: "发布10份报告", Threshold: 10},
		{ID: "report_100", Name: "报告大师", Description: "发布100份报告", Icon: "trophy", Category: "contribution", Condition: "发布100份报告", Threshold: 100},

		// 活跃类成就
		{ID: "first_like", Name: "赞不绝口", Description: "收到第一个点赞", Icon: "heart", Category: "engagement", Condition: "收到1个点赞", Threshold: 1},
		{ID: "likes_100", Name: "点赞达人", Description: "收到100个赞", Icon: "heart", Category: "engagement", Condition: "收到100个赞", Threshold: 100},
		{ID: "first_comment", Name: "议论纷纷", Description: "收到第一条评论", Icon: "message", Category: "engagement", Condition: "收到1条评论", Threshold: 1},
		{ID: "streak_7", Name: "连续活跃", Description: "连续登录7天", Icon: "zap", Category: "engagement", Condition: "连续登录7天", Threshold: 7},
		{ID: "streak_30", Name: "习惯养成", Description: "连续登录30天", Icon: "zap", Category: "engagement", Condition: "连续登录30天", Threshold: 30},

		// 社交类成就
		{ID: "first_follower", Name: "收获粉丝", Description: "获得第一个粉丝", Icon: "users", Category: "social", Condition: "粉丝数=1", Threshold: 1},
		{ID: "followers_100", Name: "小有名气", Description: "获得100个粉丝", Icon: "users", Category: "social", Condition: "粉丝数=100", Threshold: 100},

		// 里程碑成就
		{ID: "first_scout", Name: "慧眼识珠", Description: "球员被球探发掘", Icon: "star", Category: "milestone", Condition: "球探报告提及", Threshold: 1},
		{ID: "player_joined", Name: "星探伯乐", Description: "发掘的球员加入俱乐部", Icon: "award", Category: "milestone", Condition: "球员被俱乐部选中", Threshold: 1},
	}

	for _, achievement := range achievements {
		// 使用 Where 条件避免重复插入
		result := db.Where("id = ?", achievement.ID).FirstOrCreate(&achievement)
		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}
