package services

import (
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
)

// MessageService 私信服务
type MessageService struct {
	repo                *repositories.MessageRepository
	userRepo            *models.UserRepository
	socialRepo          *repositories.SocialRepository
	notificationService *NotificationService
}

// NewMessageService 创建私信服务
func NewMessageService(
	repo *repositories.MessageRepository,
	userRepo *models.UserRepository,
	socialRepo *repositories.SocialRepository,
	notificationService *NotificationService,
) *MessageService {
	return &MessageService{
		repo:                repo,
		userRepo:            userRepo,
		socialRepo:          socialRepo,
		notificationService: notificationService,
	}
}

// SendMessage 发送私信（未互相关注用户每天限发1条）
func (s *MessageService) SendMessage(senderID, receiverID uint, content string) (*models.Message, error) {
	// 检查是否互相关注
	isMutualFollow := false
	if s.socialRepo != nil {
		isMutualFollow, _ = s.socialRepo.IsMutualFollow(senderID, receiverID)
	}

	// 未互相关注则限制每天1条
	if !isMutualFollow {
		count, err := s.repo.GetDailyMessageCount(senderID, receiverID)
		if err != nil {
			return nil, err
		}
		if count >= 1 {
			return nil, &MessageLimitError{Message: "未互相关注的用户每天只能发送一条私信，请先关注对方"}
		}
	}

	msg := &models.Message{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
	}

	err := s.repo.CreateMessage(msg)
	if err != nil {
		return nil, err
	}

	// 发送通知
	go s.sendMessageNotification(senderID, receiverID, content)

	// 重新加载关联
	return s.repo.GetMessageByID(msg.ID)
}

// MessageLimitError 私信限制错误
type MessageLimitError struct {
	Message string
}

func (e *MessageLimitError) Error() string {
	return e.Message
}

// sendMessageNotification 发送新私信通知
func (s *MessageService) sendMessageNotification(senderID, receiverID uint, content string) {
	if senderID == receiverID {
		return
	}

	triggerUser, _ := s.userRepo.FindByID(senderID)
	triggerName := "用户"
	triggerAvatar := ""
	if triggerUser != nil {
		triggerName = triggerUser.Nickname
		triggerAvatar = triggerUser.Avatar
	}

	// 截取预览
	preview := content
	if len([]rune(content)) > 30 {
		preview = string([]rune(content)[:30]) + "..."
	}

	data := &models.NotificationData{
		TriggerUserID:   senderID,
		TriggerUserName: triggerName,
		TriggerAvatar:   triggerAvatar,
	}

	if s.notificationService != nil {
		s.notificationService.CreateNotification(receiverID, models.NotificationTypeMessage, "新私信", triggerName+" 发来一条私信: "+preview, data)
	}
}

// GetMessages 获取两个用户之间的私信
func (s *MessageService) GetMessages(userID1, userID2 uint, limit, offset int) ([]models.Message, int64, error) {
	return s.repo.GetMessagesBetweenUsers(userID1, userID2, limit, offset)
}

// GetConversations 获取会话列表
func (s *MessageService) GetConversations(userID uint) ([]models.Conversation, error) {
	return s.repo.GetConversations(userID)
}

// MarkAsRead 标记已读
func (s *MessageService) MarkAsRead(messageID, userID uint) error {
	return s.repo.MarkAsRead(messageID, userID)
}

// MarkConversationAsRead 标记会话已读
func (s *MessageService) MarkConversationAsRead(userID, otherUserID uint) error {
	return s.repo.MarkConversationAsRead(userID, otherUserID)
}

// GetUnreadCount 获取未读数
func (s *MessageService) GetUnreadCount(userID uint) (int64, error) {
	return s.repo.GetUnreadCount(userID)
}

// DeleteMessage 删除私信
func (s *MessageService) DeleteMessage(id uint) error {
	return s.repo.DeleteMessage(id)
}
