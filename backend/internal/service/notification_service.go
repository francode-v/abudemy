package service

import (
	"context"

	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/notify"
	"github.com/abudemy/backend/internal/validate"
)

// NotificationService manages push device registration and the in-app
// notification feed.
type NotificationService struct{ d Deps }

// DeviceInput is the payload for registering a push device.
type DeviceInput struct {
	Platform   models.DevicePlatform
	Token      string
	DeviceName string
	AppVersion string
}

// RegisterDevice registers or refreshes a push device for the caller.
func (s *NotificationService) RegisterDevice(ctx context.Context, p Principal, in DeviceInput) (*models.Device, error) {
	var v validate.Errors
	v.Required("token", in.Token)
	v.MaxLen("token", in.Token, 512)
	v.MaxLen("device_name", in.DeviceName, 120)
	v.MaxLen("app_version", in.AppVersion, 40)
	if in.Platform != "" && !in.Platform.Valid() {
		v.Add("platform", "must be one of: ios, android, web")
	}
	if err := v.Err(); err != nil {
		return nil, err
	}
	platform := in.Platform
	if platform == "" {
		platform = models.DevicePlatformAndroid
	}
	d, err := s.d.Store.UpsertDevice(ctx, &models.Device{
		UserID: p.UserID, Platform: platform, Token: in.Token,
		DeviceName: in.DeviceName, AppVersion: in.AppVersion,
	})
	if err != nil {
		return nil, mapErr(err, "device")
	}
	return d, nil
}

// ListDevices returns the caller's active devices.
func (s *NotificationService) ListDevices(ctx context.Context, p Principal) ([]models.Device, error) {
	devices, err := s.d.Store.ListDevices(ctx, p.UserID)
	if err != nil {
		return nil, mapErr(err, "device")
	}
	return devices, nil
}

// DeleteDevice revokes one of the caller's devices.
func (s *NotificationService) DeleteDevice(ctx context.Context, p Principal, id string) error {
	return mapErr(s.d.Store.RevokeDevice(ctx, id, p.UserID), "device")
}

// List returns a page of the caller's notifications.
func (s *NotificationService) List(ctx context.Context, p Principal, page models.PageParams) (models.Page[models.Notification], error) {
	items, total, err := s.d.Store.ListUserNotifications(ctx, p.UserID, page)
	if err != nil {
		return models.Page[models.Notification]{}, mapErr(err, "notification")
	}
	return models.NewPage(items, total, page), nil
}

// UnreadCount returns the number of unread notifications.
func (s *NotificationService) UnreadCount(ctx context.Context, p Principal) (int, error) {
	n, err := s.d.Store.CountUnreadNotifications(ctx, p.UserID)
	if err != nil {
		return 0, mapErr(err, "notification")
	}
	return n, nil
}

// MarkRead marks one of the caller's notifications as read.
func (s *NotificationService) MarkRead(ctx context.Context, p Principal, id string) error {
	return mapErr(s.d.Store.MarkNotificationRead(ctx, id, p.UserID), "notification")
}

// MarkAllRead marks every unread notification as read.
func (s *NotificationService) MarkAllRead(ctx context.Context, p Principal) (int, error) {
	n, err := s.d.Store.MarkAllNotificationsRead(ctx, p.UserID)
	if err != nil {
		return 0, mapErr(err, "notification")
	}
	return n, nil
}

// notifyUser stores an in-app notification and attempts push delivery to the
// user's registered devices. Delivery is never reported as successful unless a
// configured provider accepted it; otherwise the attempt is stored as skipped.
func (d Deps) notifyUser(ctx context.Context, userID, kind, title, body string, data map[string]any) {
	n := &models.Notification{
		UserID: userID, Kind: kind, Title: title, Body: body, Data: data,
		PushStatus: models.NotificationSkipped, PushProvider: d.Notifier.Name(),
	}
	tokens, err := d.Store.ListActiveDeviceTokens(ctx, userID)
	if err != nil {
		d.Logger.WarnContext(ctx, "failed to list device tokens", "error", err)
	}
	switch {
	case d.Notifier.Name() == "none":
		n.PushError = "push provider not configured"
	case len(tokens) == 0:
		n.PushError = "no registered devices"
	default:
		n.PushStatus = models.NotificationPending
	}
	created, err := d.Store.CreateUserNotification(ctx, n)
	if err != nil {
		d.Logger.ErrorContext(ctx, "failed to store notification", "error", err)
		return
	}
	if n.PushStatus != models.NotificationPending {
		return
	}

	var lastErr error
	sent := false
	for _, token := range tokens {
		if _, err := d.Notifier.Send(ctx, notify.Message{
			Channel: "push", To: token, Subject: title, Body: body, Meta: data,
		}); err != nil {
			lastErr = err
			continue
		}
		sent = true
	}
	status := models.NotificationSent
	errMsg := ""
	if !sent {
		status = models.NotificationFailed
		if lastErr != nil {
			errMsg = lastErr.Error()
		}
	}
	if err := d.Store.SetNotificationPushResult(ctx, created.ID, status, d.Notifier.Name(), errMsg); err != nil {
		d.Logger.WarnContext(ctx, "failed to record push result", "error", err)
	}
}
