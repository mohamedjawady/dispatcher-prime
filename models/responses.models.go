package models

import "go.mongodb.org/mongo-driver/bson"

// Response is used for standardizing the API responses.
type Response struct {
	Status         string      `json:"status"`                   // Status of the request (success/error)
	Error          string      `json:"error,omitempty"`          // Error message if any
	Message        interface{} `json:"message,omitempty"`        // Main message or data of the response
	ConversationID string      `json:"conversationId,omitempty"` // ID of the conversation
	Messages       []bson.M    `json:"messages,omitempty"`       // List of messages for the conversation (this is the field causing the issue)
}

type Conversations struct {
	Status        string   `json:"status"`
	Error         string   `json:"error,omitempty"`
	Conversations []bson.M `json:"conversationids,omitempty"`
}
