package notifier

import (
	"jobradar/internal/model"
)

// Notifier is the interface for notification channels
type Notifier interface {
	// Name returns the name of the notification channel
	Name() string

	// Send sends a notification for a matched job
	Send(matched *model.MatchedJob) error
}
