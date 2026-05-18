package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

type OfficialAnalysisTaskController struct {
	service *services.OfficialAnalysisTaskService
}

func NewOfficialAnalysisTaskController(service *services.OfficialAnalysisTaskService) *OfficialAnalysisTaskController {
	return &OfficialAnalysisTaskController{service: service}
}

func (ctrl *OfficialAnalysisTaskController) ListAdminTasks(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 20)
	tasks, total, err := ctrl.service.ListAdminTasks(pagination.Page, pagination.PageSize, services.OfficialTaskListFilters{
		Status:          c.Query("status"),
		Keyword:         c.Query("keyword"),
		MatchName:       c.Query("match_name"),
		AgeGroup:        c.Query("age_group"),
		VisibleLevelMin: c.Query("visible_level_min"),
	})
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取官方选题单失败")
		return
	}
	utils.Success(c, "", gin.H{
		"list":     tasks,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *OfficialAnalysisTaskController) GetAdminTask(c *gin.Context) {
	taskID, ok := parseUintParam(c, "id", "无效的官方选题单ID")
	if !ok {
		return
	}
	task, err := ctrl.service.GetAdminTask(taskID)
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "", gin.H{"task": task})
}

func (ctrl *OfficialAnalysisTaskController) CreateAdminTask(c *gin.Context) {
	var req services.OfficialAnalysisTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	task, err := ctrl.service.CreateTask(req, c.GetUint("userId"), time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "官方选题单已创建", gin.H{"task": task})
}

func (ctrl *OfficialAnalysisTaskController) BatchCreateAdminTasks(c *gin.Context) {
	var req services.OfficialAnalysisTaskBatchCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := ctrl.service.CreateTasksBatch(req, c.GetUint("userId"), time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "官方选题单已批量创建", gin.H{
		"list":  result.Created,
		"total": result.Total,
	})
}

func (ctrl *OfficialAnalysisTaskController) UpdateAdminTask(c *gin.Context) {
	taskID, ok := parseUintParam(c, "id", "无效的官方选题单ID")
	if !ok {
		return
	}
	var req services.OfficialAnalysisTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	task, err := ctrl.service.UpdateTask(taskID, req, time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "官方选题单已更新", gin.H{"task": task})
}

func (ctrl *OfficialAnalysisTaskController) PublishAdminTask(c *gin.Context) {
	taskID, ok := parseUintParam(c, "id", "无效的官方选题单ID")
	if !ok {
		return
	}
	task, err := ctrl.service.PublishTask(taskID, c.GetUint("userId"), time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "官方选题单已发布", gin.H{"task": task})
}

func (ctrl *OfficialAnalysisTaskController) CloseAdminTask(c *gin.Context) {
	taskID, ok := parseUintParam(c, "id", "无效的官方选题单ID")
	if !ok {
		return
	}
	task, err := ctrl.service.CloseTask(taskID, c.GetUint("userId"), time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "官方选题单已关闭", gin.H{"task": task})
}

