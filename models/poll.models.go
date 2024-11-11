package models

type PollRequest struct {
	ConversationID []string `json:"conversationId"` // Updated field name to match JSON convention (camelCase)
}
