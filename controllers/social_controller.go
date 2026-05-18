package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

// SocialController 社交互动控制器
type SocialController struct {
	socialService *services.SocialService
}

// NewSocialController 创建社交控制器
func NewSocialController(socialService *services.SocialService) *SocialController {
	return &SocialController{socialService: socialService}
}

// ToggleLike 切换点赞状态
func (c *SocialController) ToggleLike(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	var data struct {
		TargetType string `json:"target_type" binding:"required"`
		TargetID   uint   `json:"target_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&data); err != nil {
		utils.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	liked, err := c.socialService.ToggleLike(userID, data.TargetType, data.TargetID)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "操作失败")
		return
	}

	msg := "取消点赞成功"
	if liked {
		msg = "点赞成功"
	}
	utils.Success(ctx, msg, map[string]interface{}{"liked": liked})
}

// RemoveLike 取消点赞
func (c *SocialController) RemoveLike(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	targetType := ctx.Query("targetType")
	targetIDStr := ctx.Query("targetId")
	targetID, _ := strconv.ParseUint(targetIDStr, 10, 64)

	_, err := c.socialService.ToggleLike(userID, targetType, uint(targetID))
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "操作失败")
		return
	}

	utils.Success(ctx, "取消点赞成功", nil)
}

// GetLikes 获取点赞列表
func (c *SocialController) GetLikes(ctx *gin.Context) {
	targetType := ctx.Query("targetType")
	targetIDStr := ctx.Query("targetId")
	targetID, _ := strconv.ParseUint(targetIDStr, 10, 64)
	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	likes, total, err := c.socialService.GetTargetLikes(targetType, uint(targetID), page, pageSize)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取失败")
		return
	}

	utils.Success(ctx, "查询成功", gin.H{
		"list":  likes,
		"total": total,
	})
}

// GetMyLikes 获取我的点赞列表
func (c *SocialController) GetMyLikes(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	likes, total, err := c.socialService.GetUserLikes(userID, page, pageSize)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取失败")
		return
	}

	utils.Success(ctx, "查询成功", gin.H{
		"list":  likes,
		"total": total,
	})
}

// ToggleFavorite 切换收藏状态
func (c *SocialController) ToggleFavorite(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	var data struct {
		TargetType string `json:"target_type" binding:"required"`
		TargetID   uint   `json:"target_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&data); err != nil {
		utils.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	favorited, err := c.socialService.ToggleFavorite(userID, data.TargetType, data.TargetID)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "操作失败")
		return
	}

	msg := "取消收藏成功"
	if favorited {
		msg = "收藏成功"
	}
	utils.Success(ctx, msg, map[string]interface{}{"favorited": favorited})
}

// RemoveFavorite 取消收藏
func (c *SocialController) RemoveFavorite(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	targetType := ctx.Query("targetType")
	targetIDStr := ctx.Query("targetId")
	targetID, _ := strconv.ParseUint(targetIDStr, 10, 64)

	_, err := c.socialService.ToggleFavorite(userID, targetType, uint(targetID))
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "操作失败")
		return
	}

	utils.Success(ctx, "取消收藏成功", nil)
}

// GetMyFavorites 获取我的收藏列表
func (c *SocialController) GetMyFavorites(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	targetType := ctx.Query("type")
	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	favorites, total, err := c.socialService.GetUserFavorites(userID, targetType, page, pageSize)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取失败")
		return
	}

	utils.Success(ctx, "查询成功", gin.H{
		"list":  favorites,
		"total": total,
	})
}

// GetComments 获取评论列表
func (c *SocialController) GetComments(ctx *gin.Context) {
	targetType := ctx.Query("targetType")
	if targetType == "" {
		targetType = ctx.Query("target_type")
	}
	targetIDStr := ctx.Query("targetId")
	if targetIDStr == "" {
		targetIDStr = ctx.Query("target_id")
	}
	targetID, _ := strconv.ParseUint(targetIDStr, 10, 64)
	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	comments, total, err := c.socialService.GetComments(targetType, uint(targetID), page, pageSize)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取失败")
		return
	}

	utils.Success(ctx, "查询成功", gin.H{
		"list":  comments,
		"total": total,
	})
}

// CreateComment 创建评论
func (c *SocialController) CreateComment(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	var data struct {
		TargetType string `json:"target_type" binding:"required"`
		TargetID   uint   `json:"target_id" binding:"required"`
		ParentID   *uint  `json:"parent_id"`
		Content    string `json:"content" binding:"required,max=500"`
	}
	if err := ctx.ShouldBindJSON(&data); err != nil {
		utils.Error(ctx, http.StatusBadRequest, "评论内容不能超过500字")
		return
	}

	var comment *services.CommentItem
	var err error
	if data.TargetType == "post" {
		comment, err = c.socialService.CreatePostComment(userID, data.TargetID, data.ParentID, data.Content)
	} else {
		comment, err = c.socialService.CreateComment(userID, data.TargetType, data.TargetID, data.ParentID, data.Content)
	}
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "评论失败")
		return
	}

	utils.Success(ctx, "评论成功", comment)
}