func (ctrl *OfficialAnalysisTaskController) ListTaskAcceptances(c *gin.Context) {
	taskID, ok := parseUintParam(c, "id", "无效的官方选题单ID")
	if !ok {
		return
	}
	pagination := utils.ParsePaginationWithSize(c, 20)
	acceptances, total, err := ctrl.service.ListTaskAcceptances(taskID, pagination.Page, pagination.PageSize)
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "", gin.H{
		"list":     acceptances,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *OfficialAnalysisTaskController) ListTaskSubmissions(c *gin.Context) {
	taskID, ok := parseUintParam(c, "id", "无效的官方选题单ID")
	if !ok {
		return
	}
	pagination := utils.ParsePaginationWithSize(c, 20)
	submissions, total, err := ctrl.service.ListAdminSubmissions(pagination.Page, pagination.PageSize, services.OfficialTaskListFilters{
		TaskID: taskID,
		Status: c.Query("status"),
	})
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取官方任务提交失败")
		return
	}
	utils.Success(c, "", gin.H{
		"list":     submissions,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *OfficialAnalysisTaskController) ListAdminSubmissions(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 20)
	taskID := uint(0)
	if raw := c.Query("task_id"); raw != "" {
		parsed, err := strconv.ParseUint(raw, 10, 32)
		if err != nil {
			utils.Error(c, http.StatusBadRequest, "无效的官方选题单ID")
			return
		}
		taskID = uint(parsed)
	}
	submissions, total, err := ctrl.service.ListAdminSubmissions(pagination.Page, pagination.PageSize, services.OfficialTaskListFilters{
		Status: c.Query("status"),
		TaskID: taskID,
	})
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取官方任务提交失败")
		return
	}
	utils.Success(c, "", gin.H{
		"list":     submissions,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *OfficialAnalysisTaskController) GetSubmission(c *gin.Context) {
	submissionID, ok := parseUintParam(c, "id", "无效的官方任务提交ID")
	if !ok {
		return
	}
	submission, err := ctrl.service.GetSubmission(submissionID)
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "", gin.H{"submission": submission})
}

func (ctrl *OfficialAnalysisTaskController) ReviewSubmission(c *gin.Context) {
	submissionID, ok := parseUintParam(c, "id", "无效的官方任务提交ID")
	if !ok {
		return
	}
	var req services.OfficialSubmissionReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	submission, err := ctrl.service.ReviewSubmission(submissionID, c.GetUint("userId"), req, time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "官方任务提交已审核", gin.H{"submission": submission})
}

func (ctrl *OfficialAnalysisTaskController) AdoptSubmission(c *gin.Context) {
	submissionID, ok := parseUintParam(c, "id", "无效的官方任务提交ID")
	if !ok {
		return
	}
	var req services.OfficialSubmissionAdoptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	adoption, err := ctrl.service.AdoptSubmission(submissionID, c.GetUint("userId"), req, time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "官方任务提交已采用", gin.H{"adoption": adoption})
}

func (ctrl *OfficialAnalysisTaskController) ListOfficialMaterials(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 20)
	analystID, ok := parseOptionalUintQuery(c, "analyst_id", "无效的分析师ID")
	if !ok {
		return
	}
	materials, total, err := ctrl.service.ListOfficialMaterials(pagination.Page, pagination.PageSize, services.OfficialMaterialListFilters{
		Keyword:        c.Query("keyword"),
		MatchName:      c.Query("match_name"),
		AgeGroup:       c.Query("age_group"),
		Channel:        c.Query("channel"),
		AdoptionStatus: c.Query("adoption_status"),
		AssetKind:      c.Query("asset_kind"),
		PublishReady:   c.Query("publish_ready") == "true",
		PublishStatus:  c.Query("publish_status"),
		MetricsStatus:  c.Query("metrics_status"),
		BonusStatus:    c.Query("bonus_status"),
		SortBy:         c.Query("sort_by"),
		AnalystID:      analystID,
	})
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取官方素材库失败")
		return
	}
	utils.Success(c, "", gin.H{
		"list":     materials,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *OfficialAnalysisTaskController) CreatePublishRecord(c *gin.Context) {
	adoptionID, ok := parseUintParam(c, "id", "无效的官方采用记录ID")
	if !ok {
		return
	}
	var req services.OfficialPublishRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	record, err := ctrl.service.CreatePublishRecord(adoptionID, c.GetUint("userId"), req, time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "官方发布记录已保存", gin.H{"publish_record": record})
}

func (ctrl *OfficialAnalysisTaskController) UpdatePublishRecord(c *gin.Context) {
	adoptionID, ok := parseUintParam(c, "id", "无效的官方采用记录ID")
	if !ok {
		return
	}
	recordID, ok := parseUintParam(c, "recordId", "无效的官方发布记录ID")
	if !ok {
		return
	}
	var req services.OfficialPublishRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	record, err := ctrl.service.UpdatePublishRecord(adoptionID, recordID, req, time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "官方发布记录已更新", gin.H{"publish_record": record})
}

func (ctrl *OfficialAnalysisTaskController) DeletePublishRecord(c *gin.Context) {
	adoptionID, ok := parseUintParam(c, "id", "无效的官方采用记录ID")
	if !ok {
		return
	}
	recordID, ok := parseUintParam(c, "recordId", "无效的官方发布记录ID")
	if !ok {
		return
	}
	if err := ctrl.service.DeletePublishRecord(adoptionID, recordID); err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "官方发布记录已删除", gin.H{"id": recordID})
}

