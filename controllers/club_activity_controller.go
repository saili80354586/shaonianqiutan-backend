package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
	"gorm.io/gorm"
)

// ClubActivityController 俱乐部活动控制器
type ClubActivityController struct {
	db                  *gorm.DB
	notificationService *services.NotificationService
}

// NewClubActivityController 创建控制器
func NewClubActivityController(db *gorm.DB, notificationService *services.NotificationService) *ClubActivityController {
	return &ClubActivityController{db: db, notificationService: notificationService}
}

// 活动类型映射（兼容老数据 external/internal -> 前端类型）
func mapActivityType(t string) string {
	switch t {
	case "external":
		return "trial"
	case "internal":
		return "camp"
	case "trial", "camp", "match", "exchange":
		return t
	default:
		return "trial"
	}
}

func applyActivityTimeRange(ctx *gin.Context, db *gorm.DB, defaultDays int) *gorm.DB {
	startAfter := ctx.Query("startAfter")
	startBefore := ctx.Query("startBefore")
	days := ctx.Query("days")
	if days == "" && ctx.Query("timeRange") == "future30" {
		days = "30"
	}
	if days == "" && startAfter == "" && startBefore == "" && defaultDays > 0 {
		days = strconv.Itoa(defaultDays)
	}

	if days != "" {
		if n, err := strconv.Atoi(days); err == nil && n > 0 {
			now := time.Now()
			return db.Where("start_time >= ? AND start_time < ?", now, now.AddDate(0, 0, n))
		}
	}

	if startAfter != "" {
		if t, err := time.ParseInLocation("2006-01-02", startAfter, time.Local); err == nil {
			db = db.Where("start_time >= ?", t)
		}
	}
	if startBefore != "" {
		if t, err := time.ParseInLocation("2006-01-02", startBefore, time.Local); err == nil {
			db = db.Where("start_time < ?", t.AddDate(0, 0, 1))
		}
	}
	return db
}

// 常见城市经纬度映射
var cityCoordMap = map[string][2]float64{
	"北京": {116.4074, 39.9042}, "上海": {121.4737, 31.2304}, "广州": {113.2644, 23.1291},
	"深圳": {114.0579, 22.5431}, "成都": {104.0668, 30.5728}, "杭州": {120.1551, 30.2741},
	"武汉": {114.3054, 30.5931}, "西安": {108.9398, 34.3416}, "南京": {118.7969, 32.0603},
	"重庆": {106.5516, 29.563}, "天津": {117.2009, 39.0842}, "苏州": {120.5853, 31.2989},
	"长沙": {112.9388, 28.2282}, "郑州": {113.6253, 34.7466}, "沈阳": {123.4315, 41.8057},
	"青岛": {120.3826, 36.0671}, "宁波": {121.5509, 29.875}, "东莞": {113.7518, 23.0207},
	"佛山": {113.1214, 23.0215}, "合肥": {117.2272, 31.8206}, "昆明": {102.8329, 24.8801},
	"大连": {121.6147, 38.914}, "厦门": {118.0894, 24.4798}, "哈尔滨": {126.535, 45.8038},
	"济南": {117.1205, 36.651}, "长春": {125.3235, 43.8171}, "南宁": {108.3665, 22.817},
	"贵阳": {106.6302, 26.6477}, "福州": {119.2965, 26.0745}, "太原": {112.5489, 37.8706},
	"石家庄": {114.5149, 38.0423}, "南昌": {115.854, 28.683}, "兰州": {103.8343, 36.0611},
	"海口": {110.3492, 20.0174}, "呼和浩特": {111.7492, 40.8426}, "乌鲁木齐": {87.6168, 43.8256},
	"银川": {106.2309, 38.4872}, "西宁": {101.7782, 36.6171}, "拉萨": {91.1409, 29.6456},
}

var provinceNames = []string{
	"北京", "天津", "上海", "重庆", "河北", "山西", "辽宁", "吉林", "黑龙江",
	"江苏", "浙江", "安徽", "福建", "江西", "山东", "河南", "湖北", "湖南",
	"广东", "海南", "四川", "贵州", "云南", "陕西", "甘肃", "青海", "台湾",
	"内蒙古", "广西", "西藏", "宁夏", "新疆", "香港", "澳门",
}

var activeActivityRegistrationStatuses = []string{"pending", "confirmed", "checked_in"}

