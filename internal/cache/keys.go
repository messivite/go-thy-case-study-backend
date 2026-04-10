package cache

import "net/url"

// PrefixChatList invalidates all GET /api/chats cache entries for a user.
func PrefixChatList(userID string) string {
	return "v1:resp:chatlist:" + url.QueryEscape(userID) + ":"
}

// PrefixChatMessages invalidates GET /api/chats/{id}/messages for one session.
func PrefixChatMessages(userID, chatID string) string {
	return "v1:resp:chatmsgs:" + url.QueryEscape(userID) + ":" + url.QueryEscape(chatID) + ":"
}

func KeyChatList(userID, rawQuery string) string {
	return PrefixChatList(userID) + url.QueryEscape(rawQuery)
}

func KeyChatMessages(userID, chatID, rawQuery string) string {
	return PrefixChatMessages(userID, chatID) + url.QueryEscape(rawQuery)
}
