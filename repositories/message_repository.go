package repositories

import (
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// MessageRepository 私信Repository
type MessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建私信Repository
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// CreateMessage 创建私信
func (r *MessageRepository) CreateMessage(msg *models.Message) error {
	return r.db.Create(msg).Error
}

// GetMessageByID 根据ID获取私信
func (r *MessageRepository) GetMessageByID(id uint) (*models.Message, error) {
	var msg models.Message
	err := r.db.Preload("Sender").Preload("Receiver").First(&msg, id).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetMessagesBetweenUsers 获取两个用户之间的私信列表
func (r *MessageRepository) GetMessagesBetweenUsers(userID1, userID2 uint, limit, offset int) ([]models.Message, int64, error) {
	var messages []models.Message
	var total int64

	query := r.db.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID1, userID2, userID2, userID1,
	)

	err := query.Model(&models.Message{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("Sender").Preload("Receiver").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&messages).Error

	return messages, total, err
}

// GetConversations 获取用户的会话列表
func (r *MessageRepository) GetConversations(userID uint) ([]models.Conversation, error) {
	// 使用子查询获取每个会话的最后一条消息
	type result struct {
		OtherUserID uint
		LastMsgID   uint
	}

	var results []result
	err := r.db.Raw(`
		SELECT 
			CASE WHEN sender_id = ? THEN receiver_id ELSE sender_id END as other_user_id,
			MAX(id) as last_msg_id
		FROM messages
		WHERE sender_id = ? OR receiver_id = ?
		GROUP BY other_user_id
		ORDER BY MAX(created_at) DESC
	`, userID, userID, userID).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	var conversations []models.Conversation
	for _, res := range results {
		var lastMsg models.Message
		err := r.db.Preload("Sender").Preload("Receiver").First(&lastMsg, res.LastMsgID).Error
		if err != nil {
			continue
		}

		var otherUser models.User
		err = r.db.First(&otherUser, res.OtherUserID).Error
		if err != nil {
			continue
		}

		// 计算未读数
		var unreadCount int64
		r.db.Model(&models.Message{}).
			Where("sender_id = ? AND receiver_id = ? AND is_read = ?", res.OtherUserID, userID, false).
			Count(&unreadCount)

		conv := models.Conversation{
			UserID:      otherUser.ID,
			UserName:    otherUser.Nickname,
			UserAvatar:  otherUser.Avatar,
			LastMessage: lastMsg.Content,
			LastTime:    lastMsg.CreatedAt,
			UnreadCount: unreadCount,
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// MarkAsRead 标记消息为已读
func (r *MessageRepository) MarkAsRead(messageID, userID uint) error {
	return r.db.Model(&models.Message{}).
		Where("id = ? AND receiver_id = ?", messageID, userID).
		Update("is_read", true).Error
}

// MarkConversationAsRead 标记与某用户的所有消息为已读
func (r *MessageRepository) MarkConversationAsRead(userID, otherUserID uint) error {
	return r.db.Model(&models.Message{}).
		Where("sender_id = ? AND receiver_id = ? AND is_read = ?", otherUserID, userID, false).
		Update("is_read", true).Error
}

// GetUnreadCount 获取用户未读私信数
func (r *MessageRepository) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Message{}).
		Where("receiver_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// DeleteMessage 删除私信
func (r *MessageRepository) DeleteMessage(id uint) error {
	return r.db.Delete(&models.Message{}, id).Error
}

// GetDailyMessageCount 获取发送者今天向接收者发送的消息数
func (r *MessageRepository) GetDailyMessageCount(senderID, receiverID uint) (int64, error) {
	var count int64
	now := models.GetTime()
	today := now.Format("2006-01-02")
	err := r.db.Model(&models.Message{}).
		Where("sender_id = ? AND receiver_id = ? AND DATE(created_at) = ?", senderID, receiverID, today).
		Count(&count).Error
	return count, err
}
