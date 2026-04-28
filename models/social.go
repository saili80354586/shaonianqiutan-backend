package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Like 点赞模型
type Like struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"user_id" gorm:"index;not null"`
	TargetType string         `json:"target_type" gorm:"size:50;not null"`
	TargetID   uint           `json:"target_id" gorm:"not null"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	User       *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 表名
func (Like) TableName() string {
	return "likes"
}

// Favorite 收藏模型
type Favorite struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"user_id" gorm:"index;not null"`
	TargetType string         `json:"target_type" gorm:"size:50;not null"`
	TargetID   uint           `json:"target_id" gorm:"not null"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	User       *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 表名
func (Favorite) TableName() string {
	return "favorites"
}

// Comment 评论模型
type Comment struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"user_id" gorm:"index;not null"`
	TargetType string         `json:"target_type" gorm:"size:50;not null"`
	TargetID   uint           `json:"target_id" gorm:"not null"`
	ParentID   *uint          `json:"parent_id" gorm:"index"`
	Content    string         `json:"content" gorm:"type:text;not null"`
	LikesCount int            `json:"likes_count" gorm:"default:0"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	User       *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Parent     *Comment       `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Replies    []Comment      `json:"replies,omitempty" gorm:"foreignKey:ParentID"`
}

// TableName 表名
func (Comment) TableName() string {
	return "comments"
}

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeLike     NotificationType = "like"
	NotificationTypeFavorite NotificationType = "favorite"
	NotificationTypeComment  NotificationType = "comment"
	NotificationTypeMention  NotificationType = "mention"
	NotificationTypeSystem   NotificationType = "system"
	NotificationTypeOrder    NotificationType = "order"
	NotificationTypeReport   NotificationType = "report"
	NotificationTypeTask     NotificationType = "task"
	NotificationTypeInquiry  NotificationType = "inquiry" // 咨询意向
	NotificationTypeFollow   NotificationType = "follow"  // 被关注
	NotificationTypeMessage  NotificationType = "message" // 新私信
	// 周报相关
	NotificationTypeWeeklyReportCreated  NotificationType = "weekly_report_created"  // 发起周报
	NotificationTypeWeeklyReportRejected NotificationType = "weekly_report_rejected" // 周报被退回
	NotificationTypeWeeklyReportApproved NotificationType = "weekly_report_approved" // 周报审核完成
	NotificationTypeWeeklyReportReminder NotificationType = "weekly_report_reminder" // 周报截止提醒
	// 比赛总结相关
	NotificationTypeMatchSummaryCreated  NotificationType = "match_summary_created"  // 创建比赛
	NotificationTypeMatchPlayerReminder  NotificationType = "match_player_reminder"  // 待自评提醒
	NotificationTypeMatchCoachReminder   NotificationType = "match_coach_reminder"   // 待点评提醒
	NotificationTypeMatchSummaryComplete NotificationType = "match_summary_complete" // 点评完成
	// 俱乐部活动相关
	NotificationTypeActivityRegistration NotificationType = "activity_registration" // 新报名
	NotificationTypeActivityApproved     NotificationType = "activity_approved"     // 报名通过
	NotificationTypeActivityRejected     NotificationType = "activity_rejected"     // 报名拒绝
	// 邀请相关
	NotificationTypeInvitation  NotificationType = "invitation"   // 收到邀请
	NotificationTypeTrialInvite NotificationType = "trial_invite" // 收到试训邀请
	NotificationTypeScoutReport NotificationType = "scout_report" // 收到球探报告
)

// Notification 通知模型
type Notification struct {
	ID        uint             `json:"id" gorm:"primaryKey"`
	UserID    uint             `json:"user_id" gorm:"index;not null"`
	Type      NotificationType `json:"type" gorm:"size:50;not null"`
	Title     string           `json:"title" gorm:"size:200;not null"`
	Content   string           `json:"content" gorm:"type:text"`
	Data      string           `json:"data" gorm:"type:json"`
	IsRead    bool             `json:"is_read" gorm:"default:false"`
	Priority  int              `json:"priority" gorm:"default:3"`
	CreatedAt time.Time        `json:"created_at"`
	User      *User            `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 表名
func (Notification) TableName() string {
	return "notifications"
}