func (ctrl *OfficialAnalysisTaskController) ListAdminEventTopics(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 20)
	topics, total, err := ctrl.service.ListAdminEventTopics(pagination.Page, pagination.PageSize, services.OfficialEventTopicFilters{
		Keyword:      c.Query("keyword"),
		AgeGroup:     c.Query("age_group"),
		FeaturedOnly: c.Query("featured_only") == "true",
	})
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取专题运营列表失败")
		return
	}
	utils.Success(c, "", gin.H{
		"list":     topics,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *OfficialAnalysisTaskController) SaveAdminEventTopic(c *gin.Context) {
	matchName := c.Param("matchName")
	if strings.TrimSpace(matchName) == "" {
		utils.Error(c, http.StatusBadRequest, "无效的专题赛事名")
		return
	}
	var req services.OfficialEventTopicConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	req.MatchName = matchName
	config, err := ctrl.service.SaveEventTopicConfig(req, c.GetUint("userId"), time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "专题运营配置已保存", gin.H{"topic_config": config})
}

func (ctrl *OfficialAnalysisTaskController) ListPublicEventTopics(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 20)
	topics, total, err := ctrl.service.ListPublicEventTopics(pagination.Page, pagination.PageSize, services.OfficialEventTopicFilters{
		Keyword:      c.Query("keyword"),
		AgeGroup:     c.Query("age_group"),
		FeaturedOnly: c.Query("featured_only") == "true",
	})
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取官方赛事专题失败")
		return
	}
	utils.Success(c, "", gin.H{
		"list":     topics,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *OfficialAnalysisTaskController) GetPublicEventTopic(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 20)
	topic, total, err := ctrl.service.GetPublicEventTopic(c.Param("matchName"), pagination.Page, pagination.PageSize, c.Query("age_group"))
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "", gin.H{
		"topic":    topic,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *OfficialAnalysisTaskController) CreatePlaybackBonus(c *gin.Context) {
	adoptionID, ok := parseUintParam(c, "id", "无效的官方采用记录ID")
	if !ok {
		return
	}
	var req services.OfficialPlaybackBonusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	reward, err := ctrl.service.CreatePlaybackBonus(adoptionID, c.GetUint("userId"), req, time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "播放表现奖金已记录", gin.H{"reward": reward})
}

func (ctrl *OfficialAnalysisTaskController) UpdateAdoptionPublic(c *gin.Context) {
	adoptionID, ok := parseUintParam(c, "id", "无效的官方采用记录ID")
	if !ok {
		return
	}
	var req services.OfficialAdoptionPublicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	adoption, err := ctrl.service.UpdateAdoptionPublic(adoptionID, req, time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "官方采用展示状态已更新", gin.H{"adoption": adoption})
}

func (ctrl *OfficialAnalysisTaskController) ListAdminRewards(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 20)
	analystID, ok := parseOptionalUintQuery(c, "analyst_id", "无效的分析师ID")
	if !ok {
		return
	}
	batchID, ok := parseOptionalUintQuery(c, "settlement_batch_id", "无效的结算批次ID")
	if !ok {
		return
	}
	rewards, total, err := ctrl.service.ListAdminRewards(pagination.Page, pagination.PageSize, services.OfficialRewardListFilters{
		Status:     c.Query("status"),
		RewardType: c.Query("reward_type"),
		SourceType: c.Query("source_type"),
		AnalystID:  analystID,
		BatchID:    batchID,
	})
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取官方奖励记录失败")
		return
	}
	utils.Success(c, "", gin.H{
		"list":     rewards,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *OfficialAnalysisTaskController) ListRewardSettlementBatches(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 20)
	batches, total, err := ctrl.service.ListRewardSettlementBatches(pagination.Page, pagination.PageSize, services.OfficialRewardBatchListFilters{
		Status: c.Query("status"),
	})
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取官方奖励结算批次失败")
		return
	}
	utils.Success(c, "", gin.H{
		"list":     batches,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *OfficialAnalysisTaskController) GetRewardSettlementBatch(c *gin.Context) {
	batchID, ok := parseUintParam(c, "id", "无效的结算批次ID")
	if !ok {
		return
	}
	batch, err := ctrl.service.GetRewardSettlementBatch(batchID)
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "", gin.H{"batch": batch})
}

func (ctrl *OfficialAnalysisTaskController) SettleReward(c *gin.Context) {
	rewardID, ok := parseUintParam(c, "id", "无效的奖励记录ID")
	if !ok {
		return
	}
	var req services.OfficialRewardActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	reward, err := ctrl.service.SettleReward(rewardID, c.GetUint("userId"), req, time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "官方奖励已结算", gin.H{"reward": reward})
}

func (ctrl *OfficialAnalysisTaskController) BatchSettleRewards(c *gin.Context) {
	var req services.OfficialRewardBatchSettleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := ctrl.service.BatchSettleRewards(req, c.GetUint("userId"), time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "官方奖励已批量结算", result)
}

func (ctrl *OfficialAnalysisTaskController) ReverseReward(c *gin.Context) {
	rewardID, ok := parseUintParam(c, "id", "无效的奖励记录ID")
	if !ok {
		return
	}
	var req services.OfficialRewardActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	reward, err := ctrl.service.ReverseReward(rewardID, c.GetUint("userId"), req, time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "官方奖励已冲正", gin.H{"reward": reward})
}

