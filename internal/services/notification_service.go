package services

import (
	"context"
	"fmt"

	"github.com/tunecent/backend/internal/database"
	"github.com/tunecent/backend/internal/models"
)

type NotificationService struct {
	db *database.DB
}

func NewNotificationService(db *database.DB) *NotificationService {
	return &NotificationService{db: db}
}

type CreateNotificationRequest struct {
	UserAddress string `json:"user_address" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Message     string `json:"message" binding:"required"`
	RelatedID   uint64 `json:"related_id"`
	TxHash      string `json:"tx_hash"`
}

func (s *NotificationService) CreateNotification(ctx context.Context, req *CreateNotificationRequest) (*models.Notification, error) {
	notification := &models.Notification{
		UserAddress: req.UserAddress,
		Type:        req.Type,
		Title:       req.Title,
		Message:     req.Message,
		IsRead:      false,
		RelatedID:   req.RelatedID,
		TxHash:      req.TxHash,
	}

	if err := s.db.Create(notification).Error; err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return notification, nil
}

func (s *NotificationService) GetNotifications(ctx context.Context, userAddress string, limit, offset int, unreadOnly bool) ([]*models.Notification, int64, error) {
	var notifications []*models.Notification
	var total int64

	query := s.db.Model(&models.Notification{}).Where("user_address = ?", userAddress)

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	query.Count(&total)
	query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&notifications)

	return notifications, total, nil
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userAddress string) (int64, error) {
	var count int64
	err := s.db.Model(&models.Notification{}).
		Where("user_address = ? AND is_read = ?", userAddress, false).
		Count(&count).Error

	return count, err
}

func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID uint, userAddress string) error {
	result := s.db.Model(&models.Notification{}).
		Where("id = ? AND user_address = ?", notificationID, userAddress).
		Update("is_read", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, userAddress string) error {
	return s.db.Model(&models.Notification{}).
		Where("user_address = ? AND is_read = ?", userAddress, false).
		Update("is_read", true).Error
}

func (s *NotificationService) DeleteNotification(ctx context.Context, notificationID uint, userAddress string) error {
	result := s.db.Where("id = ? AND user_address = ?", notificationID, userAddress).
		Delete(&models.Notification{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

func (s *NotificationService) GetPreferences(ctx context.Context, userAddress string) (*models.NotificationPreference, error) {
	var prefs models.NotificationPreference
	err := s.db.Where("user_address = ?", userAddress).First(&prefs).Error

	if err != nil {
		// Create default preferences if not exists
		prefs = models.NotificationPreference{
			UserAddress:        userAddress,
			EmailNotifications: true,
			RoyaltyAlerts:      true,
			ContributionAlerts: true,
			MilestoneAlerts:    true,
			MarketingEmails:    false,
		}
		s.db.Create(&prefs)
	}

	return &prefs, nil
}

func (s *NotificationService) UpdatePreferences(ctx context.Context, userAddress string, prefs map[string]bool) error {
	var existing models.NotificationPreference
	err := s.db.Where("user_address = ?", userAddress).First(&existing).Error

	if err != nil {
		// Create if not exists
		existing = models.NotificationPreference{
			UserAddress: userAddress,
		}
	}

	// Update preferences
	if val, ok := prefs["email_notifications"]; ok {
		existing.EmailNotifications = val
	}
	if val, ok := prefs["royalty_alerts"]; ok {
		existing.RoyaltyAlerts = val
	}
	if val, ok := prefs["contribution_alerts"]; ok {
		existing.ContributionAlerts = val
	}
	if val, ok := prefs["milestone_alerts"]; ok {
		existing.MilestoneAlerts = val
	}
	if val, ok := prefs["marketing_emails"]; ok {
		existing.MarketingEmails = val
	}

	return s.db.Save(&existing).Error
}

// Helper function to create common notification types
func (s *NotificationService) NotifyRoyaltyReceived(ctx context.Context, userAddress string, tokenID uint64, amount string, txHash string) error {
	req := &CreateNotificationRequest{
		UserAddress: userAddress,
		Type:        "payment",
		Title:       "Royalty Payment Received",
		Message:     fmt.Sprintf("You received a royalty payment of %s wei", amount),
		RelatedID:   tokenID,
		TxHash:      txHash,
	}
	_, err := s.CreateNotification(ctx, req)
	return err
}

func (s *NotificationService) NotifyContributionConfirmed(ctx context.Context, userAddress string, campaignID uint64, amount string, txHash string) error {
	req := &CreateNotificationRequest{
		UserAddress: userAddress,
		Type:        "contribution",
		Title:       "Contribution Confirmed",
		Message:     fmt.Sprintf("Your contribution of %s wei has been confirmed", amount),
		RelatedID:   campaignID,
		TxHash:      txHash,
	}
	_, err := s.CreateNotification(ctx, req)
	return err
}

func (s *NotificationService) NotifyMilestoneReached(ctx context.Context, userAddress string, campaignID uint64, milestone string) error {
	req := &CreateNotificationRequest{
		UserAddress: userAddress,
		Type:        "milestone",
		Title:       "Milestone Reached",
		Message:     fmt.Sprintf("Campaign milestone reached: %s", milestone),
		RelatedID:   campaignID,
	}
	_, err := s.CreateNotification(ctx, req)
	return err
}
