package models

type DispatchRequest struct {
	// ConversationID string `json:"conversationid"`
	Message  string `json:"message"`
	Language string `json:"language"`
}
