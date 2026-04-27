package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
	"gorm.io/gorm"
)

// VideoAnalysisController 视频分析控制器
type VideoAnalysisController struct {
	db             *gorm.DB
	analysisRepo   *models.VideoAnalysisRepository
	highlightRepo *models.AnalysisHighlightRepository
	aiService     *services.AIService
	reportGen     *services.ReportGenerator
}

// NewVideoAnalysisController 创建视频分析控制器
func NewVideoAnalysisController(db *gorm.DB, aiService *services.AIService) *VideoAnalysisController {
	return &VideoAnalysisController{
		db:             db,
		analysisRepo:   models.NewVideoAnalysisRepository(db),
		highlightRepo:  models.NewAnalysisHighlightRepository(db),
		aiService:      aiService,
		reportGen:      services.NewReportGenerator("./uploads/reports"),
	}
}

// UpdateScoresRequest 更新评分请求
type UpdateScoresRequest struct {
	Scores       *models.VideoAnalysisScores `json:"scores"`
	Summary      string                     `json:"summary"`
	Improvements string                     `json:"improvements"`
	AnalystNotes string                     `json:"analyst_notes"`
}

// UpdateScores 更新评分
func (ctrl *VideoAnalysisController) UpdateScores(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	var req UpdateScoresRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	analysis, err := ctrl.analysisRepo.FindByID(uint(id))
	if err != nil || analysis == nil {
		utils.Error(c, http.StatusNotFound, "分析记录不存在")
		return
	}

	overallScore := req.Scores.CalculateOverallScore()
	potentialLevel := models.GetPotentialLevel(overallScore)

	scoresJSON, err := req.Scores.ToJSON()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "评分序列化失败")
		return
	}

	updates := map[string]interface{}{
		"scores":          scoresJSON,
		"overall_score":   overallScore,
		"potential_level": potentialLevel,
		"summary":         req.Summary,
		"improvements":    req.Improvements,
		"analyst_notes":   req.AnalystNotes,
		"status":          models.AnalysisStatusScoring,
	}

	err = ctrl.analysisRepo.Update(uint(id), updates)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "保存评分失败")
		return
	}

	utils.Success(c, "评分保存成功", gin.H{
		"overall_score":   overallScore,
		"potential_level": potentialLevel,
	})
}

// ConfirmAnalysis 确认并生成 MD 文档
func (ctrl *VideoAnalysisController) ConfirmAnalysis(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	analysis, err := ctrl.analysisRepo.FindByID(uint(id))
	if err != nil || analysis == nil {
		utils.Error(c, http.StatusNotFound, "分析记录不存在")
		return
	}

	// 获取球员信息
	var user models.User
	if err := ctrl.db.First(&user, analysis.UserID).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取球员信息失败")
		return
	}

	// 获取分析师名称
	analystName := "未知分析师"
	var analyst models.Analyst
	if err := ctrl.db.First(&analyst, analysis.AnalystID).Error; err == nil {
		analystName = analyst.Name
	}

	// 生成 MD 文档
	ratingMD, playerInfoMD, err := ctrl.reportGen.GenerateFromVideoAnalysis(analysis, analystName, &user)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "生成文档失败: "+err.Error())
		return
	}

	// 更新记录
	updates := map[string]interface{}{
		"rating_report_md": ratingMD,
		"player_info_md":   playerInfoMD,
		"status":           models.AnalysisStatusCompleted,
	}
	if err := ctrl.analysisRepo.Update(uint(id), updates); err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新记录失败")
		return
	}

	utils.Success(c, "文档生成成功", gin.H{
		"rating_report_md": ratingMD,
		"player_info_md":   playerInfoMD,
	})
}

// GetAnalysis 获取分析详情
func (ctrl *VideoAnalysisController) GetAnalysis(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	analysis, err := ctrl.analysisRepo.FindByID(uint(id))
	if err != nil || analysis == nil {
		utils.Error(c, http.StatusNotFound, "分析记录不存在")
		return
	}

	scores, _ := models.ParseScoresFromJSON(analysis.Scores)
	highlights, _ := ctrl.highlightRepo.FindByAnalysisID(uint(id))

	utils.Success(c, "", gin.H{
		"analysis":  analysis,
		"scores":    scores,
		"highlights": highlights,
	})
}

