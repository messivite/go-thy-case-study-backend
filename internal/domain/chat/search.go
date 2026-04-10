package chat

import "time"

// SearchCursor is keyset cursor for chat search pagination.
type SearchCursor struct {
	SortAt    time.Time
	SessionID string
}

// SearchChatParams contains user scoped search options.
type SearchChatParams struct {
	UserID  string
	Query   string
	Limit   int
	Cursor  *SearchCursor
}

// SearchChatHit returns one matched session and its strongest preview evidence.
type SearchChatHit struct {
	SessionID       string
	Title           string
	SessionCreatedAt time.Time
	SessionUpdatedAt time.Time
	LastMessageAt   time.Time
	TitleMatched    bool
	MatchedMessageID string
	MatchedRole     Role
	MatchedContent  string
	MatchedAt       time.Time
	SortAt          time.Time
}

type SearchChatsResult struct {
	TotalCount int
	Items      []SearchChatHit
}
