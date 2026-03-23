package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/websocket"
)

// NotificationService handles notification creation and WebSocket delivery
type NotificationService struct {
	repo       *repository.NotificationRepo
	userRepo   *repository.UserRepo
	hub        *websocket.Hub
	webhookSvc *WebhookService
	log        *slog.Logger
}

// NewNotificationService creates a new notification service
func NewNotificationService(repo *repository.NotificationRepo, userRepo *repository.UserRepo, hub *websocket.Hub, log *slog.Logger) *NotificationService {
	return &NotificationService{
		repo:     repo,
		userRepo: userRepo,
		hub:      hub,
		log:      log,
	}
}

// SetWebhookService wires the webhook dispatcher into the notification service.
func (s *NotificationService) SetWebhookService(svc *WebhookService) {
	s.webhookSvc = svc
}

// CreateAndSend creates a notification in DB and delivers it via WebSocket.
// Webhook dispatch is intentionally NOT done here — callers invoke dispatchWebhook
// once per business event, before any per-user fan-out, to avoid duplicate POSTs.
func (s *NotificationService) CreateAndSend(ctx context.Context, notif *model.Notification) error {
	if err := s.repo.Create(ctx, notif); err != nil {
		return fmt.Errorf("creating notification: %w", err)
	}

	s.hub.Broadcast(notif)

	s.log.Info("notification sent",
		"type", notif.Type,
		"user_id", notif.UserID,
		"title", notif.Title)

	return nil
}

// dispatchWebhook fires the webhook for an event exactly once, in a goroutine.
// Call this before CreateAndSend / broadcastToAll so fan-out doesn't multiply it.
func (s *NotificationService) dispatchWebhook(event string, payload any) {
	if s.webhookSvc != nil {
		go s.webhookSvc.Dispatch(event, payload)
	}
}

// broadcastToAll fans out one notification row per user so each user can independently
// mark it read or delete it. Avoids the shared mutable state of user_id = NULL rows.
func (s *NotificationService) broadcastToAll(ctx context.Context, notif *model.Notification) error {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("listing users for broadcast: %w", err)
	}
	for _, u := range users {
		uid := u.ID
		n := *notif // copy so each row gets its own pointer
		n.ID = 0    // reset so DB assigns a new ID
		n.UserID = &uid
		if err := s.CreateAndSend(ctx, &n); err != nil {
			s.log.Error("failed to send broadcast notification", "user_id", uid, "error", err)
		}
	}
	return nil
}

// NotifyScanComplete sends notification when scan job completes.
// Pass nil userID to broadcast to all users.
func (s *NotificationService) NotifyScanComplete(ctx context.Context, userID *int64, libraryID int64, libraryName string, scannedCount, newCount, errorCount int) error {
	data := model.NotificationData{
		LibraryID:    &libraryID,
		ScannedCount: scannedCount,
		NewCount:     newCount,
		ErrorCount:   errorCount,
	}

	notif := &model.Notification{
		Type:    model.NotificationScanComplete,
		Title:   "Scan Complete",
		Message: fmt.Sprintf("Scanned %d files in %s. Found %d new.", scannedCount, libraryName, newCount),
		Data:    data.ToJSON(),
		Read:    false,
	}

	s.dispatchWebhook(string(notif.Type), notif)

	if userID != nil {
		notif.UserID = userID
		return s.CreateAndSend(ctx, notif)
	}
	return s.broadcastToAll(ctx, notif)
}

// NotifyMediaAdded sends notification when new media is detected.
// Pass nil userID to broadcast to all users.
func (s *NotificationService) NotifyMediaAdded(ctx context.Context, userID *int64, mediaID int64, title, mediaType string) error {
	data := model.NotificationData{
		MediaID:    &mediaID,
		MediaTitle: title,
		MediaType:  mediaType,
	}

	notif := &model.Notification{
		Type:    model.NotificationMediaAdded,
		Title:   "New Media Added",
		Message: fmt.Sprintf("%s: %s", mediaType, title),
		Data:    data.ToJSON(),
		Read:    false,
	}

	s.dispatchWebhook(string(notif.Type), notif)

	if userID != nil {
		notif.UserID = userID
		return s.CreateAndSend(ctx, notif)
	}
	return s.broadcastToAll(ctx, notif)
}