// GetAnalysisByOrder 根据订单获取分析
func (ctrl *VideoAnalysisController) GetAnalysisByOrder(c *gin.Context) {
	orderIDStr := c.Query("order_id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	analysis, err := ctrl.analysisRepo.FindByOrderID(uint(orderID))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	if analysis == nil {
		utils.Success(c, "", nil)
		return
	}

	scores, _ := models.ParseScoresFromJSON(analysis.Scores)
	highlights, _ := ctrl.highlightRepo.FindByAnalysisID(analysis.ID)

	utils.Success(c, "", gin.H{
		"analysis":   analysis,
		"scores":     scores,
		"highlights": highlights,
	})
}

// CreateHighlightRequest 创建高光请求
type CreateHighlightRequest struct {
	AnalysisID      uint                      `json:"analysis_id"`
	Timestamp       string                    `json:"timestamp"`
	TagType         models.HighlightTagType   `json:"tag_type"`
	Description     string                    `json:"description"`
	VideoClipURL    string                    `json:"video_clip_url"`
	IncludeInReport bool                      `json:"include_in_report"`
}

// CreateHighlight 创建高光标记
func (ctrl *VideoAnalysisController) CreateHighlight(c *gin.Context) {
	var req CreateHighlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	highlight := &models.AnalysisHighlight{
		AnalysisID:      req.AnalysisID,
		Timestamp:       req.Timestamp,
		TagType:         req.TagType,
		Description:     req.Description,
		VideoClipURL:    req.VideoClipURL,
		IncludeInReport: req.IncludeInReport,
	}

	err := ctrl.highlightRepo.Create(highlight)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建高光失败")
		return
	}

	utils.Success(c, "高光标记成功", highlight)
}

// UpdateHighlight 更新高光
func (ctrl *VideoAnalysisController) UpdateHighlight(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的高光ID")
		return
	}

	var req CreateHighlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	updates := map[string]interface{}{
		"timestamp":         req.Timestamp,
		"tag_type":          req.TagType,
		"description":       req.Description,
		"video_clip_url":    req.VideoClipURL,
		"include_in_report": req.IncludeInReport,
	}

	err = ctrl.highlightRepo.Update(uint(id), updates)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新高光失败")
		return
	}

	utils.Success(c, "更新成功", nil)
}

// DeleteHighlight 删除高光
func (ctrl *VideoAnalysisController) DeleteHighlight(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的高光ID")
		return
	}

	err = ctrl.highlightRepo.Delete(uint(id))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "删除高光失败")
		return
	}

	utils.Success(c, "删除成功", nil)
}

// GetHighlights 获取分析的所有高光
func (ctrl *VideoAnalysisController) GetHighlights(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	highlights, err := ctrl.highlightRepo.FindByAnalysisID(uint(id))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	utils.Success(c, "", highlights)
}

// GenerateAIReportRequest AI报告生成请求
type GenerateAIReportRequest struct {
	AnalysisID uint `json:"analysis_id"`
}

