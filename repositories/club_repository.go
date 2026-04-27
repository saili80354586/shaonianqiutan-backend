package repositories

import (
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// ClubRepository 俱乐部仓储
type ClubRepository struct {
	db *gorm.DB
}

// NewClubRepository 创建俱乐部仓储
func NewClubRepository(db *gorm.DB) *ClubRepository {
	return &ClubRepository{db: db}
}

// GetByUserID 根据用户ID获取俱乐部
func (r *ClubRepository) GetByUserID(userID uint) (*models.Club, error) {
	var club models.Club
	err := r.db.Where("user_id = ?", userID).First(&club).Error
	if err != nil {
		return nil, err
	}
	return &club, nil
}

// CreateClubInvitation 创建俱乐部邀请记录
func (r *ClubRepository) CreateClubInvitation(inv *models.ClubInvitation) error {
	return r.db.Create(inv).Error
}

// FindClubInvitationByCode 根据邀请码查找俱乐部邀请
func (r *ClubRepository) FindClubInvitationByCode(code string) (*models.ClubInvitation, error) {
	var inv models.ClubInvitation
	err := r.db.Preload("Club").Preload("TargetUser").Preload("Creator").
		Where("invite_code = ?", code).
		First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// UpdateClubInvitationStatus 更新俱乐部邀请状态
func (r *ClubRepository) UpdateClubInvitationStatus(id uint, status models.InvitationStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == models.InvitationStatusAccepted {
		updates["accepted_at"] = time.Now()
	} else if status == models.InvitationStatusRejected {
		updates["rejected_at"] = time.Now()
	}
	return r.db.Model(&models.ClubInvitation{}).Where("id = ?", id).Updates(updates).Error
}

// GetClubInvitations 获取俱乐部的邀请列表
func (r *ClubRepository) GetClubInvitations(clubID uint, status string) ([]models.ClubInvitation, error) {
	var invitations []models.ClubInvitation
	query := r.db.Where("club_id = ?", clubID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Preload("TargetUser").Order("created_at DESC").Find(&invitations).Error
	return invitations, err
}
