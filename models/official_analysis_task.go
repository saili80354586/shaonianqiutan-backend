package models

import "time"

type OfficialAnalysisTaskStatus string

const (
	OfficialAnalysisTaskDraft     OfficialAnalysisTaskStatus = "draft"
	OfficialAnalysisTaskPublished OfficialAnalysisTaskStatus = "published"
	OfficialAnalysisTaskFull      OfficialAnalysisTaskStatus = "full"
	OfficialAnalysisTaskClosed    OfficialAnalysisTaskStatus = "closed"
	OfficialAnalysisTaskExpired   OfficialAnalysisTaskStatus = "expired"
)

type OfficialAnalysisAcceptanceStatus string

const (
	OfficialAnalysisAcceptanceAccepted  OfficialAnalysisAcceptanceStatus = "accepted"
	OfficialAnalysisAcceptanceSubmitted OfficialAnalysisAcceptanceStatus = "submitted"
	OfficialAnalysisAcceptanceCancelled OfficialAnalysisAcceptanceStatus = "cancelled"
	OfficialAnalysisAcceptanceExpired   OfficialAnalysisAcceptanceStatus = "expired"
)

type OfficialAnalysisSubmissionStatus string

const (
	OfficialAnalysisSubmissionDraft            OfficialAnalysisSubmissionStatus = "draft"
	OfficialAnalysisSubmissionSubmitted        OfficialAnalysisSubmissionStatus = "submitted"
	OfficialAnalysisSubmissionApproved         OfficialAnalysisSubmissionStatus = "approved"
	OfficialAnalysisSubmissionRevisionRequired OfficialAnalysisSubmissionStatus = "revision_required"
	OfficialAnalysisSubmissionRejected         OfficialAnalysisSubmissionStatus = "rejected"
	OfficialAnalysisSubmissionAdopted          OfficialAnalysisSubmissionStatus = "adopted"
)

type OfficialContentAdoptionStatus string

const (
	OfficialContentAdoptionMaterial          OfficialContentAdoptionStatus = "material"
	OfficialContentAdoptionOfficialPublished OfficialContentAdoptionStatus = "official_published"
	OfficialContentAdoptionKeySpread         OfficialContentAdoptionStatus = "key_spread"
	OfficialContentAdoptionLongTerm          OfficialContentAdoptionStatus = "long_term"
)

type AnalystRewardStatus string

const (
	AnalystRewardPending   AnalystRewardStatus = "pending"
	AnalystRewardSettled   AnalystRewardStatus = "settled"
	AnalystRewardCancelled AnalystRewardStatus = "cancelled"
	AnalystRewardReversed  AnalystRewardStatus = "reversed"
)

