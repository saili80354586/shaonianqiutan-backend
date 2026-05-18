package repositories

import (
	"strings"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// SocialRepository 社交互动Repository
type SocialRepository struct {
	db *gorm.DB
}

// NewSocialRepository 创建社交Repository
func NewSocialRepository(db *gorm.DB) *SocialRepository {
	return &SocialRepository{db: db}
}

// ============ 点赞相关 ============

// Like 点赞
func (r *SocialRepository) Like(userID uint, targetType string, targetID uint) (*models.Like, error) {
	like := &models.Like{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
	}
	err := r.db.Create(like).Error
	return like, err
}

// Unlike 取消点赞
func (r *SocialRepository) Unlike(userID uint, targetType string, targetID uint) error {
	return r.db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Delete(&models.Like{}).Error
}

// IsLiked 检查是否已点赞
func (r *SocialRepository) IsLiked(userID uint, targetType string, targetID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Like{}).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Count(&count).Error
	return count > 0, err
}

// GetLikesByTarget 获取目标的点赞列表
func (r *SocialRepository) GetLikesByTarget(targetType string, targetID uint, limit, offset int) ([]models.Like, int64, error) {
	var likes []models.Like
	var total int64

	query := r.db.Model(&models.Like{}).Where("target_type = ? AND target_id = ?", targetType, targetID)
	query.Count(&total)

	err := query.Preload("User").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&likes).Error
	return likes, total, err
}

// GetUserLikes 获取用户的点赞列表
func (r *SocialRepository) GetUserLikes(userID uint, limit, offset int) ([]models.Like, int64, error) {
	var likes []models.Like
	var total int64

	query := r.db.Model(&models.Like{}).Where("user_id = ?", userID)
	query.Count(&total)

	err := query.
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&likes).Error
	return likes, total, err
}

// GetTargetLikeCount 获取目标的点赞数
func (r *SocialRepository) GetTargetLikeCount(targetType string, targetID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Like{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		Count(&count).Error
	return count, err
}

// ============ 收藏相关 ============

// Favorite 收藏
func (r *SocialRepository) Favorite(userID uint, targetType string, targetID uint) (*models.Favorite, error) {
	fav := &models.Favorite{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
	}
	err := r.db.Create(fav).Error
	return fav, err
}

// Unfavorite 取消收藏
func (r *SocialRepository) Unfavorite(userID uint, targetType string, targetID uint) error {
	return r.db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Delete(&models.Favorite{}).Error
}

// IsFavorited 检查是否已收藏
func (r *SocialRepository) IsFavorited(userID uint, targetType string, targetID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Favorite{}).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Count(&count).Error
	return count > 0, err
}

// GetUserFavorites 获取用户的收藏列表
func (r *SocialRepository) GetUserFavorites(userID uint, targetType string, limit, offset int) ([]models.Favorite, int64, error) {
	var favorites []models.Favorite
	var total int64

	query := r.db.Model(&models.Favorite{}).Where("user_id = ?", userID)
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	query.Count(&total)

	err := query.
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&favorites).Error
	return favorites, total, err
}

// ============ 评论相关 ============

