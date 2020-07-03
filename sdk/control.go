package sdk

import "time"

// Control is an interface for notifying of control states
type Control interface {
	// Paused must be called when the integration is paused for any reason such as rate limiting
	Paused(resetAt time.Time) error
	// Resumed must be called when a paused integration is resumed
	Resumed() error
}