// OfficialAnalysisTask 官方选题单，独立于付费订单。
type OfficialAnalysisTask struct {
	ID                   uint                       `json:"id" gorm:"primaryKey"`
	TaskNo               string                     `json:"task_no" gorm:"uniqueIndex;size:40;not null"`
	Title                string                     `json:"title" gorm:"size:120;not null"`
	MatchName            string                     `json:"match_name" gorm:"size:120"`
	AgeGroup             string                     `json:"age_group" gorm:"size:50"`
	MatchDate            string                     `json:"match_date" gorm:"size:20"`
	VideoURL             string                     `json:"video_url" gorm:"size:500"`
	VideoFirstHalfURL    string                     `json:"video_first_half_url" gorm:"size:500"`
	VideoSecondHalfURL   string                     `json:"video_second_half_url" gorm:"size:500"`
	VideoSource          string                     `json:"video_source" gorm:"size:100"`
	AuthorizationStatus  string                     `json:"authorization_status" gorm:"size:30;not null;default:'unknown'"`
	TargetPlayerUserID   uint                       `json:"target_player_user_id" gorm:"default:0;index"`
	TargetPlayerName     string                     `json:"target_player_name" gorm:"size:80"`
	TargetPlayerTeam     string                     `json:"target_player_team" gorm:"size:100"`
	TargetJerseyColor    string                     `json:"target_jersey_color" gorm:"size:30"`
	TargetJerseyNumber   string                     `json:"target_jersey_number" gorm:"size:20"`
	TargetPlayerPosition string                     `json:"target_player_position" gorm:"size:50"`
	TaskType             string                     `json:"task_type" gorm:"size:30;not null;default:'composite'"`
	BaseRewardAmount     float64                    `json:"base_reward_amount" gorm:"type:decimal(10,2);default:0"`
	AdoptionRewardRule   string                     `json:"adoption_reward_rule" gorm:"type:text"`
	BonusRule            string                     `json:"bonus_rule" gorm:"type:text"`
	Requirements         string                     `json:"requirements" gorm:"type:text"`
	MaxAcceptCount       int                        `json:"max_accept_count" gorm:"not null;default:1"`
	CurrentAcceptCount   int                        `json:"current_accept_count" gorm:"not null;default:0"`
	VisibleLevelMin      string                     `json:"visible_level_min" gorm:"size:20;not null;default:'L1';index"`
	VisibleLevelCodes    string                     `json:"visible_level_codes" gorm:"size:80;index"`
	PriorityLevelMin     string                     `json:"priority_level_min" gorm:"size:20"`
	PriorityUntil        *time.Time                 `json:"priority_until"`
	Deadline             *time.Time                 `json:"deadline" gorm:"index"`
	Status               OfficialAnalysisTaskStatus `json:"status" gorm:"size:20;not null;default:'draft';index"`
	CreatedBy            uint                       `json:"created_by" gorm:"index"`
	ClosedBy             uint                       `json:"closed_by" gorm:"default:0"`
	ClosedAt             *time.Time                 `json:"closed_at"`
	CreatedAt            time.Time                  `json:"created_at"`
	UpdatedAt            time.Time                  `json:"updated_at"`
}

// OfficialAnalysisTaskAcceptance 官方选题接单记录。
type OfficialAnalysisTaskAcceptance struct {
	ID              uint                             `json:"id" gorm:"primaryKey"`
	TaskID          uint                             `json:"task_id" gorm:"not null;index;uniqueIndex:idx_official_acceptance_task_analyst"`
	Task            *OfficialAnalysisTask            `json:"task,omitempty" gorm:"foreignKey:TaskID"`
	AnalystID       uint                             `json:"analyst_id" gorm:"not null;index;uniqueIndex:idx_official_acceptance_task_analyst"`
	Analyst         *Analyst                         `json:"analyst,omitempty" gorm:"foreignKey:AnalystID"`
	AnalysisOrderID uint                             `json:"analysis_order_id" gorm:"default:0;index"`
	Submissions     []OfficialAnalysisSubmission     `json:"submissions,omitempty" gorm:"foreignKey:AcceptanceID"`
	AcceptedAt      time.Time                        `json:"accepted_at" gorm:"not null;index"`
	SubmittedAt     *time.Time                       `json:"submitted_at"`
	Status          OfficialAnalysisAcceptanceStatus `json:"status" gorm:"size:20;not null;default:'accepted';index"`
	CancelReason    string                           `json:"cancel_reason" gorm:"size:500"`
	CreatedAt       time.Time                        `json:"created_at"`
	UpdatedAt       time.Time                        `json:"updated_at"`
}

