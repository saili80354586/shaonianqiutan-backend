package services

import (
	"encoding/json"

	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
)

// SocialService 社交服务
type SocialService struct {
	repo                *repositories.SocialRepository
	notificationService *NotificationService
}

// NewSocialService 创建社交服务
func NewSocialService(repo *repositories.SocialRepository, notificationService *NotificationService) *SocialService {
	return &SocialService{repo: repo, notificationService: notificationService}
}

// ============ 点赞服务 ============

// ToggleLike 切换点赞状态
func (s *SocialService) ToggleLike(userID uint, targetType string, targetID uint) (bool, error) {
	liked, err := s.repo.IsLiked(userID, targetType, targetID)
	if err != nil {
		return false, err
	}

	if liked {
		err = s.repo.Unlike(userID, targetType, targetID)
		return false, err
	}

	_, err = s.repo.Like(userID, targetType, targetID)
	if err != nil {
		return false, err
	}

	// 发送通知
	go s.sendLikeNotification(userID, targetType, targetID)
	// 检查成就
	go s.checkLikeAchievements(userID)

	return true, nil
}

func (s *SocialService) sendLikeNotification(userID uint, targetType string, targetID uint) {
	ownerID, _, _ := s.repo.GetTargetOwner(targetType, targetID)
	if ownerID == 0 || ownerID == userID {
		return
	}

	triggerUser, _ := s.repo.GetUserByID(userID)
	triggerName := "用户"
	triggerAvatar := ""
	if triggerUser != nil {
		if triggerUser.Nickname != "" {
			triggerName = triggerUser.Nickname
		}
		triggerAvatar = triggerUser.Avatar
	}

	data := &models.NotificationData{
		TriggerUserID:   userID,
		TriggerUserName: triggerName,
		TriggerAvatar:   triggerAvatar,
		TargetType:      targetType,
		TargetID:        targetID,
	}

	if s.notificationService != nil {
		s.notificationService.CreateNotification(ownerID, models.NotificationTypeLike, "收到点赞", triggerName+" 点赞了你的"+s.getTargetTypeName(targetType), data)
		return
	}

	notification := &models.Notification{
		UserID:   ownerID,
		Type:     models.NotificationType("like"),
		Title:    "收到点赞",
		Content:  triggerName + " 点赞了你的" + s.getTargetTypeName(targetType),
		Priority: 3,
	}

	if triggerAvatar != "" {
		data, _ := json.Marshal(map[string]string{"avatar": triggerAvatar})
		notification.Data = string(data)
	}

	s.repo.CreateNotification(notification)
}

func (s *SocialService) getTargetTypeName(targetType string) string {
	switch targetType {
	case "player_homepage":
		return "主页"
	case "scout_report":
		return "球探报告"
	case "analyst_report":
		return "分析报告"
	case "comment":
		return "评论"
	case "growth_record":
		return "成长记录"
	case "post":
		return "动态"
	case "video":
		return "视频"
	default:
		return "内容"
	}
}

// IsLiked 检查点赞状态
func (s *SocialService) IsLiked(userID uint, targetType string, targetID uint) bool {
	liked, _ := s.repo.IsLiked(userID, targetType, targetID)
	return liked
}

// GetTargetLikes 获取目标点赞列表
func (s *SocialService) GetTargetLikes(targetType string, targetID uint, page, pageSize int) ([]models.Like, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetLikesByTarget(targetType, targetID, pageSize, offset)
}

// GetUserLikes 获取用户点赞列表
func (s *SocialService) GetUserLikes(userID uint, page, pageSize int) ([]models.Like, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetUserLikes(userID, pageSize, offset)
}

// ============ 收藏服务 ============

// ToggleFavorite 切换收藏状态
func (s *SocialService) ToggleFavorite(userID uint, targetType string, targetID uint) (bool, error) {
	favorited, err := s.repo.IsFavorited(userID, targetType, targetID)
	if err != nil {
		return false, err
	}

	if favorited {
		err = s.repo.Unfavorite(userID, targetType, targetID)
		return false, err
	}

	_, err = s.repo.Favorite(userID, targetType, targetID)
	return true, err
}

// IsFavorited 检查收藏状态
func (s *SocialService) IsFavorited(userID uint, targetType string, targetID uint) bool {
	favorited, _ := s.repo.IsFavorited(userID, targetType, targetID)
	return favorited
}