type activityRegistrationRequest struct {
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	Wechat         string `json:"wechat"`
	Remark         string `json:"remark"`
	PlayerName     string `json:"player_name"`
	PlayerAge      int    `json:"player_age"`
	PlayerPosition string `json:"player_position"`
	ContactPhone   string `json:"contact_phone"`
}

// 从 location 中解析 province / city / address
func parseLocation(location string) (province, city, address string) {
	if location == "" {
		return "", "", ""
	}

	// 1. 直辖市特殊处理（location 可能直接以"北京朝阳区..."开头）
	for _, muni := range []string{"北京", "上海", "天津", "重庆"} {
		if strings.HasPrefix(location, muni) {
			province = muni + "市"
			city = muni + "市"
			remaining := strings.TrimPrefix(location, muni)
			remaining = strings.TrimLeft(remaining, "市")
			address = remaining
			return
		}
	}

	// 2. 先识别省份前缀
	remaining := location
	for _, p := range provinceNames {
		if strings.HasPrefix(location, p) {
			province = p + "省"
			if p == "内蒙古" || p == "广西" || p == "西藏" || p == "宁夏" || p == "新疆" {
				province = p + "自治区"
			} else if p == "香港" || p == "澳门" {
				province = p + "特别行政区"
			}
			remaining = strings.TrimPrefix(location, p)
			remaining = strings.TrimLeft(remaining, "省市")
			break
		}
	}

	// 3. 从剩余部分识别城市
	for c := range cityCoordMap {
		if strings.HasPrefix(remaining, c) {
			city = c + "市"
			address = strings.TrimPrefix(remaining, c)
			return
		}
	}

	// 4. 若 location 直接以城市名开头（无省份前缀，如"成都武侯区..."）
	for c := range cityCoordMap {
		if strings.HasPrefix(location, c) {
			city = c + "市"
			address = strings.TrimPrefix(location, c)
			return
		}
	}

	// 5. fallback：取剩余部分前2个字作为城市
	runes := []rune(remaining)
	if len(runes) >= 2 {
		candidate := string(runes[:2])
		city = candidate + "市"
		if len(runes) > 2 {
			address = string(runes[2:])
		}
	} else {
		city = remaining
	}
	return
}

func getCityCoord(city string) (lat, lng float64) {
	// 去掉 "市" 后缀匹配
	c := strings.TrimSuffix(city, "市")
	if coord, ok := cityCoordMap[c]; ok {
		return coord[1], coord[0] // lat, lng
	}
	return 0, 0
}