// CreateComment 创建评论
func (r *SocialRepository) CreateComment(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

// UpdateComment 更新评论
func (r *SocialRepository) UpdateComment(commentID uint, content string) error {
	return r.db.Model(&models.Comment{}).Where("id = ?", commentID).Updates(map[string]interface{}{
		"content": content,
	}).Error
}

// DeleteComment 删除评论
func (r *SocialRepository) DeleteComment(commentID uint) error {
	return r.db.Where("id = ?", commentID).Delete(&models.Comment{}).Error
}

// GetComment 获取评论详情
func (r *SocialRepository) GetComment(commentID uint) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.Preload("User").First(&comment, commentID).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// GetCommentsByTarget 获取目标的评论列表
func (r *SocialRepository) GetCommentsByTarget(targetType string, targetID uint, limit, offset int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	// 获取根评论
	err := r.db.Model(&models.Comment{}).
		Where("target_type = ? AND target_id = ? AND parent_id IS NULL", targetType, targetID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("target_type = ? AND target_id = ? AND parent_id IS NULL", targetType, targetID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	// 手动加载用户和回复
	for i := range comments {
		var user models.User
		if err := r.db.First(&user, comments[i].UserID).Error; err == nil {
			comments[i].User = &user
		}
		var replies []models.Comment
		if err := r.db.Where("parent_id = ?", comments[i].ID).Order("created_at ASC").Find(&replies).Error; err == nil {
			for j := range replies {
				var replyUser models.User
				if err := r.db.First(&replyUser, replies[j].UserID).Error; err == nil {
					replies[j].User = &replyUser
				}
			}
			comments[i].Replies = replies
		}
	}

	return comments, total, nil
}

// GetCommentReplies 获取评论的回复
func (r *SocialRepository) GetCommentReplies(parentID uint, limit, offset int) ([]models.Comment, int64, error) {
	var replies []models.Comment
	var total int64

	query := r.db.Model(&models.Comment{}).Where("parent_id = ?", parentID)
	query.Count(&total)

	err := query.Preload("User").
		Order("created_at ASC").
		Limit(limit).Offset(offset).
		Find(&replies).Error
	return replies, total, err
}

// IncrementCommentLikes 增加评论点赞数
func (r *SocialRepository) IncrementCommentLikes(commentID uint) error {
	return r.db.Model(&models.Comment{}).Where("id = ?", commentID).
		UpdateColumn("likes_count", gorm.Expr("likes_count + 1")).Error
}

// DecrementCommentLikes 减少评论点赞数
func (r *SocialRepository) DecrementCommentLikes(commentID uint) error {
	return r.db.Model(&models.Comment{}).Where("id = ?", commentID).
		UpdateColumn("likes_count", gorm.Expr("likes_count - 1")).Error
}

// ============ 通知相关 ============

// CreateNotification 创建通知
func (r *SocialRepository) CreateNotification(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

// GetNotifications 获取用户通知列表
func (r *SocialRepository) GetNotifications(userID uint, notificationType string, limit, offset int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := r.db.Model(&models.Notification{}).Where("user_id = ?", userID)
	if notificationType != "" {
		query = query.Where("type = ?", notificationType)
	}
	query.Count(&total)

	err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notifications).Error
	return notifications, total, err
}

// GetUnreadNotificationCount 获取未读通知数
func (r *SocialRepository) GetUnreadNotificationCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// MarkNotificationAsRead 标记通知已读
func (r *SocialRepository) MarkNotificationAsRead(notificationID uint) error {
	return r.db.Model(&models.Notification{}).Where("id = ?", notificationID).
		Update("is_read", true).Error
}

// MarkAllNotificationsAsRead 标记所有通知已读
func (r *SocialRepository) MarkAllNotificationsAsRead(userID uint) error {
	return r.db.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

// DeleteNotification 删除通知
func (r *SocialRepository) DeleteNotification(notificationID uint) error {
	return r.db.Delete(&models.Notification{}, notificationID).Error
}

// ============ 成就相关 ============

// GetAllAchievements 获取所有成就定义
func (r *SocialRepository) GetAllAchievements() ([]models.SocialAchievement, error) {
	var achievements []models.SocialAchievement
	err := r.db.Find(&achievements).Error
	return achievements, err
}

// GetUserSocialAchievements 获取用户已获得的成就
func (r *SocialRepository) GetUserSocialAchievements(userID uint) ([]models.UserSocialAchievement, error) {
	var userAchievements []models.UserSocialAchievement
	err := r.db.Preload("SocialAchievement").
		Where("user_id = ?", userID).
		Find(&userAchievements).Error
	return userAchievements, err
}

// HasSocialAchievement 检查用户是否已有成就
func (r *SocialRepository) HasSocialAchievement(userID uint, achievementID models.SocialAchievementID) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserSocialAchievement{}).
		Where("user_id = ? AND achievement_id = ?", userID, achievementID).
		Count(&count).Error
	return count > 0, err
}

// GrantSocialAchievement 授予成就
func (r *SocialRepository) GrantSocialAchievement(userID uint, achievementID models.SocialAchievementID) error {
	userAchievement := &models.UserSocialAchievement{
		UserID:        userID,
		AchievementID: achievementID,
	}
	return r.db.Create(userAchievement).Error
}

// ============ 关注相关 ============

// Follow 关注用户
func (r *SocialRepository) Follow(followerID, followingID uint) (*models.Follow, error) {
	follow := &models.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}
	err := r.db.Create(follow).Error
	return follow, err
}

// Unfollow 取消关注
func (r *SocialRepository) Unfollow(followerID, followingID uint) error {
	return r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&models.Follow{}).Error
}