// SetData 设置通知扩展数据
func (n *Notification) SetData(data *NotificationData) {
	if data == nil {
		return
	}
	jsonBytes, _ := json.Marshal(data)
	n.Data = string(jsonBytes)
}

// GetData 获取通知扩展数据
func (n *Notification) GetData() *NotificationData {
	if n.Data == "" {
		return nil
	}
	var data NotificationData
	json.Unmarshal([]byte(n.Data), &data)
	return &data
}

// NotificationData 通知扩展数据
type NotificationData struct {
	TriggerUserID   uint                   `json:"trigger_user_id,omitempty"`
	TriggerUserName string                 `json:"trigger_user_name,omitempty"`
	TriggerAvatar   string                 `json:"trigger_avatar,omitempty"`
	TargetType      string                 `json:"target_type,omitempty"`
	TargetID        uint                   `json:"target_id,omitempty"`
	TargetIDs       []uint                 `json:"target_ids,omitempty"`
	CommentID       uint                   `json:"comment_id,omitempty"`
	CommentContent  string                 `json:"comment_content,omitempty"`
	ReportID        uint                   `json:"report_id,omitempty"`
	ReportTitle     string                 `json:"report_title,omitempty"`
	Link            string                 `json:"link,omitempty"`
	Extra           map[string]interface{} `json:"extra,omitempty"`
}

// SocialAchievementID 社交成就ID
type SocialAchievementID string

const (
	AchievementFirstReport   SocialAchievementID = "first_report"
	AchievementReport10      SocialAchievementID = "report_10"
	AchievementReport100     SocialAchievementID = "report_100"
	AchievementFirstFollower SocialAchievementID = "first_follower"
	AchievementFollowers100  SocialAchievementID = "followers_100"
	AchievementFirstLike     SocialAchievementID = "first_like"
	AchievementLikes100      SocialAchievementID = "likes_100"
	AchievementFirstComment  SocialAchievementID = "first_comment"
	AchievementFirstScout    SocialAchievementID = "first_scout"
	AchievementPlayerJoined  SocialAchievementID = "player_joined"
	AchievementStreak7       SocialAchievementID = "streak_7"
	AchievementStreak30      SocialAchievementID = "streak_30"
)

// SocialAchievement 社交成就定义
type SocialAchievement struct {
	ID          SocialAchievementID `json:"id" gorm:"primaryKey;size:50"`
	Name        string              `json:"name" gorm:"size:100;not null"`
	Description string              `json:"description" gorm:"size:500"`
	Icon        string              `json:"icon" gorm:"size:100"`
	Category    string              `json:"category" gorm:"size:50"`
	Condition   string              `json:"condition" gorm:"size:200"`
	Threshold   int                 `json:"threshold"`
}

// TableName 表名
func (SocialAchievement) TableName() string {
	return "social_achievements"
}

// UserAchievement 用户成就记录
type UserSocialAchievement struct {
	ID            uint                `json:"id" gorm:"primaryKey"`
	UserID        uint                `json:"user_id" gorm:"index;not null"`
	AchievementID SocialAchievementID `json:"achievement_id" gorm:"size:50;not null"`
	AchievedAt    *time.Time          `json:"achieved_at"`
	User          *User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Achievement   *SocialAchievement  `json:"achievement,omitempty" gorm:"foreignKey:AchievementID"`
}

// TableName 表名
func (UserSocialAchievement) TableName() string {
	return "user_social_achievements"
}

// UserStats 用户统计
type UserStats struct {
	UserID            uint       `json:"user_id" gorm:"primaryKey"`
	LikesReceived     int        `json:"likes_received" gorm:"default:0"`
	FavoritesReceived int        `json:"favorites_received" gorm:"default:0"`
	CommentsReceived  int        `json:"comments_received" gorm:"default:0"`
	FollowersCount    int        `json:"followers_count" gorm:"default:0"`
	FollowingCount    int        `json:"following_count" gorm:"default:0"`
	LoginStreak       int        `json:"login_streak" gorm:"default:0"`
	LastLoginDate     *time.Time `json:"last_login_date"`
	UpdatedAt         time.Time  `json:"updated_at"`
	User              *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 表名
func (UserStats) TableName() string {
	return "user_social_stats"
}

// GetTime 获取当前时间
func GetTime() time.Time {
	return time.Now()
}

// RemindResult 催办结果
type RemindResult struct {
	Sent   int `json:"sent"`
	Failed int `json:"failed"`
}