// NotifyTranscodeComplete sends notification when transcode finishes
func (s *NotificationService) NotifyTranscodeComplete(ctx context.Context, userID int64, mediaID int64, mediaTitle string, success bool, quality string, duration int) error {
	data := model.NotificationData{
		MediaID:  &mediaID,
		Quality:  quality,
		Duration: duration,
	}

	var notifType model.NotificationType
	var title, message string

	if success {
		notifType = model.NotificationTranscodeComplete
		title = "Transcode Complete"
		message = fmt.Sprintf("%s is ready for streaming", mediaTitle)
	} else {
		notifType = model.NotificationTranscodeFailed
		title = "Transcode Failed"
		message = fmt.Sprintf("Failed to transcode %s", mediaTitle)
	}

	notif := &model.Notification{
		UserID:  &userID,
		Type:    notifType,
		Title:   title,
		Message: message,
		Data:    data.ToJSON(),
		Read:    false,
	}

	s.dispatchWebhook(string(notif.Type), notif)
	return s.CreateAndSend(ctx, notif)
}

// NotifySubtitleDownloaded sends notification when subtitle download completes.
// Pass nil userID to broadcast to all users (e.g. from the scan pipeline).
func (s *NotificationService) NotifySubtitleDownloaded(ctx context.Context, userID *int64, mediaID int64, mediaTitle, language, provider string) error {
	data := model.NotificationData{
		MediaID:  &mediaID,
		Language: language,
		Provider: provider,
	}

	notif := &model.Notification{
		UserID:  userID,
		Type:    model.NotificationSubtitleDownloaded,
		Title:   "Subtitle Downloaded",
		Message: fmt.Sprintf("Downloaded %s subtitle for %s from %s", language, mediaTitle, provider),
		Data:    data.ToJSON(),
		Read:    false,
	}

	s.dispatchWebhook(string(notif.Type), notif)
	if userID != nil {
		return s.CreateAndSend(ctx, notif)
	}
	return s.broadcastToAll(ctx, notif)
}

// NotifyIdentifyComplete sends notification when media identification finishes.
// Pass nil userID to broadcast to all users.
func (s *NotificationService) NotifyIdentifyComplete(ctx context.Context, userID *int64, mediaID int64, title string, success bool) error {
	data := model.NotificationData{
		MediaID:    &mediaID,
		MediaTitle: title,
	}

	var message string
	if success {
		message = fmt.Sprintf("Successfully identified: %s", title)
	} else {
		message = fmt.Sprintf("Could not identify: %s", title)
	}

	notif := &model.Notification{
		Type:    model.NotificationIdentifyComplete,
		Title:   "Identification Complete",
		Message: message,
		Data:    data.ToJSON(),
		Read:    false,
	}

	s.dispatchWebhook(string(notif.Type), notif)
	if userID != nil {
		notif.UserID = userID
		return s.CreateAndSend(ctx, notif)
	}
	return s.broadcastToAll(ctx, notif)
}

// NotifyLibraryWatcher sends notification when watcher detects new files (broadcast to all users)
func (s *NotificationService) NotifyLibraryWatcher(ctx context.Context, libraryID int64, libraryName string, fileCount int) error {
	data := model.NotificationData{
		LibraryID: &libraryID,
	}

	notif := &model.Notification{
		Type:    model.NotificationLibraryWatcher,
		Title:   "New Files Detected",
		Message: fmt.Sprintf("Detected %d new files in %s", fileCount, libraryName),
		Data:    data.ToJSON(),
		Read:    false,
	}

	s.dispatchWebhook(string(notif.Type), notif)
	return s.broadcastToAll(ctx, notif)
}

// GetByUser retrieves notifications for a user
func (s *NotificationService) GetByUser(ctx context.Context, userID int64, unreadOnly bool, limit, offset int) ([]model.Notification, error) {
	filter := model.NotificationFilter{
		UserID:     &userID,
		UnreadOnly: unreadOnly,
		Limit:      limit,
		Offset:     offset,
	}
	return s.repo.GetByUser(ctx, filter)
}

// MarkAsRead marks notifications as read
func (s *NotificationService) MarkAsRead(ctx context.Context, userID int64, ids []int64) error {
	return s.repo.MarkAsRead(ctx, userID, ids)
}

// MarkAllAsRead marks all notifications as read
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID int64) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

// Delete removes notifications
func (s *NotificationService) Delete(ctx context.Context, userID int64, ids []int64) error {
	return s.repo.Delete(ctx, userID, ids)
}

// CountUnread returns unread notification count
func (s *NotificationService) CountUnread(ctx context.Context, userID int64) (int64, error) {
	return s.repo.CountUnread(ctx, userID)
}