// GenerateAIReport 触发AI生成报告（异步）
func (ctrl *VideoAnalysisController) GenerateAIReport(c *gin.Context) {
	var req GenerateAIReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	analysis, err := ctrl.analysisRepo.FindByID(req.AnalysisID)
	if err != nil || analysis == nil {
		utils.Error(c, http.StatusNotFound, "分析记录不存在")
		return
	}

	if analysis.OverallScore == 0 {
		utils.Error(c, http.StatusBadRequest, "请先完成评分")
		return
	}

	// 如果已经在生成中，直接返回
	if analysis.AIReportStatus == "generating" {
		utils.Success(c, "AI报告生成中，请耐心等待", nil)
		return
	}

	// 更新状态为生成中
	ctrl.analysisRepo.Update(req.AnalysisID, map[string]interface{}{
		"ai_report_status": "generating",
	})

	// 异步在后台生成报告，避免前端请求超时
	go func(analysisID uint, currentVersion int) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[AIReport] panic recovered for analysis %d: %v", analysisID, r)
				ctrl.analysisRepo.Update(analysisID, map[string]interface{}{
					"ai_report_status": "failed",
				})
			}
		}()

		analysis, err := ctrl.analysisRepo.FindByID(analysisID)
		if err != nil || analysis == nil {
			ctrl.analysisRepo.Update(analysisID, map[string]interface{}{
				"ai_report_status": "failed",
			})
			return
		}

		scores, _ := models.ParseScoresFromJSON(analysis.Scores)
		highlights, _ := ctrl.highlightRepo.FindIncludedInReport(analysis.ID)

		var highlightInputs []services.HighlightInput
		for _, h := range highlights {
			highlightInputs = append(highlightInputs, services.HighlightInput{
				Timestamp:   h.Timestamp,
				Description: h.Description,
			})
		}

		reportInput := &services.VideoAnalysisReportInput{
			PlayerName:      analysis.PlayerName,
			PlayerAge:       analysis.PlayerAge,
			PlayerPosition:  analysis.PlayerPosition,
			PlayerFoot:      analysis.PlayerFoot,
			PlayerHeight:    analysis.PlayerHeight,
			PlayerWeight:    analysis.PlayerWeight,
			PlayerTeam:      analysis.PlayerTeam,
			MatchName:       analysis.MatchName,
			MatchDate:       analysis.MatchDate,
			MatchType:       analysis.MatchType,
			OpponentLevel:   analysis.OpponentLevel,
			Opponent:        analysis.Opponent,
			PlayTime:        analysis.PlayTime,
			Goals:           analysis.Goals,
			Assists:         analysis.Assists,
			OverallScore:    analysis.OverallScore,
			PotentialLevel:  string(analysis.PotentialLevel),
			Highlights:      highlightInputs,
			Summary:         analysis.Summary,
			Improvements:    analysis.Improvements,
			AnalystNotes:    analysis.AnalystNotes,
		}

		scoresInput := services.ScoresInput{
			BallControl:        services.ScoreInput{Score: scores.BallControl.Score, Weight: scores.BallControl.Weight, Comment: scores.BallControl.Comment},
			OffBallMovement:    services.ScoreInput{Score: scores.OffBallMovement.Score, Weight: scores.OffBallMovement.Weight, Comment: scores.OffBallMovement.Comment},
			PressingAwareness:  services.ScoreInput{Score: scores.PressingAwareness.Score, Weight: scores.PressingAwareness.Weight, Comment: scores.PressingAwareness.Comment},
			Positioning:        services.ScoreInput{Score: scores.Positioning.Score, Weight: scores.Positioning.Weight, Comment: scores.Positioning.Comment},
			WidthParticipation: services.ScoreInput{Score: scores.WidthParticipation.Score, Weight: scores.WidthParticipation.Weight, Comment: scores.WidthParticipation.Comment},
			OffBallSupport:     services.ScoreInput{Score: scores.OffBallSupport.Score, Weight: scores.OffBallSupport.Weight, Comment: scores.OffBallSupport.Comment},
			OneVOne:            services.ScoreInput{Score: scores.OneVOne.Score, Weight: scores.OneVOne.Weight, Comment: scores.OneVOne.Comment},
			CrossingAssist:     services.ScoreInput{Score: scores.CrossingAssist.Score, Weight: scores.CrossingAssist.Weight, Comment: scores.CrossingAssist.Comment},
			CombatAbility:      services.ScoreInput{Score: scores.CombatAbility.Score, Weight: scores.CombatAbility.Weight, Comment: scores.CombatAbility.Comment},
			PaceRhythm:         services.ScoreInput{Score: scores.PaceRhythm.Score, Weight: scores.PaceRhythm.Weight, Comment: scores.PaceRhythm.Comment},
			PassVision:         services.ScoreInput{Score: scores.PassVision.Score, Weight: scores.PassVision.Weight, Comment: scores.PassVision.Comment},
			BodyPosture:        services.ScoreInput{Score: scores.BodyPosture.Score, Weight: scores.BodyPosture.Weight, Comment: scores.BodyPosture.Comment},
			DefensiveCommitment:  services.ScoreInput{Score: scores.DefensiveCommitment.Score, Weight: scores.DefensiveCommitment.Weight, Comment: scores.DefensiveCommitment.Comment},
			LossRecovery:         services.ScoreInput{Score: scores.LossRecovery.Score, Weight: scores.LossRecovery.Weight, Comment: scores.LossRecovery.Comment},
			TeammateCoordination: services.ScoreInput{Score: scores.TeammateCoordination.Score, Weight: scores.TeammateCoordination.Weight, Comment: scores.TeammateCoordination.Comment},
			SecondBall:           services.ScoreInput{Score: scores.SecondBall.Score, Weight: scores.SecondBall.Weight, Comment: scores.SecondBall.Comment},
			AerialDuel:           services.ScoreInput{Score: scores.AerialDuel.Score, Weight: scores.AerialDuel.Weight, Comment: scores.AerialDuel.Comment},
			DefensiveShape:       services.ScoreInput{Score: scores.DefensiveShape.Score, Weight: scores.DefensiveShape.Weight, Comment: scores.DefensiveShape.Comment},
			RoleAdjustment:       services.ScoreInput{Score: scores.RoleAdjustment.Score, Weight: scores.RoleAdjustment.Weight, Comment: scores.RoleAdjustment.Comment},
			DefensiveRhythm:      services.ScoreInput{Score: scores.DefensiveRhythm.Score, Weight: scores.DefensiveRhythm.Weight, Comment: scores.DefensiveRhythm.Comment},
		}
		reportInput.Scores = scoresInput

		prompt := services.BuildReportPrompt(reportInput)
		aiReport, err := ctrl.aiService.GenerateReport(prompt)
		if err != nil {
			log.Printf("[AIReport] generation failed for analysis %d: %v", analysisID, err)
			ctrl.analysisRepo.Update(analysisID, map[string]interface{}{
				"ai_report_status": "failed",
			})
			return
		}

		newVersion := currentVersion + 1
		ctrl.analysisRepo.Update(analysisID, map[string]interface{}{
			"ai_report":         aiReport,
			"ai_report_status":  "draft",
			"ai_report_version": newVersion,
		})
	}(analysis.ID, analysis.AIReportVersion)

	utils.Success(c, "AI报告生成任务已提交，预计需要3-5分钟", gin.H{
		"status": "generating",
	})
}