// OfficialAnalysisSubmission 分析师提交的官方任务版本。
type OfficialAnalysisSubmission struct {
	ID                       uint                             `json:"id" gorm:"primaryKey"`
	TaskID                   uint                             `json:"task_id" gorm:"not null;index"`
	Task                     *OfficialAnalysisTask            `json:"task,omitempty" gorm:"foreignKey:TaskID"`
	AnalystID                uint                             `json:"analyst_id" gorm:"not null;index"`
	Analyst                  *Analyst                         `json:"analyst,omitempty" gorm:"foreignKey:AnalystID"`
	AcceptanceID             uint                             `json:"acceptance_id" gorm:"not null;index"`
	Acceptance               *OfficialAnalysisTaskAcceptance  `json:"acceptance,omitempty" gorm:"foreignKey:AcceptanceID"`
	ReportID                 uint                             `json:"report_id" gorm:"default:0;index"`
	AnalysisID               uint                             `json:"analysis_id" gorm:"default:0;index"`
	VideoFileURL             string                           `json:"video_file_url" gorm:"size:500"`
	VideoAuthorizationStatus string                           `json:"video_authorization_status" gorm:"size:30;not null;default:'unknown'"`
	ScriptText               string                           `json:"script_text" gorm:"type:text"`
	Summary                  string                           `json:"summary" gorm:"type:text"`
	SubmitNote               string                           `json:"submit_note" gorm:"type:text"`
	Status                   OfficialAnalysisSubmissionStatus `json:"status" gorm:"size:30;not null;default:'draft';index"`
	ReviewedBy               uint                             `json:"reviewed_by" gorm:"default:0;index"`
	ReviewedAt               *time.Time                       `json:"reviewed_at"`
	ReviewNote               string                           `json:"review_note" gorm:"size:500"`
	IsPublic                 bool                             `json:"is_public" gorm:"default:false;index"`
	CreatedAt                time.Time                        `json:"created_at"`
	UpdatedAt                time.Time                        `json:"updated_at"`
}

// OfficialContentAdoption 官方采用记录。
type OfficialContentAdoption struct {
	ID             uint                           `json:"id" gorm:"primaryKey"`
	TaskID         uint                           `json:"task_id" gorm:"not null;index"`
	Task           *OfficialAnalysisTask          `json:"task,omitempty" gorm:"foreignKey:TaskID"`
	SubmissionID   uint                           `json:"submission_id" gorm:"not null;index"`
	Submission     *OfficialAnalysisSubmission    `json:"submission,omitempty" gorm:"foreignKey:SubmissionID"`
	AnalystID      uint                           `json:"analyst_id" gorm:"not null;index"`
	Analyst        *Analyst                       `json:"analyst,omitempty" gorm:"foreignKey:AnalystID"`
	AdoptionStatus OfficialContentAdoptionStatus  `json:"adoption_status" gorm:"size:30;not null;index"`
	Channel        string                         `json:"channel" gorm:"size:50"`
	WorkTitle      string                         `json:"work_title" gorm:"size:120"`
	WorkSummary    string                         `json:"work_summary" gorm:"type:text"`
	CoverURL       string                         `json:"cover_url" gorm:"size:500"`
	BonusBasis     string                         `json:"bonus_basis" gorm:"size:500"`
	RewardAmount   float64                        `json:"reward_amount" gorm:"type:decimal(10,2);default:0"`
	AdoptionNote   string                         `json:"adoption_note" gorm:"size:500"`
	IsPublic       bool                           `json:"is_public" gorm:"default:false;index"`
	PublishRecords []OfficialContentPublishRecord `json:"publish_records,omitempty" gorm:"foreignKey:AdoptionID"`
	CreatedBy      uint                           `json:"created_by" gorm:"index"`
	CreatedAt      time.Time                      `json:"created_at"`
	UpdatedAt      time.Time                      `json:"updated_at"`
}