// DeleteComment 删除评论
func (c *SocialController) DeleteComment(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	commentID, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)

	comment, err := c.socialService.GetComment(uint(commentID))
	if err != nil {
		utils.Error(ctx, http.StatusNotFound, "评论不存在")
		return
	}
	if comment.UserID != userID {
		utils.Error(ctx, http.StatusForbidden, "无权删除此评论")
		return
	}

	err = c.socialService.DeleteComment(uint(commentID))
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "删除失败")
		return
	}

	utils.Success(ctx, "删除成功", nil)
}

// GetNotifications 获取通知列表
func (c *SocialController) GetNotifications(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	notifType := ctx.Query("type")
	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	notifications, total, err := c.socialService.GetNotifications(userID, notifType, page, pageSize)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取失败")
		return
	}

	// 将 Data 字符串解析为对象返回给前端
	result := make([]gin.H, 0, len(notifications))
	for _, n := range notifications {
		item := gin.H{
			"id":         n.ID,
			"user_id":    n.UserID,
			"type":       n.Type,
			"title":      n.Title,
			"content":    n.Content,
			"is_read":    n.IsRead,
			"priority":   n.Priority,
			"created_at": n.CreatedAt,
		}
		if n.Data != "" {
			var dataMap map[string]interface{}
			if json.Unmarshal([]byte(n.Data), &dataMap) == nil {
				item["data"] = dataMap
			}
		}
		result = append(result, item)
	}

	utils.Success(ctx, "查询成功", gin.H{
		"list":  result,
		"total": total,
	})
}

// GetUnreadCount 获取未读通知数
func (c *SocialController) GetUnreadCount(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	count, err := c.socialService.GetUnreadCount(userID)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取失败")
		return
	}

	utils.Success(ctx, "查询成功", count)
}

// MarkAllRead 标记全部已读
func (c *SocialController) MarkAllRead(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	err := c.socialService.MarkAllAsRead(userID)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "操作失败")
		return
	}

	utils.Success(ctx, "操作成功", nil)
}

// GetAchievements 获取成就列表
func (c *SocialController) GetAchievements(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	achievements, err := c.socialService.GetAllAchievements(userID)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取失败")
		return
	}

	utils.Success(ctx, "查询成功", achievements)
}

// ============ 动态帖子接口 ============

// GetFeed 获取动态流
func (c *SocialController) GetFeed(ctx *gin.Context) {
	roleTag := ctx.DefaultQuery("role_tag", "all")
	province := ctx.Query("province")
	city := ctx.Query("city")
	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize
	userIDStr := ctx.Query("user_id")

	var posts []models.Post
	var total int64
	var err error

	if userIDStr != "" {
		userID, parseErr := strconv.ParseUint(userIDStr, 10, 64)
		if parseErr == nil && userID > 0 {
			posts, total, err = c.socialService.GetUserPosts(uint(userID), page, pageSize)
		} else {
			utils.Error(ctx, http.StatusBadRequest, "user_id 参数无效")
			return
		}
	} else {
		posts, total, err = c.socialService.GetFeedPosts(roleTag, province, city, page, pageSize)
	}

	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取失败")
		return
	}

	userID := ctx.GetUint("userId")
	result := make([]gin.H, 0, len(posts))
	for _, p := range posts {
		isLiked := false
		if userID > 0 {
			isLiked = c.socialService.IsLiked(userID, "post", p.ID)
		}
		authorName := ""
		authorAvatar := ""
		authorRole := ""
		authorProvince := ""
		authorCity := ""
		if p.User != nil {
			authorName = p.User.Name
			if p.User.Nickname != "" {
				authorName = p.User.Nickname
			}
			authorAvatar = p.User.Avatar
			authorRole = string(p.User.Role)
			authorProvince = p.User.Province
			authorCity = p.User.City
		}
		result = append(result, gin.H{
			"id":             p.ID,
			"user_id":        p.UserID,
			"content":        p.Content,
			"images":         p.GetImagesArray(),
			"target_type":    p.TargetType,
			"target_id":      p.TargetID,
			"role_tag":       p.RoleTag,
			"likes_count":    p.LikesCount,
			"comments_count": p.CommentsCount,
			"is_top":         p.IsTop,
			"is_liked":       isLiked,
			"created_at":     p.CreatedAt,
			"author": gin.H{
				"name":     authorName,
				"avatar":   authorAvatar,
				"role":     authorRole,
				"province": authorProvince,
				"city":     authorCity,
			},
		})
	}

	utils.Success(ctx, "查询成功", gin.H{
		"list":  result,
		"total": total,
	})
}