// UpdateAIReport 手动修改AI报告
func (ctrl *VideoAnalysisController) UpdateAIReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	var req struct {
		Report string `json:"report"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	analysis, err := ctrl.analysisRepo.FindByID(uint(id))
	if err != nil || analysis == nil {
		utils.Error(c, http.StatusNotFound, "分析记录不存在")
		return
	}

	updates := map[string]interface{}{
		"ai_report":        req.Report,
		"ai_report_status": "draft",
	}

	err = ctrl.analysisRepo.Update(uint(id), updates)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新报告失败")
		return
	}

	utils.Success(c, "报告已保存", nil)
}

// ConfirmAIReport 确认AI报告
// 核心操作：1.更新video_analyses状态 2.标记订单完成 3.创建reports记录桥接
func (ctrl *VideoAnalysisController) ConfirmAIReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	// 1. 查询分析记录
	analysis, err := ctrl.analysisRepo.FindByID(uint(id))
	if err != nil || analysis == nil {
		utils.Error(c, http.StatusNotFound, "分析记录不存在")
		return
	}

	// 校验：AI报告必须已生成
	if analysis.AIReport == "" {
		utils.Error(c, http.StatusBadRequest, "请先生成AI报告")
		return
	}

	// 2. 更新 video_analyses 状态
	updates := map[string]interface{}{
		"ai_report_status": "confirmed",
		"status":           models.AnalysisStatusCompleted,
	}
	err = ctrl.analysisRepo.Update(uint(id), updates)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "确认报告失败")
		return
	}

	// 3. 更新订单状态为已完成
	orderRepo := models.NewOrderRepository(ctrl.db)
	err = orderRepo.UpdateStatus(analysis.OrderID, models.OrderStatusCompleted)
	if err != nil {
		// 订单更新失败不阻断主流程，但记录日志
		// 实际生产环境应接入日志框架
	}

	// 4. 创建/更新 reports 记录（桥接 video_analyses → reports）
	reportRepo := models.NewReportRepository(ctrl.db)
	existingReport, _ := reportRepo.FindByOrderID(analysis.OrderID)

	if existingReport != nil {
		// 更新已有报告
		reportRepo.Update(existingReport.ID, map[string]interface{}{
			"content":        analysis.AIReport,
			"status":         models.ReportStatusCompleted,
			"overall_rating": analysis.OverallScore,
			"potential":      string(analysis.PotentialLevel),
			"summary":        analysis.Summary,
			"suggestions":    analysis.Improvements,
			"rating_details": analysis.Scores, // JSON 格式的完整评分
		})
	} else {
		// 创建新的桥接报告
		report := &models.Report{
			OrderID:         analysis.OrderID,
			UserID:          analysis.UserID,
			AnalystID:       analysis.AnalystID,
			PlayerName:      analysis.PlayerName,
			PlayerPosition:  analysis.PlayerPosition,
			Content:         analysis.AIReport,
			Status:          models.ReportStatusCompleted,
			OverallRating:   analysis.OverallScore,
			Potential:       string(analysis.PotentialLevel),
			Summary:         analysis.Summary,
			Suggestions:     analysis.Improvements,
			RatingDetails:   analysis.Scores,
		}
		reportRepo.Create(report)
		// 同时更新 order 的 report_id
		ctrl.db.Model(&models.Order{}).Where("id = ?", analysis.OrderID).
			Update("report_id", report.ID)
	}

	// 更新 video_analysis 状态为已完成
	ctrl.analysisRepo.Update(analysis.ID, map[string]interface{}{
		"status": models.AnalysisStatusCompleted,
	})

	// 生成 MD 文档
	if ctrl.reportGen != nil {
		var user models.User
		_ = ctrl.db.First(&user, analysis.UserID).Error
		var analyst models.Analyst
		_ = ctrl.db.First(&analyst, analysis.AnalystID).Error
		ratingMD, playerMD, _ := ctrl.reportGen.GenerateFromVideoAnalysis(analysis, analyst.Name, &user)
		if ratingMD != "" {
			ctrl.analysisRepo.Update(analysis.ID, map[string]interface{}{
				"rating_report_md": ratingMD,
				"player_info_md": playerMD,
			})
		}
	}

	utils.Success(c, "报告已确认，文档已生成", gin.H{
		"order_id":    analysis.OrderID,
		"analysis_id": id,
	})
}

// GetAIReport 获取AI报告
func (ctrl *VideoAnalysisController) GetAIReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	analysis, err := ctrl.analysisRepo.FindByID(uint(id))
	if err != nil || analysis == nil {
		utils.Error(c, http.StatusNotFound, "分析记录不存在")
		return
	}

	utils.Success(c, "", gin.H{
		"report":  analysis.AIReport,
		"status":  analysis.AIReportStatus,
		"version": analysis.AIReportVersion,
	})
}

// CreateAnalysisFromOrderRequest 创建分析请求
type CreateAnalysisFromOrderRequest struct {
	OrderID uint `json:"order_id"`
}

// CreateFromOrder 从订单创建分析记录
func (ctrl *VideoAnalysisController) CreateFromOrder(c *gin.Context) {
	var req CreateAnalysisFromOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	existing, _ := ctrl.analysisRepo.FindByOrderID(req.OrderID)
	if existing != nil {
		utils.Error(c, http.StatusBadRequest, "该订单已有分析记录")
		return
	}

	// 查询订单信息，补全 analyst_id 和 user_id
	var order models.Order
	if err := ctrl.db.First(&order, req.OrderID).Error; err != nil {
		utils.Error(c, http.StatusBadRequest, "订单不存在")
		return
	}

	if order.AnalystID == nil {
		utils.Error(c, http.StatusBadRequest, "订单尚未分配分析师")
		return
	}

	analysis := &models.VideoAnalysis{
		OrderID:        req.OrderID,
		AnalystID:      *order.AnalystID,
		UserID:         order.UserID,
		PlayerName:     order.PlayerName,
		PlayerAge:      order.PlayerAge,
		PlayerPosition: order.PlayerPosition,
		MatchName:      order.MatchName,
		Opponent:       order.Opponent,
		Status:         models.AnalysisStatusScoring,
	}

	err := ctrl.analysisRepo.Create(analysis)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建失败")
		return
	}

	utils.Success(c, "分析记录已创建", analysis)
}

// ========== 球员端 API ==========

// GetMyAnalyses 获取当前用户的视频分析列表（球员视角）
func (ctrl *VideoAnalysisController) GetMyAnalyses(c *gin.Context) {
	userID := c.GetUint("userId")

	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	analyses, total, err := ctrl.analysisRepo.FindByUserID(userID, page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	utils.Success(c, "", gin.H{
		"list":      analyses,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetMyAnalysisDetail 获取当前用户的某条视频分析详情（含评分+报告）
func (ctrl *VideoAnalysisController) GetMyAnalysisDetail(c *gin.Context) {
	userID := c.GetUint("userId")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	analysis, err := ctrl.analysisRepo.FindByID(uint(id))
	if err != nil || analysis == nil {
		utils.Error(c, http.StatusNotFound, "分析记录不存在")
		return
	}

	// 权限校验：只能查看自己的分析
	if analysis.UserID != userID {
		utils.Error(c, http.StatusForbidden, "无权查看此分析")
		return
	}

	scores, _ := models.ParseScoresFromJSON(analysis.Scores)
	highlights, _ := ctrl.highlightRepo.FindByAnalysisID(analysis.ID)

	utils.Success(c, "", gin.H{
		"analysis":      analysis,
		"scores":        scores,
		"highlights":    highlights,
		"ai_report":     analysis.AIReport,
		"ai_report_status": analysis.AIReportStatus,
		"ai_report_version": analysis.AIReportVersion,
	})
}