// IsFollowing 检查是否已关注
func (r *SocialRepository) IsFollowing(followerID, followingID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Follow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count).Error
	return count > 0, err
}

// IsMutualFollow 检查双方是否互相关注
func (r *SocialRepository) IsMutualFollow(userID1, userID2 uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Follow{}).
		Where("(follower_id = ? AND following_id = ?) OR (follower_id = ? AND following_id = ?)",
			userID1, userID2, userID2, userID1).
		Count(&count).Error
	return count >= 2, err
}

// GetFollowers 获取用户的粉丝列表
func (r *SocialRepository) GetFollowers(userID uint, limit, offset int) ([]models.Follow, int64, error) {
	var follows []models.Follow
	var total int64

	query := r.db.Model(&models.Follow{}).Where("following_id = ?", userID)
	query.Count(&total)

	err := query.Preload("Follower").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&follows).Error
	return follows, total, err
}

// GetFollowing 获取用户的关注列表
func (r *SocialRepository) GetFollowing(userID uint, limit, offset int) ([]models.Follow, int64, error) {
	var follows []models.Follow
	var total int64

	query := r.db.Model(&models.Follow{}).Where("follower_id = ?", userID)
	query.Count(&total)

	err := query.Preload("Following").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&follows).Error
	return follows, total, err
}

// GetFollowerCount 获取粉丝数
func (r *SocialRepository) GetFollowerCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Follow{}).
		Where("following_id = ?", userID).
		Count(&count).Error
	return count, err
}

// GetFollowingCount 获取关注数
func (r *SocialRepository) GetFollowingCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Follow{}).
		Where("follower_id = ?", userID).
		Count(&count).Error
	return count, err
}

// ============ 用户统计相关 ============

// GetOrCreateUserStats 获取或创建用户统计
func (r *SocialRepository) GetOrCreateUserStats(userID uint) (*models.UserStats, error) {
	var stats models.UserStats
	result := r.db.Where("user_id = ?", userID).First(&stats)
	if result.Error == gorm.ErrRecordNotFound {
		stats = models.UserStats{UserID: userID}
		err := r.db.Create(&stats).Error
		return &stats, err
	}
	return &stats, result.Error
}

// IncrementField 增加统计字段
func (r *SocialRepository) IncrementField(userID uint, field string, delta int) error {
	return r.db.Model(&models.UserStats{}).
		Where("user_id = ?", userID).
		Update(field, gorm.Expr(field+" + ?", delta)).Error
}

// UpdateLoginStreak 更新登录连续天数
func (r *SocialRepository) UpdateLoginStreak(userID uint) error {
	stats, err := r.GetOrCreateUserStats(userID)
	if err != nil {
		return err
	}

	today := models.GetTime().Format("2006-01-02")

	if stats.LastLoginDate != nil {
		lastLogin := stats.LastLoginDate.Format("2006-01-02")
		if lastLogin == today {
			// 今天已登录，不更新连续天数
			return nil
		}

		yesterday := models.GetTime().AddDate(0, 0, -1).Format("2006-01-02")
		if lastLogin == yesterday {
			// 昨天登录过，增加连续天数
			return r.IncrementField(userID, "login_streak", 1)
		}
	}

	// 重置连续天数
	return r.db.Model(&models.UserStats{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"login_streak":    1,
		"last_login_date": today,
	}).Error
}

// ============ 辅助方法 ============

