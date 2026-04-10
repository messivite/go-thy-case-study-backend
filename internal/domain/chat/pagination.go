package chat

import "time"

type MessageCursor struct {
	CreatedAt time.Time
	MessageID string
}