func (ctrl *OfficialAnalysisTaskController) ListAvailableTasks(c *gin.Context) {
	analyst, ok := ctrl.currentAnalyst(c)
	if !ok {
		return
	}
	pagination := utils.ParsePaginationWithSize(c, 20)
	tasks, total, err := ctrl.service.ListAvailableTasksForAnalyst(analyst.ID, pagination.Page, pagination.PageSize, time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "", gin.H{
		"list":     tasks,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *OfficialAnalysisTaskController) GetAvailableTask(c *gin.Context) {
	analyst, ok := ctrl.currentAnalyst(c)
	if !ok {
		return
	}
	taskID, ok := parseUintParam(c, "id", "无效的官方选题单ID")
	if !ok {
		return
	}
	task, err := ctrl.service.GetAvailableTaskForAnalyst(analyst.ID, taskID, time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "", gin.H{"task": task})
}

func (ctrl *OfficialAnalysisTaskController) ListMyTasks(c *gin.Context) {
	analyst, ok := ctrl.currentAnalyst(c)
	if !ok {
		return
	}
	pagination := utils.ParsePaginationWithSize(c, 20)
	acceptances, total, err := ctrl.service.ListMyAcceptances(analyst.ID, pagination.Page, pagination.PageSize)
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "", gin.H{
		"list":     acceptances,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *OfficialAnalysisTaskController) ListMyRewards(c *gin.Context) {
	analyst, ok := ctrl.currentAnalyst(c)
	if !ok {
		return
	}
	pagination := utils.ParsePaginationWithSize(c, 20)
	rewards, total, err := ctrl.service.ListAnalystRewards(analyst.ID, pagination.Page, pagination.PageSize)
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "", gin.H{
		"list":     rewards,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *OfficialAnalysisTaskController) ListMyAdoptions(c *gin.Context) {
	analyst, ok := ctrl.currentAnalyst(c)
	if !ok {
		return
	}
	pagination := utils.ParsePaginationWithSize(c, 20)
	adoptions, total, err := ctrl.service.ListAnalystAdoptions(analyst.ID, pagination.Page, pagination.PageSize)
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "", gin.H{
		"list":     adoptions,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *OfficialAnalysisTaskController) AcceptTask(c *gin.Context) {
	analyst, ok := ctrl.currentAnalyst(c)
	if !ok {
		return
	}
	taskID, ok := parseUintParam(c, "id", "无效的官方选题单ID")
	if !ok {
		return
	}
	acceptance, err := ctrl.service.AcceptTask(analyst.ID, taskID, time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "接取成功", gin.H{"acceptance": acceptance})
}

func (ctrl *OfficialAnalysisTaskController) SubmitTask(c *gin.Context) {
	analyst, ok := ctrl.currentAnalyst(c)
	if !ok {
		return
	}
	taskID, ok := parseUintParam(c, "id", "无效的官方选题单ID")
	if !ok {
		return
	}
	var req services.OfficialTaskSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	submission, err := ctrl.service.SubmitTask(analyst.ID, taskID, req, time.Now())
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return
	}
	utils.Success(c, "官方任务已提交审核", gin.H{"submission": submission})
}

func (ctrl *OfficialAnalysisTaskController) currentAnalyst(c *gin.Context) (*models.Analyst, bool) {
	userID := c.GetUint("userId")
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return nil, false
	}
	analyst, err := ctrl.service.FindActiveAnalystByUserID(userID)
	if err != nil {
		ctrl.writeOfficialTaskError(c, err)
		return nil, false
	}
	return analyst, true
}

func (ctrl *OfficialAnalysisTaskController) writeOfficialTaskError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrOfficialTaskNotFound):
		utils.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrOfficialTaskInvalid):
		utils.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrOfficialSubmissionInvalid):
		utils.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrOfficialSubmissionNotFound):
		utils.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrOfficialAdoptionNotFound):
		utils.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrOfficialPublishRecordNotFound):
		utils.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrAnalystRewardNotFound):
		utils.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrAnalystRewardInvalid):
		utils.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrOfficialTaskDuplicate):
		utils.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrOfficialTaskFull):
		utils.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrOfficialTaskLevelDenied):
		utils.Error(c, http.StatusForbidden, err.Error())
	case errors.Is(err, services.ErrOfficialTaskDailyLimit):
		utils.Error(c, http.StatusTooManyRequests, err.Error())
	case errors.Is(err, services.ErrOfficialTaskAnalystInvalid):
		utils.Error(c, http.StatusForbidden, err.Error())
	case errors.Is(err, services.ErrOfficialTaskUnavailable):
		utils.Error(c, http.StatusBadRequest, err.Error())
	default:
		utils.Error(c, http.StatusInternalServerError, "官方选题单操作失败")
	}
}

func parseUintParam(c *gin.Context, key, message string) (uint, bool) {
	raw := c.Param(key)
	id, err := strconv.ParseUint(raw, 10, 32)
	if err != nil || id == 0 {
		utils.Error(c, http.StatusBadRequest, message)
		return 0, false
	}
	return uint(id), true
}

func parseOptionalUintQuery(c *gin.Context, key, message string) (uint, bool) {
	raw := c.Query(key)
	if raw == "" {
		return 0, true
	}
	id, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, message)
		return 0, false
	}
	return uint(id), true
}