// CreatePost 发布动态
func (c *SocialController) CreatePost(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	var req struct {
		Content    string   `json:"content" binding:"required,max=2000"`
		Images     []string `json:"images"`
		RoleTag    string   `json:"role_tag"`
		TargetType string   `json:"target_type"`
		TargetID   uint     `json:"target_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Error(ctx, http.StatusBadRequest, "参数错误：内容不能为空且不超过2000字")
		return
	}

	post, err := c.socialService.CreatePost(userID, req.Content, req.Images, req.RoleTag, req.TargetType, req.TargetID)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "发布失败")
		return
	}

	utils.Success(ctx, "发布成功", post)
}

// DeletePost 删除动态
func (c *SocialController) DeletePost(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	postID, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err := c.socialService.DeletePost(uint(postID), userID); err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "删除失败")
		return
	}

	utils.Success(ctx, "删除成功", nil)
}

// TogglePostLike 帖子点赞/取消点赞
func (c *SocialController) TogglePostLike(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	postID, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	liked, err := c.socialService.TogglePostLike(userID, uint(postID))
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "操作失败")
		return
	}

	msg := "取消点赞成功"
	if liked {
		msg = "点赞成功"
	}
	utils.Success(ctx, msg, gin.H{"liked": liked})
}

// ============ 关注接口 ============

// ToggleFollow 关注/取消关注用户
func (c *SocialController) ToggleFollow(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "请先登录")
		return
	}

	var req struct {
		FollowingID uint `json:"following_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Error(ctx, http.StatusBadRequest, "参数错误：缺少被关注用户ID")
		return
	}

	if userID == req.FollowingID {
		utils.Error(ctx, http.StatusBadRequest, "不能关注自己")
		return
	}

	following, err := c.socialService.ToggleFollow(userID, req.FollowingID)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "操作失败")
		return
	}

	// 获取最新统计
	followersCount, followingCount, _ := c.socialService.GetFollowCounts(req.FollowingID)

	msg := "取消关注成功"
	if following {
		msg = "关注成功"
	}
	utils.Success(ctx, msg, gin.H{
		"following":       following,
		"follower_count":  followersCount,
		"following_count": followingCount,
	})
}

// GetFollowers 获取用户粉丝列表
func (c *SocialController) GetFollowers(ctx *gin.Context) {
	userIDStr := ctx.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "用户ID格式错误")
		return
	}

	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	follows, total, err := c.socialService.GetFollowers(uint(userID), page, pageSize)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取失败")
		return
	}

	result := make([]gin.H, 0, len(follows))
	for _, f := range follows {
		if f.Follower != nil {
			result = append(result, gin.H{
				"id":          f.Follower.ID,
				"nickname":    f.Follower.Nickname,
				"avatar":      f.Follower.Avatar,
				"role":        f.Follower.Role,
				"followed_at": f.CreatedAt,
			})
		}
	}

	utils.Success(ctx, "查询成功", gin.H{
		"list":  result,
		"total": total,
	})
}

// GetFollowing 获取用户关注列表
func (c *SocialController) GetFollowing(ctx *gin.Context) {
	userIDStr := ctx.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "用户ID格式错误")
		return
	}

	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	follows, total, err := c.socialService.GetFollowing(uint(userID), page, pageSize)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取失败")
		return
	}

	result := make([]gin.H, 0, len(follows))
	for _, f := range follows {
		if f.Following != nil {
			result = append(result, gin.H{
				"id":          f.Following.ID,
				"nickname":    f.Following.Nickname,
				"avatar":      f.Following.Avatar,
				"role":        f.Following.Role,
				"followed_at": f.CreatedAt,
			})
		}
	}

	utils.Success(ctx, "查询成功", gin.H{
		"list":  result,
		"total": total,
	})
}

// GetFollowStatus 获取关注状态
func (c *SocialController) GetFollowStatus(ctx *gin.Context) {
	currentUserID := ctx.GetUint("userId")

	targetIDStr := ctx.Param("userId")
	targetID, err := strconv.ParseUint(targetIDStr, 10, 64)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "用户ID格式错误")
		return
	}

	isFollowing := false
	isFollowedBy := false
	if currentUserID > 0 {
		isFollowing = c.socialService.IsFollowing(currentUserID, uint(targetID))
		isFollowedBy = c.socialService.IsFollowing(uint(targetID), currentUserID)
	}

	followersCount, followingCount, _ := c.socialService.GetFollowCounts(uint(targetID))

	utils.Success(ctx, "查询成功", gin.H{
		"is_following":    isFollowing,
		"is_followed_by":  isFollowedBy,
		"is_mutual":       isFollowing && isFollowedBy,
		"followers_count": followersCount,
		"following_count": followingCount,
	})
}