// ListActivities 获取俱乐部活动列表（公开）
func (c *ClubActivityController) ListActivities(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	activityType := ctx.Query("type")           // external, internal, 空表示全部
	status := ctx.Query("status")               // upcoming, ongoing, ended, 空表示全部
	isReview := ctx.Query("isReview")           // true, false
	publishStatus := ctx.Query("publishStatus") // draft, published, unpublished

	db := c.db.Where("club_id = ?", clubID)
	if activityType != "" {
		db = db.Where("type = ?", activityType)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if isReview == "true" {
		db = db.Where("is_review = ? AND status = ?", true, "ended")
	}
	if publishStatus != "" {
		if publishStatus != "all" {
			db = db.Where("publish_status = ?", publishStatus)
		}
	} else {
		// 默认只显示已发布的活动
		db = db.Where("publish_status = ?", "published")
	}

	var activities []models.ClubActivity
	if err := db.Order("start_time DESC").Find(&activities).Error; err != nil {
		utils.ServerError(ctx, "查询失败")
		return
	}

	result := make([]map[string]interface{}, 0, len(activities))
	for _, a := range activities {
		regCount := int64(0)
		c.db.Model(&models.ClubActivityRegistration{}).Where("activity_id = ? AND status IN ?", a.ID, activeActivityRegistrationStatuses).Count(&regCount)
		result = append(result, map[string]interface{}{
			"id":              a.ID,
			"title":           a.Title,
			"type":            a.Type,
			"status":          a.Status,
			"description":     a.Description,
			"coverImage":      a.CoverImage,
			"startTime":       a.StartTime.Format("2006-01-02 15:04"),
			"endTime":         a.EndTime.Format("2006-01-02 15:04"),
			"location":        a.Location,
			"maxParticipants": a.MaxParticipants,
			"contactPhone":    a.ContactPhone,
			"contactWechat":   a.ContactWechat,
			"isReview":        a.IsReview,
			"reviewContent":   a.ReviewContent,
			"reviewImages":    a.GetReviewImagesArray(),
			"publishStatus":   a.PublishStatus,
			"regCount":        regCount,
		})
	}

	utils.SuccessResponse(ctx, result)
}

// CreateActivity 创建活动
func (c *ClubActivityController) CreateActivity(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	var req struct {
		Title           string   `json:"title" binding:"required"`
		Type            string   `json:"type"`
		Description     string   `json:"description"`
		CoverImage      string   `json:"coverImage"`
		StartTime       string   `json:"startTime"`
		EndTime         string   `json:"endTime"`
		Location        string   `json:"location"`
		MaxParticipants int      `json:"maxParticipants"`
		ContactPhone    string   `json:"contactPhone"`
		ContactWechat   string   `json:"contactWechat"`
		PublishStatus   string   `json:"publishStatus"`
		IsReview        bool     `json:"isReview"`
		ReviewContent   string   `json:"reviewContent"`
		ReviewImages    []string `json:"reviewImages"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	start, _ := time.ParseInLocation("2006-01-02T15:04", req.StartTime, time.Local)
	end, _ := time.ParseInLocation("2006-01-02T15:04", req.EndTime, time.Local)
	if start.IsZero() {
		start, _ = time.ParseInLocation("2006-01-02 15:04", req.StartTime, time.Local)
	}
	if end.IsZero() {
		end, _ = time.ParseInLocation("2006-01-02 15:04", req.EndTime, time.Local)
	}
	if req.PublishStatus == "" {
		req.PublishStatus = "published"
	}

	status := "upcoming"
	now := time.Now()
	if !start.IsZero() && now.After(start) {
		if !end.IsZero() && now.After(end) {
			status = "ended"
		} else {
			status = "ongoing"
		}
	}

	activity := models.ClubActivity{
		ClubID:          uint(clubID),
		Title:           req.Title,
		Type:            req.Type,
		Description:     req.Description,
		CoverImage:      req.CoverImage,
		StartTime:       start,
		EndTime:         end,
		Location:        req.Location,
		MaxParticipants: req.MaxParticipants,
		ContactPhone:    req.ContactPhone,
		ContactWechat:   req.ContactWechat,
		PublishStatus:   req.PublishStatus,
		Status:          status,
		IsReview:        req.IsReview,
		ReviewContent:   req.ReviewContent,
	}
	activity.SetReviewImagesArray(req.ReviewImages)

	if err := c.db.Create(&activity).Error; err != nil {
		utils.ServerError(ctx, "创建失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{"id": activity.ID}, "创建成功")
}

// UpdateActivity 更新活动
func (c *ClubActivityController) UpdateActivity(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}
	activityIDStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的活动ID")
		return
	}

	var activity models.ClubActivity
	if err := c.db.Where("id = ? AND club_id = ?", activityID, clubID).First(&activity).Error; err != nil {
		utils.NotFoundError(ctx, "活动不存在")
		return
	}

	var req struct {
		Title           string   `json:"title"`
		Type            string   `json:"type"`
		Description     string   `json:"description"`
		CoverImage      string   `json:"coverImage"`
		StartTime       string   `json:"startTime"`
		EndTime         string   `json:"endTime"`
		Location        string   `json:"location"`
		MaxParticipants int      `json:"maxParticipants"`
		ContactPhone    string   `json:"contactPhone"`
		ContactWechat   string   `json:"contactWechat"`
		PublishStatus   string   `json:"publishStatus"`
		IsReview        bool     `json:"isReview"`
		ReviewContent   string   `json:"reviewContent"`
		ReviewImages    []string `json:"reviewImages"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if req.Title != "" {
		activity.Title = req.Title
	}
	if req.Type != "" {
		activity.Type = req.Type
	}
	activity.Description = req.Description
	activity.CoverImage = req.CoverImage
	activity.Location = req.Location
	activity.MaxParticipants = req.MaxParticipants
	activity.ContactPhone = req.ContactPhone
	activity.ContactWechat = req.ContactWechat
	if req.PublishStatus != "" {
		activity.PublishStatus = req.PublishStatus
	}
	activity.IsReview = req.IsReview
	activity.ReviewContent = req.ReviewContent
	activity.SetReviewImagesArray(req.ReviewImages)

	if req.StartTime != "" {
		start, _ := time.ParseInLocation("2006-01-02T15:04", req.StartTime, time.Local)
		if start.IsZero() {
			start, _ = time.ParseInLocation("2006-01-02 15:04", req.StartTime, time.Local)
		}
		if !start.IsZero() {
			activity.StartTime = start
		}
	}
	if req.EndTime != "" {
		end, _ := time.ParseInLocation("2006-01-02T15:04", req.EndTime, time.Local)
		if end.IsZero() {
			end, _ = time.ParseInLocation("2006-01-02 15:04", req.EndTime, time.Local)
		}
		if !end.IsZero() {
			activity.EndTime = end
		}
	}

	// 自动计算状态
	now := time.Now()
	if now.Before(activity.StartTime) {
		activity.Status = "upcoming"
	} else if now.After(activity.EndTime) {
		activity.Status = "ended"
	} else {
		activity.Status = "ongoing"
	}

	if err := c.db.Save(&activity).Error; err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// DeleteActivity 删除活动
func (c *ClubActivityController) DeleteActivity(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}
	activityIDStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的活动ID")
		return
	}

	if err := c.db.Where("id = ? AND club_id = ?", activityID, clubID).Delete(&models.ClubActivity{}).Error; err != nil {
		utils.ServerError(ctx, "删除失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "删除成功")
}

// RegisterActivity 报名活动
func (c *ClubActivityController) RegisterActivity(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubIDValue, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}
	activityIDStr := ctx.Param("id")
	activityIDValue, err := strconv.ParseUint(activityIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的活动ID")
		return
	}

	clubID := uint(clubIDValue)
	c.registerActivity(ctx, uint(activityIDValue), &clubID)
}

// RegisterPublicActivity 通过全局活动路径报名
func (c *ClubActivityController) RegisterPublicActivity(ctx *gin.Context) {
	activityIDStr := ctx.Param("id")
	activityIDValue, err := strconv.ParseUint(activityIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的活动ID")
		return
	}

	c.registerActivity(ctx, uint(activityIDValue), nil)
}

func (c *ClubActivityController) registerActivity(ctx *gin.Context, activityID uint, clubID *uint) {
	var activity models.ClubActivity
	db := c.db.Where("id = ?", activityID)
	if clubID != nil {
		db = db.Where("club_id = ?", *clubID)
	}
	if err := db.First(&activity).Error; err != nil {
		utils.NotFoundError(ctx, "活动不存在")
		return
	}

	if activity.PublishStatus != "published" {
		utils.NotFoundError(ctx, "活动不存在")
		return
	}

	if activity.Status == "ended" {
		utils.ValidationError(ctx, "活动已结束，无法报名")
		return
	}

	var req activityRegistrationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	// 兼容前端字段命名
	name := req.Name
	if name == "" {
		name = req.PlayerName
	}
	phone := req.Phone
	if phone == "" {
		phone = req.ContactPhone
	}
	if name == "" {
		utils.ValidationError(ctx, "请填写球员姓名")
		return
	}
	if phone == "" {
		utils.ValidationError(ctx, "请填写联系电话")
		return
	}
	if req.Remark == "" && req.PlayerPosition != "" {
		req.Remark = "位置：" + req.PlayerPosition
		if req.PlayerAge > 0 {
			req.Remark += "，年龄：" + strconv.Itoa(req.PlayerAge)
		}
	}

	// 检查报名人数
	if activity.MaxParticipants > 0 {
		var count int64
		c.db.Model(&models.ClubActivityRegistration{}).Where("activity_id = ? AND status IN ?", activity.ID, activeActivityRegistrationStatuses).Count(&count)
		if int(count) >= activity.MaxParticipants {
			utils.ValidationError(ctx, "报名人数已满")
			return
		}
	}

	var userID *uint
	if uid, exists := ctx.Get("userId"); exists {
		id := uid.(uint)
		userID = &id
	}
	if userID != nil {
		var existing models.ClubActivityRegistration
		err := c.db.Where("activity_id = ? AND user_id = ? AND status IN ?", activity.ID, *userID, activeActivityRegistrationStatuses).First(&existing).Error
		if err == nil {
			utils.ValidationError(ctx, "你已报名该活动，请勿重复报名")
			return
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			utils.ServerError(ctx, "报名状态检查失败")
			return
		}
	}

	reg := models.ClubActivityRegistration{
		ActivityID: activity.ID,
		UserID:     userID,
		Name:       name,
		Phone:      phone,
		Wechat:     req.Wechat,
		Remark:     req.Remark,
		Status:     "pending",
	}

	if err := c.db.Create(&reg).Error; err != nil {
		utils.ServerError(ctx, "报名失败")
		return
	}

	// 发送通知给俱乐部管理员
	go c.notifyClubAdminNewRegistration(activity, reg)

	utils.SuccessResponseWithMessage(ctx, gin.H{"id": reg.ID}, "报名成功")
}

// ListRegistrations 获取活动报名列表
func (c *ClubActivityController) ListRegistrations(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	_, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}
	activityIDStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的活动ID")
		return
	}

	var regs []models.ClubActivityRegistration
	if err := c.db.Where("activity_id = ?", activityID).Order("created_at DESC").Find(&regs).Error; err != nil {
		utils.ServerError(ctx, "查询失败")
		return
	}

	utils.SuccessResponse(ctx, regs)
}

