package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tunecent/backend/internal/services"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
}

func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetNotifications handles GET /api/v1/notifications
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userAddress := c.Query("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	unreadOnlyStr := c.DefaultQuery("unread_only", "false")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	unreadOnly := unreadOnlyStr == "true"

	if limit > 100 {
		limit = 100
	}

	notifications, total, err := h.notificationService.GetNotifications(c.Request.Context(), userAddress, limit, offset, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   notifications,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetUnreadCount handles GET /api/v1/notifications/unread/count
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userAddress := c.Query("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address is required"})
		return
	}

	count, err := h.notificationService.GetUnreadCount(c.Request.Context(), userAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_address":  userAddress,
		"unread_count":  count,
	})
}

// MarkAsRead handles PUT /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	userAddress := c.Query("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address is required"})
		return
	}

	err = h.notificationService.MarkAsRead(c.Request.Context(), uint(notificationID), userAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
	})
}

// MarkAllAsRead handles PUT /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userAddress := c.Query("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address is required"})
		return
	}

	err := h.notificationService.MarkAllAsRead(c.Request.Context(), userAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All notifications marked as read",
	})
}

// DeleteNotification handles DELETE /api/v1/notifications/:id
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	userAddress := c.Query("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address is required"})
		return
	}

	err = h.notificationService.DeleteNotification(c.Request.Context(), uint(notificationID), userAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification deleted",
	})
}

// GetPreferences handles GET /api/v1/notifications/preferences
func (h *NotificationHandler) GetPreferences(c *gin.Context) {
	userAddress := c.Query("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address is required"})
		return
	}

	prefs, err := h.notificationService.GetPreferences(c.Request.Context(), userAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, prefs)
}

// UpdatePreferences handles PUT /api/v1/notifications/preferences
func (h *NotificationHandler) UpdatePreferences(c *gin.Context) {
	userAddress := c.Query("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address is required"})
		return
	}

	var req map[string]bool
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.notificationService.UpdatePreferences(c.Request.Context(), userAddress, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Preferences updated successfully",
	})
}