// GetUserFavorites 获取用户收藏列表
func (s *SocialService) GetUserFavorites(userID uint, targetType string, page, pageSize int) ([]models.Favorite, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetUserFavorites(userID, targetType, pageSize, offset)
}

// ============ 评论服务 ============

// CommentItem 评论项
type CommentItem struct {
	models.Comment
	User    *models.User  `json:"user"`
	Replies []CommentItem `json:"replies"`
	IsLiked bool          `json:"is_liked"`
}

// GetComments 获取评论列表
func (s *SocialService) GetComments(targetType string, targetID uint, page, pageSize int) ([]CommentItem, int64, error) {
	comments, total, err := s.repo.GetCommentsByTarget(targetType, targetID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]CommentItem, len(comments))
	for i := range comments {
		items[i] = CommentItem{Comment: comments[i], User: comments[i].User, IsLiked: false}
	}
	return items, total, nil
}

// CreateComment 创建评论
func (s *SocialService) CreateComment(userID uint, targetType string, targetID uint, parentID *uint, content string) (*CommentItem, error) {
	comment := &models.Comment{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
		ParentID:   parentID,
		Content:    content,
	}

	err := s.repo.CreateComment(comment)
	if err != nil {
		return nil, err
	}

	// 加载用户信息
	user, _ := s.repo.GetUserByID(userID)

	// 异步发送通知
	go s.sendCommentNotification(userID, targetType, targetID, parentID, content)

	return &CommentItem{Comment: *comment, User: user, IsLiked: false}, nil
}

func (s *SocialService) sendCommentNotification(userID uint, targetType string, targetID uint, parentID *uint, content string) {
	triggerUser, _ := s.repo.GetUserByID(userID)
	triggerName := "用户"
	if triggerUser != nil {
		if triggerUser.Nickname != "" {
			triggerName = triggerUser.Nickname
		}
	}

	// 回复评论
	if parentID != nil {
		parentComment, _ := s.repo.GetComment(*parentID)
		if parentComment != nil && parentComment.UserID != userID {
			data := &models.NotificationData{
				TriggerUserID:   userID,
				TriggerUserName: triggerName,
				TargetType:      "comment",
				TargetID:        *parentID,
				CommentContent:  content,
			}
			if s.notificationService != nil {
				s.notificationService.CreateNotification(parentComment.UserID, models.NotificationTypeComment, "收到回复", triggerName+" 回复了你的评论", data)
			} else {
				notification := &models.Notification{
					UserID:   parentComment.UserID,
					Type:     models.NotificationType("comment"),
					Title:    "收到回复",
					Content:  triggerName + " 回复了你的评论",
					Priority: 3,
				}
				s.repo.CreateNotification(notification)
			}
			return
		}
	}

	// 新评论通知内容作者
	ownerID, _, _ := s.repo.GetTargetOwner(targetType, targetID)
	if ownerID == 0 || ownerID == userID {
		return
	}

	data := &models.NotificationData{
		TriggerUserID:   userID,
		TriggerUserName: triggerName,
		TargetType:      targetType,
		TargetID:        targetID,
		CommentContent:  content,
	}
	if s.notificationService != nil {
		s.notificationService.CreateNotification(ownerID, models.NotificationTypeComment, "收到评论", triggerName+" 评论了你的"+s.getTargetTypeName(targetType), data)
		return
	}

	notification := &models.Notification{
		UserID:   ownerID,
		Type:     models.NotificationType("comment"),
		Title:    "收到评论",
		Content:  triggerName + " 评论了你的" + s.getTargetTypeName(targetType),
		Priority: 3,
	}
	s.repo.CreateNotification(notification)
}

// GetComment 获取评论详情
func (s *SocialService) GetComment(commentID uint) (*models.Comment, error) {
	return s.repo.GetComment(commentID)
}

// DeleteComment 删除评论
func (s *SocialService) DeleteComment(commentID uint) error {
	return s.repo.DeleteComment(commentID)
}

// ============ 通知服务 ============

// NotificationItem 通知项
type NotificationItem struct {
	models.Notification
}

// GetNotifications 获取通知列表
func (s *SocialService) GetNotifications(userID uint, notifType string, page, pageSize int) ([]NotificationItem, int64, error) {
	notifications, total, err := s.repo.GetNotifications(userID, notifType, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]NotificationItem, len(notifications))
	for i := range notifications {
		items[i] = NotificationItem{Notification: notifications[i]}
	}
	return items, total, nil
}

