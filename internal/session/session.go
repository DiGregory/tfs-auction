package session

import (
	"time"

)

type Session struct {
	SessionID string
	UserID    int64
	CreatedAt time.Time
 	ValidUntil *time.Time
}

