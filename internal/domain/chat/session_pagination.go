package chat

import "time"

type SessionCursor struct {
	SortAt    time.Time
	SessionID string
}

type SessionListItem struct {
	Session          ChatSession
	LastMessagePreview string
	UpdatedAt        time.Time
	SortAt           time.Time
}

type SessionListPage struct {
	TotalCount int
	Items      []SessionListItem
}