// OfficialContentPublishRecord 官方采用素材的发布/复用记录。
type OfficialContentPublishRecord struct {
	ID            uint                     `json:"id" gorm:"primaryKey"`
	AdoptionID    uint                     `json:"adoption_id" gorm:"not null;index"`
	Adoption      *OfficialContentAdoption `json:"adoption,omitempty" gorm:"foreignKey:AdoptionID"`
	Channel       string                   `json:"channel" gorm:"size:50;not null;index"`
	AccountName   string                   `json:"account_name" gorm:"size:120"`
	PublishURL    string                   `json:"publish_url" gorm:"size:500"`
	PublishTitle  string                   `json:"publish_title" gorm:"size:160"`
	ReusePurpose  string                   `json:"reuse_purpose" gorm:"size:120"`
	PublishedAt   *time.Time               `json:"published_at" gorm:"index"`
	PlayCount     int64                    `json:"play_count" gorm:"default:0"`
	LikeCount     int64                    `json:"like_count" gorm:"default:0"`
	CommentCount  int64                    `json:"comment_count" gorm:"default:0"`
	ShareCount    int64                    `json:"share_count" gorm:"default:0"`
	FavoriteCount int64                    `json:"favorite_count" gorm:"default:0"`
	MetricsAt     *time.Time               `json:"metrics_recorded_at" gorm:"column:metrics_recorded_at"`
	Note          string                   `json:"note" gorm:"size:500"`
	CreatedBy     uint                     `json:"created_by" gorm:"index"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
}

// OfficialEventTopicConfig 官方赛事专题运营配置。
type OfficialEventTopicConfig struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	MatchName        string    `json:"match_name" gorm:"size:120;not null;uniqueIndex"`
	DisplayName      string    `json:"display_name" gorm:"size:120"`
	Summary          string    `json:"summary" gorm:"type:text"`
	CoverURL         string    `json:"cover_url" gorm:"size:500"`
	AliasNames       []string  `json:"alias_names" gorm:"serializer:json"`
	PinnedAdoptionID uint      `json:"pinned_adoption_id" gorm:"default:0;index"`
	IsFeatured       bool      `json:"is_featured" gorm:"default:false;index"`
	SortOrder        int       `json:"sort_order" gorm:"default:0;index"`
	CreatedBy        uint      `json:"created_by" gorm:"index"`
	UpdatedBy        uint      `json:"updated_by" gorm:"index"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// AnalystRewardRecord 分析师官方奖励记录。
type AnalystRewardRecord struct {
	ID                uint                          `json:"id" gorm:"primaryKey"`
	AnalystID         uint                          `json:"analyst_id" gorm:"not null;index"`
	Analyst           *Analyst                      `json:"analyst,omitempty" gorm:"foreignKey:AnalystID"`
	SourceType        string                        `json:"source_type" gorm:"size:30;not null;index:idx_analyst_reward_source"`
	SourceID          uint                          `json:"source_id" gorm:"not null;index:idx_analyst_reward_source"`
	RewardType        string                        `json:"reward_type" gorm:"size:30;not null"`
	Amount            float64                       `json:"amount" gorm:"type:decimal(10,2);default:0"`
	Status            AnalystRewardStatus           `json:"status" gorm:"size:20;not null;default:'pending';index"`
	SettlementBatchID uint                          `json:"settlement_batch_id" gorm:"default:0;index"`
	SettlementBatch   *AnalystRewardSettlementBatch `json:"settlement_batch,omitempty" gorm:"foreignKey:SettlementBatchID"`
	SettledAt         *time.Time                    `json:"settled_at"`
	SettledBy         uint                          `json:"settled_by" gorm:"default:0"`
	Note              string                        `json:"note" gorm:"size:500"`
	CreatedAt         time.Time                     `json:"created_at"`
	UpdatedAt         time.Time                     `json:"updated_at"`
}

// AnalystRewardSettlementBatch 官方奖励结算批次。
type AnalystRewardSettlementBatch struct {
	ID          uint                  `json:"id" gorm:"primaryKey"`
	BatchNo     string                `json:"batch_no" gorm:"size:40;uniqueIndex;not null"`
	RewardCount int                   `json:"reward_count" gorm:"not null;default:0"`
	TotalAmount float64               `json:"total_amount" gorm:"type:decimal(10,2);default:0"`
	Status      AnalystRewardStatus   `json:"status" gorm:"size:20;not null;default:'settled';index"`
	SettledAt   time.Time             `json:"settled_at" gorm:"index"`
	SettledBy   uint                  `json:"settled_by" gorm:"not null;index"`
	Note        string                `json:"note" gorm:"size:500"`
	Rewards     []AnalystRewardRecord `json:"rewards,omitempty" gorm:"foreignKey:SettlementBatchID"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}