// UpdateRegistrationStatus 更新报名状态
func (c *ClubActivityController) UpdateRegistrationStatus(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	_, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}
	regIDStr := ctx.Param("regId")
	regID, err := strconv.ParseUint(regIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的报名ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	var reg models.ClubActivityRegistration
	if err := c.db.First(&reg, regID).Error; err != nil {
		utils.NotFoundError(ctx, "报名记录不存在")
		return
	}

	if err := c.db.Model(&models.ClubActivityRegistration{}).Where("id = ?", regID).Update("status", req.Status).Error; err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	// 发送审批结果通知
	go func() {
		var activity models.ClubActivity
		if err := c.db.First(&activity, reg.ActivityID).Error; err == nil {
			c.notifyRegistrationResult(activity, reg, req.Status)
		}
	}()

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// PublishActivity 发布活动
func (c *ClubActivityController) PublishActivity(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}
	activityIDStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的活动ID")
		return
	}

	if err := c.db.Model(&models.ClubActivity{}).
		Where("id = ? AND club_id = ?", activityID, clubID).
		Update("publish_status", "published").Error; err != nil {
		utils.ServerError(ctx, "发布失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "发布成功")
}

// UnpublishActivity 下架活动
func (c *ClubActivityController) UnpublishActivity(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}
	activityIDStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的活动ID")
		return
	}

	if err := c.db.Model(&models.ClubActivity{}).
		Where("id = ? AND club_id = ?", activityID, clubID).
		Update("publish_status", "unpublished").Error; err != nil {
		utils.ServerError(ctx, "下架失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "下架成功")
}

// BatchUpdateRegistrationStatus 批量更新报名状态
func (c *ClubActivityController) BatchUpdateRegistrationStatus(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	_, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}
	activityIDStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的活动ID")
		return
	}

	var req struct {
		IDs    []uint `json:"ids" binding:"required"`
		Status string `json:"status" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	var regs []models.ClubActivityRegistration
	c.db.Where("activity_id = ? AND id IN ?", activityID, req.IDs).Find(&regs)

	if err := c.db.Model(&models.ClubActivityRegistration{}).
		Where("activity_id = ? AND id IN ?", activityID, req.IDs).
		Update("status", req.Status).Error; err != nil {
		utils.ServerError(ctx, "批量更新失败")
		return
	}

	// 发送审批结果通知
	go func() {
		var activity models.ClubActivity
		if err := c.db.First(&activity, activityID).Error; err == nil {
			for _, reg := range regs {
				c.notifyRegistrationResult(activity, reg, req.Status)
			}
		}
	}()

	utils.SuccessResponseWithMessage(ctx, nil, "批量更新成功")
}

// ExportRegistrations 导出活动报名记录为 CSV
func (c *ClubActivityController) ExportRegistrations(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	_, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}
	activityIDStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的活动ID")
		return
	}

	var regs []models.ClubActivityRegistration
	if err := c.db.Where("activity_id = ?", activityID).Order("created_at DESC").Find(&regs).Error; err != nil {
		utils.ServerError(ctx, "查询失败")
		return
	}

	var csvBuilder strings.Builder
	csvBuilder.WriteString("\uFEFF姓名,手机号,微信号,备注,状态,报名时间\n")
	for _, r := range regs {
		statusLabel := "待审核"
		if r.Status == "confirmed" {
			statusLabel = "已通过"
		} else if r.Status == "rejected" {
			statusLabel = "已拒绝"
		} else if r.Status == "checked_in" {
			statusLabel = "已签到"
		} else if r.Status == "cancelled" {
			statusLabel = "已取消"
		}
		csvBuilder.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s\n",
			r.Name, r.Phone, r.Wechat, r.Remark, statusLabel, r.CreatedAt.Format("2006-01-02 15:04")))
	}

	ctx.Header("Content-Type", "text/csv; charset=utf-8")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=activity_%d_registrations.csv", activityID))
	ctx.String(http.StatusOK, csvBuilder.String())
}

// ========== 公开活动 API ==========

// ListPublicActivities 获取全部公开活动列表
func (c *ClubActivityController) ListPublicActivities(ctx *gin.Context) {
	activityType := ctx.Query("type")
	status := ctx.Query("status")
	province := ctx.Query("province")
	city := ctx.Query("city")

	db := c.db.Where("publish_status = ?", "published")
	if activityType != "" {
		db = db.Where("type = ?", activityType)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if province != "" {
		db = db.Where("location LIKE ?", "%"+province+"%")
	}
	if city != "" {
		db = db.Where("location LIKE ?", "%"+city+"%")
	}
	db = applyActivityTimeRange(ctx, db, 0)

	var activities []models.ClubActivity
	if err := db.Order("start_time ASC").Find(&activities).Error; err != nil {
		utils.ServerError(ctx, "查询失败")
		return
	}

	result := make([]map[string]interface{}, 0, len(activities))
	for _, a := range activities {
		regCount := int64(0)
		c.db.Model(&models.ClubActivityRegistration{}).Where("activity_id = ? AND status IN ?", a.ID, activeActivityRegistrationStatuses).Count(&regCount)
		prov, cit, addr := parseLocation(a.Location)
		clubName := ""
		clubLogo := ""
		var club models.Club
		if err := c.db.First(&club, a.ClubID).Error; err == nil {
			clubName = club.Name
			clubLogo = club.Logo
		}
		result = append(result, map[string]interface{}{
			"id":                  a.ID,
			"clubId":              a.ClubID,
			"clubName":            clubName,
			"clubLogo":            clubLogo,
			"title":               a.Title,
			"type":                mapActivityType(a.Type),
			"status":              a.Status,
			"description":         a.Description,
			"coverImage":          a.CoverImage,
			"startTime":           a.StartTime.Format("2006-01-02 15:04"),
			"endTime":             a.EndTime.Format("2006-01-02 15:04"),
			"location":            a.Location,
			"province":            prov,
			"city":                cit,
			"address":             addr,
			"maxParticipants":     a.MaxParticipants,
			"currentParticipants": regCount,
			"fee":                 0,
			"feeType":             "free",
			"contactPhone":        a.ContactPhone,
			"contactWechat":       a.ContactWechat,
			"publishStatus":       a.PublishStatus,
			"createdAt":           a.CreatedAt.Format("2006-01-02 15:04"),
			"updatedAt":           a.UpdatedAt.Format("2006-01-02 15:04"),
		})
	}

	utils.SuccessResponse(ctx, result)
}

// GetPublicActivity 获取公开活动详情
func (c *ClubActivityController) GetPublicActivity(ctx *gin.Context) {
	activityIDStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的活动ID")
		return
	}

	var activity models.ClubActivity
	if err := c.db.First(&activity, activityID).Error; err != nil {
		utils.NotFoundError(ctx, "活动不存在")
		return
	}

	// 未发布的活动不对外公开
	if activity.PublishStatus != "published" {
		utils.NotFoundError(ctx, "活动不存在")
		return
	}

	regCount := int64(0)
	c.db.Model(&models.ClubActivityRegistration{}).Where("activity_id = ? AND status IN ?", activity.ID, activeActivityRegistrationStatuses).Count(&regCount)
	prov, cit, addr := parseLocation(activity.Location)
	clubName := ""
	clubLogo := ""
	var club models.Club
	if err := c.db.First(&club, activity.ClubID).Error; err == nil {
		clubName = club.Name
		clubLogo = club.Logo
	}

	utils.SuccessResponse(ctx, map[string]interface{}{
		"id":                  activity.ID,
		"clubId":              activity.ClubID,
		"clubName":            clubName,
		"clubLogo":            clubLogo,
		"title":               activity.Title,
		"type":                mapActivityType(activity.Type),
		"status":              activity.Status,
		"description":         activity.Description,
		"coverImage":          activity.CoverImage,
		"startTime":           activity.StartTime.Format("2006-01-02 15:04"),
		"endTime":             activity.EndTime.Format("2006-01-02 15:04"),
		"location":            activity.Location,
		"province":            prov,
		"city":                cit,
		"address":             addr,
		"maxParticipants":     activity.MaxParticipants,
		"currentParticipants": regCount,
		"fee":                 0,
		"feeType":             "free",
		"contactPhone":        activity.ContactPhone,
		"contactWechat":       activity.ContactWechat,
		"isReview":            activity.IsReview,
		"reviewContent":       activity.ReviewContent,
		"reviewImages":        activity.GetReviewImagesArray(),
		"publishStatus":       activity.PublishStatus,
		"createdAt":           activity.CreatedAt.Format("2006-01-02 15:04"),
		"updatedAt":           activity.UpdatedAt.Format("2006-01-02 15:04"),
	})
}

// GetActivitiesMap 获取活动地图数据（按城市聚合）
func (c *ClubActivityController) GetActivitiesMap(ctx *gin.Context) {
	var activities []models.ClubActivity
	db := c.db.Where("publish_status = ? AND status IN ?", "published", []string{"upcoming", "ongoing"})
	db = applyActivityTimeRange(ctx, db, 30)
	if err := db.Order("start_time ASC").Find(&activities).Error; err != nil {
		utils.ServerError(ctx, "查询失败")
		return
	}

	// 按城市分组
	type cityGroup struct {
		province     string
		city         string
		lat          float64
		lng          float64
		activityType string
		activities   []map[string]interface{}
	}
	groups := make(map[string]*cityGroup)

	for _, a := range activities {
		prov, cit, _ := parseLocation(a.Location)
		lat, lng := getCityCoord(cit)
		if cit == "" {
			cit = a.Location
		}
		if _, ok := groups[cit]; !ok {
			groups[cit] = &cityGroup{province: prov, city: cit, lat: lat, lng: lng, activityType: mapActivityType(a.Type)}
		}
		groups[cit].activities = append(groups[cit].activities, map[string]interface{}{
			"id":        a.ID,
			"title":     a.Title,
			"type":      mapActivityType(a.Type),
			"startTime": a.StartTime.Format("2006-01-02 15:04"),
		})
	}

	result := make([]map[string]interface{}, 0, len(groups))
	for _, g := range groups {
		result = append(result, map[string]interface{}{
			"province":   g.province,
			"city":       strings.TrimSuffix(g.city, "市"),
			"type":       g.activityType,
			"lat":        g.lat,
			"lng":        g.lng,
			"count":      len(g.activities),
			"activities": g.activities,
		})
	}

	utils.SuccessResponse(ctx, result)
}

// CancelRegistration 取消报名
func (c *ClubActivityController) CancelRegistration(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubIDValue, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}
	activityIDStr := ctx.Param("id")
	activityIDValue, err := strconv.ParseUint(activityIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的活动ID")
		return
	}

	clubID := uint(clubIDValue)
	c.cancelRegistration(ctx, uint(activityIDValue), &clubID)
}

// CancelPublicRegistration 通过全局活动路径取消报名
func (c *ClubActivityController) CancelPublicRegistration(ctx *gin.Context) {
	activityIDStr := ctx.Param("id")
	activityIDValue, err := strconv.ParseUint(activityIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的活动ID")
		return
	}

	c.cancelRegistration(ctx, uint(activityIDValue), nil)
}

func (c *ClubActivityController) cancelRegistration(ctx *gin.Context, activityID uint, clubID *uint) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.ValidationError(ctx, "请先登录")
		return
	}

	if clubID != nil {
		var activity models.ClubActivity
		if err := c.db.Where("id = ? AND club_id = ?", activityID, *clubID).First(&activity).Error; err != nil {
			utils.NotFoundError(ctx, "活动不存在")
			return
		}
	}

	result := c.db.Model(&models.ClubActivityRegistration{}).
		Where("activity_id = ? AND user_id = ? AND status IN ?", activityID, userID, activeActivityRegistrationStatuses).
		Update("status", "cancelled")
	if result.Error != nil {
		utils.ServerError(ctx, "取消报名失败")
		return
	}
	if result.RowsAffected == 0 {
		utils.ValidationError(ctx, "未找到可取消的报名")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "取消报名成功")
}

// GetMyRegistrations 获取我的报名列表
func (c *ClubActivityController) GetMyRegistrations(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.ValidationError(ctx, "请先登录")
		return
	}

	status := ctx.Query("status")

	db := c.db.Where("user_id = ?", userID)
	if status != "" {
		db = db.Where("status = ?", status)
	}

	var regs []models.ClubActivityRegistration
	if err := db.Order("created_at DESC").Find(&regs).Error; err != nil {
		utils.ServerError(ctx, "查询失败")
		return
	}

	result := make([]map[string]interface{}, 0, len(regs))
	for _, r := range regs {
		var activity models.ClubActivity
		c.db.First(&activity, r.ActivityID)
		clubName := ""
		clubLogo := ""
		var club models.Club
		if err := c.db.First(&club, activity.ClubID).Error; err == nil {
			clubName = club.Name
			clubLogo = club.Logo
		}
		result = append(result, map[string]interface{}{
			"id":                r.ID,
			"activityId":        r.ActivityID,
			"activity_id":       r.ActivityID,
			"clubId":            activity.ClubID,
			"clubName":          clubName,
			"clubLogo":          clubLogo,
			"activityTitle":     activity.Title,
			"activityCover":     activity.CoverImage,
			"activityType":      activity.Type,
			"activityStatus":    activity.Status,
			"activityLocation":  activity.Location,
			"activityStartTime": activity.StartTime.Format("2006-01-02 15:04"),
			"activityEndTime":   activity.EndTime.Format("2006-01-02 15:04"),
			"maxParticipants":   activity.MaxParticipants,
			"contactPhone":      activity.ContactPhone,
			"contactWechat":     activity.ContactWechat,
			"name":              r.Name,
			"phone":             r.Phone,
			"wechat":            r.Wechat,
			"remark":            r.Remark,
			"status":            r.Status,
			"created_at":        r.CreatedAt,
			"createdAt":         r.CreatedAt,
		})
	}

	utils.SuccessResponse(ctx, result)
}

// notifyClubAdminNewRegistration 通知俱乐部管理员有新报名
func (c *ClubActivityController) notifyClubAdminNewRegistration(activity models.ClubActivity, reg models.ClubActivityRegistration) {
	if c.notificationService == nil {
		return
	}
	var club models.Club
	if err := c.db.First(&club, activity.ClubID).Error; err != nil {
		return
	}
	if club.UserID == 0 {
		return
	}
	title := "活动有新报名"
	content := fmt.Sprintf("%s 报名了活动《%s》", reg.Name, activity.Title)
	data := &models.NotificationData{
		TargetType: "club_activity",
		TargetID:   activity.ID,
	}
	c.notificationService.CreateNotification(club.UserID, models.NotificationTypeActivityRegistration, title, content, data)
}

// notifyRegistrationResult 通知报名者审批结果
func (c *ClubActivityController) notifyRegistrationResult(activity models.ClubActivity, reg models.ClubActivityRegistration, status string) {
	if c.notificationService == nil || reg.UserID == nil || *reg.UserID == 0 {
		return
	}
	title := "报名结果通知"
	content := ""
	notifType := models.NotificationTypeActivityApproved
	if status == "confirmed" {
		content = fmt.Sprintf("你报名的活动《%s》已通过审核", activity.Title)
	} else if status == "cancelled" || status == "rejected" {
		content = fmt.Sprintf("你报名的活动《%s》未通过审核", activity.Title)
		notifType = models.NotificationTypeActivityRejected
	} else {
		return
	}
	data := &models.NotificationData{
		TargetType: "club_activity",
		TargetID:   activity.ID,
	}
	c.notificationService.CreateNotification(*reg.UserID, notifType, title, content, data)
}