// GetUnreadCount 获取未读数
func (s *SocialService) GetUnreadCount(userID uint) (int64, error) {
	return s.repo.GetUnreadNotificationCount(userID)
}

// MarkAsRead 标记已读
func (s *SocialService) MarkAsRead(notificationID uint) error {
	return s.repo.MarkNotificationAsRead(notificationID)
}

// MarkAllAsRead 标记全部已读
func (s *SocialService) MarkAllAsRead(userID uint) error {
	return s.repo.MarkAllNotificationsAsRead(userID)
}

// CreateNotification 创建通知
func (s *SocialService) CreateNotification(userID uint, notifType models.NotificationType, title, content string) error {
	notification := &models.Notification{
		UserID:   userID,
		Type:     notifType,
		Title:    title,
		Content:  content,
		Priority: 3,
	}
	return s.repo.CreateNotification(notification)
}

// ============ 关注服务 ============

// ToggleFollow 切换关注状态
func (s *SocialService) ToggleFollow(followerID, followingID uint) (bool, error) {
	if followerID == followingID {
		return false, nil
	}

	isFollowing, err := s.repo.IsFollowing(followerID, followingID)
	if err != nil {
		return false, err
	}

	if isFollowing {
		// 取消关注
		err = s.repo.Unfollow(followerID, followingID)
		if err != nil {
			return false, err
		}
		// 更新统计
		s.repo.IncrementField(followerID, "following_count", -1)
		s.repo.IncrementField(followingID, "followers_count", -1)
		return false, nil
	}

	// 新增关注
	_, err = s.repo.Follow(followerID, followingID)
	if err != nil {
		return false, err
	}

	// 更新统计
	s.repo.IncrementField(followerID, "following_count", 1)
	s.repo.IncrementField(followingID, "followers_count", 1)

	// 发送通知
	go s.sendFollowNotification(followerID, followingID)

	return true, nil
}

func (s *SocialService) sendFollowNotification(followerID, followingID uint) {
	if followerID == followingID {
		return
	}

	triggerUser, _ := s.repo.GetUserByID(followerID)
	triggerName := "用户"
	triggerAvatar := ""
	if triggerUser != nil {
		if triggerUser.Nickname != "" {
			triggerName = triggerUser.Nickname
			triggerAvatar = triggerUser.Avatar
		}
	}

	data := &models.NotificationData{
		TriggerUserID:   followerID,
		TriggerUserName: triggerName,
		TriggerAvatar:   triggerAvatar,
	}

	if s.notificationService != nil {
		s.notificationService.CreateNotification(followingID, models.NotificationTypeFollow, "新增关注", triggerName+" 关注了你", data)
	}
}

// IsFollowing 检查关注状态
func (s *SocialService) IsFollowing(followerID, followingID uint) bool {
	isFollowing, _ := s.repo.IsFollowing(followerID, followingID)
	return isFollowing
}

// GetFollowers 获取粉丝列表
func (s *SocialService) GetFollowers(userID uint, page, pageSize int) ([]models.Follow, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetFollowers(userID, pageSize, offset)
}

// GetFollowing 获取关注列表
func (s *SocialService) GetFollowing(userID uint, page, pageSize int) ([]models.Follow, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetFollowing(userID, pageSize, offset)
}

// GetFollowCounts 获取关注统计
func (s *SocialService) GetFollowCounts(userID uint) (followers int64, following int64, err error) {
	followers, err = s.repo.GetFollowerCount(userID)
	if err != nil {
		return 0, 0, err
	}
	following, err = s.repo.GetFollowingCount(userID)
	if err != nil {
		return 0, 0, err
	}
	return followers, following, nil
}

// ============ 成就服务 ============

// SocialAchievementItem 成就项
type SocialAchievementItem struct {
	models.SocialAchievement
	Achieved bool `json:"achieved"`
}

// GetAllAchievements 获取所有成就
func (s *SocialService) GetAllAchievements(userID uint) ([]SocialAchievementItem, error) {
	allAchievements, err := s.repo.GetAllAchievements()
	if err != nil {
		return nil, err
	}

	userAchievements, err := s.repo.GetUserSocialAchievements(userID)
	if err != nil {
		return nil, err
	}

	achievedMap := make(map[string]bool)
	for _, ua := range userAchievements {
		achievedMap[string(ua.AchievementID)] = true
	}

	items := make([]SocialAchievementItem, len(allAchievements))
	for i := range allAchievements {
		items[i] = SocialAchievementItem{
			SocialAchievement: allAchievements[i],
			Achieved:          achievedMap[string(allAchievements[i].ID)],
		}
	}
	return items, nil
}