// GetUserByID 根据ID获取用户
func (r *SocialRepository) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetTargetOwner 获取被操作内容的主人ID
// 返回: ownerID, targetTitle, error
func (r *SocialRepository) GetTargetOwner(targetType string, targetID uint) (uint, string, error) {
	switch targetType {
	case "player_homepage", "growth_record":
		// 球员主页和成长记录的主人是球员本人（通过 player 表关联）
		var player models.Player
		err := r.db.First(&player, targetID).Error
		if err != nil {
			return 0, "", err
		}
		return player.UserID, player.Name, nil

	case "scout_report":
		// 球探报告的主人
		var scoutReport models.ScoutReport
		err := r.db.First(&scoutReport, targetID).Error
		if err != nil {
			return 0, "", err
		}
		// 通过 scout_report.scout_id 查找 scout 表获取 user_id
		var scout models.Scout
		if err := r.db.First(&scout, scoutReport.ScoutID).Error; err != nil {
			return 0, "", err
		}
		return scout.UserID, "球探报告", nil

	case "analyst_report":
		// 分析师报告的主人
		var report models.Report
		err := r.db.First(&report, targetID).Error
		if err != nil {
			return 0, "", err
		}
		return report.AnalystID, report.PlayerName, nil

	case "comment":
		// 评论的主人是被回复的评论作者
		var comment models.Comment
		err := r.db.First(&comment, targetID).Error
		if err != nil {
			return 0, "", err
		}
		return comment.UserID, comment.Content, nil

	case "post":
		// 动态帖子的主人是发布者
		var post models.Post
		err := r.db.First(&post, targetID).Error
		if err != nil {
			return 0, "", err
		}
		return post.UserID, post.Content, nil

	case "video":
		// 视频的主人
		// TODO: 根据实际视频表结构调整
		return 0, "视频", nil

	default:
		return 0, "", nil
	}
}

// ============ 动态帖子相关 ============

// CreatePost 创建帖子
func (r *SocialRepository) CreatePost(post *models.Post) error {
	return r.db.Create(post).Error
}

// GetPostByID 根据ID获取帖子
func (r *SocialRepository) GetPostByID(postID uint) (*models.Post, error) {
	var post models.Post
	err := r.db.Preload("User").First(&post, postID).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// DeletePost 删除帖子
func (r *SocialRepository) DeletePost(postID uint) error {
	return r.db.Where("id = ?", postID).Delete(&models.Post{}).Error
}

// GetFeedPosts 获取动态流列表
func (r *SocialRepository) GetFeedPosts(roleTag string, province string, city string, page, pageSize int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	query := r.db.Model(&models.Post{}).Joins("LEFT JOIN users ON users.id = posts.user_id")
	if roleTag != "" && roleTag != "all" {
		if roleTag == "scout_analyst" {
			query = query.Where("(posts.role_tag IN ? OR users.role IN ?)", []string{"scout", "analyst"}, []string{"scout", "analyst"})
		} else {
			query = query.Where("(posts.role_tag = ? OR users.role = ?)", roleTag, roleTag)
		}
	}
	if province = strings.TrimSpace(province); province != "" {
		query = query.Where("users.province LIKE ?", "%"+province+"%")
	}
	if city = strings.TrimSpace(city); city != "" {
		query = query.Where("users.city LIKE ?", "%"+city+"%")
	}
	query.Count(&total)

	err := query.Preload("User").
		Order("posts.is_top DESC, posts.created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&posts).Error
	return posts, total, err
}

// GetUserPosts 获取用户发布的帖子
func (r *SocialRepository) GetUserPosts(userID uint, page, pageSize int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	query := r.db.Model(&models.Post{}).Where("user_id = ?", userID)
	query.Count(&total)

	err := query.Preload("User").
		Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&posts).Error
	return posts, total, err
}

// IncrementPostLikes 增加帖子点赞数
func (r *SocialRepository) IncrementPostLikes(postID uint) error {
	return r.db.Model(&models.Post{}).Where("id = ?", postID).
		UpdateColumn("likes_count", gorm.Expr("likes_count + 1")).Error
}

// DecrementPostLikes 减少帖子点赞数
func (r *SocialRepository) DecrementPostLikes(postID uint) error {
	return r.db.Model(&models.Post{}).Where("id = ?", postID).
		UpdateColumn("likes_count", gorm.Expr("likes_count - 1")).Error
}

// IncrementPostComments 增加帖子评论数
func (r *SocialRepository) IncrementPostComments(postID uint) error {
	return r.db.Model(&models.Post{}).Where("id = ?", postID).
		UpdateColumn("comments_count", gorm.Expr("comments_count + 1")).Error
}