// GetUserAchievements 获取用户已获成就
func (s *SocialService) GetUserAchievements(userID uint) ([]SocialAchievementItem, error) {
	allAchievements, err := s.GetAllAchievements(userID)
	if err != nil {
		return nil, err
	}

	var achieved []SocialAchievementItem
	for _, a := range allAchievements {
		if a.Achieved {
			achieved = append(achieved, a)
		}
	}
	return achieved, nil
}

// CheckAndGrantAchievement 检查并授予成就
func (s *SocialService) CheckAndGrantAchievement(userID uint, achievementID string) error {
	has, err := s.repo.HasSocialAchievement(userID, models.SocialAchievementID(achievementID))
	if err != nil || has {
		return err
	}
	return s.repo.GrantSocialAchievement(userID, models.SocialAchievementID(achievementID))
}

// GrantReportAchievement 授予报告成就（外部调用）
func (s *SocialService) GrantReportAchievement(userID uint, totalReports int) {
	// first_report
	s.CheckAndGrantAchievement(userID, "first_report")

	// report_10
	if totalReports >= 10 {
		s.CheckAndGrantAchievement(userID, "report_10")
	}

	// report_100
	if totalReports >= 100 {
		s.CheckAndGrantAchievement(userID, "report_100")
	}
}

func (s *SocialService) checkLikeAchievements(userID uint) {
	s.CheckAndGrantAchievement(userID, "first_like")
}

func (s *SocialService) checkCommentAchievements(userID uint) {
	s.CheckAndGrantAchievement(userID, "first_comment")
}

// ============ 动态帖子服务 ============

// CreatePost 创建动态帖子
func (s *SocialService) CreatePost(userID uint, content string, images []string, roleTag string, targetType string, targetID uint) (*models.Post, error) {
	post := &models.Post{
		UserID:     userID,
		Content:    content,
		RoleTag:    roleTag,
		TargetType: targetType,
		TargetID:   targetID,
	}
	post.SetImagesArray(images)

	if err := s.repo.CreatePost(post); err != nil {
		return nil, err
	}

	// 重新加载用户信息
	loadedPost, err := s.repo.GetPostByID(post.ID)
	if err != nil {
		return post, nil
	}
	return loadedPost, nil
}

// GetFeedPosts 获取动态流
func (s *SocialService) GetFeedPosts(roleTag string, page, pageSize int) ([]models.Post, int64, error) {
	return s.repo.GetFeedPosts(roleTag, page, pageSize)
}

// GetUserPosts 获取用户帖子
func (s *SocialService) GetUserPosts(userID uint, page, pageSize int) ([]models.Post, int64, error) {
	return s.repo.GetUserPosts(userID, page, pageSize)
}

// DeletePost 删除帖子
func (s *SocialService) DeletePost(postID, userID uint) error {
	post, err := s.repo.GetPostByID(postID)
	if err != nil {
		return err
	}
	if post.UserID != userID {
		return err // 权限不足，简单返回错误
	}
	return s.repo.DeletePost(postID)
}

// TogglePostLike 切换帖子点赞（同步更新帖子点赞数）
func (s *SocialService) TogglePostLike(userID uint, postID uint) (bool, error) {
	liked, err := s.repo.IsLiked(userID, "post", postID)
	if err != nil {
		return false, err
	}

	if liked {
		err = s.repo.Unlike(userID, "post", postID)
		if err == nil {
			s.repo.DecrementPostLikes(postID)
		}
		return false, err
	}

	_, err = s.repo.Like(userID, "post", postID)
	if err == nil {
		s.repo.IncrementPostLikes(postID)
		go s.sendLikeNotification(userID, "post", postID)
		go s.checkLikeAchievements(userID)
	}
	return true, err
}

// CreatePostComment 创建帖子评论（同步更新帖子评论数）
func (s *SocialService) CreatePostComment(userID uint, postID uint, parentID *uint, content string) (*CommentItem, error) {
	comment, err := s.CreateComment(userID, "post", postID, parentID, content)
	if err != nil {
		return nil, err
	}
	s.repo.IncrementPostComments(postID)
	return comment, nil
}
