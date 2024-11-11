package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mohamedjawady/dispatcher-prime/models"
	"github.com/mohamedjawady/dispatcher-prime/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func PollHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// category := vars["category"]
	id := vars["id"]

	var response models.Response
	w.Header().Set("Content-Type", "application/json")

	// Validate and decode JWT token
	token := r.Header.Get("Authorization")
	decodedToken, err := utils.DecodeJWT(token)
	if err != nil || !decodedToken.Valid {
		response.Status = "error"
		response.Error = "Authorization token is invalid or missing"
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Safely extract email from decoded token
	email, ok := decodedToken.Decoded["email"].(string)
	if !ok || email == "" {
		response.Status = "error"
		response.Error = "email not found in token"
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// MongoDB connection setup
	collection := Client.Database("test").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query for the specific user and conversation ID in the "conversations" array
	var userDoc struct {
		Conversations []struct {
			ConversationID string `bson:"conversationId"`
			Messages       []bson.M
		} `bson:"conversations"`
	}

	err = collection.FindOne(ctx, bson.M{
		"email":                        email,
		"conversations.conversationId": id,
	}).Decode(&userDoc)

	if err == mongo.ErrNoDocuments {
		response.Status = "error"
		response.Error = fmt.Sprintf("Conversation with ID %s not found for user %s", id, email)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	} else if err != nil {
		log.Printf("Failed to fetch conversation: %v", err)
		response.Status = "error"
		response.Error = "Failed to fetch conversation"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Find the specific conversation from the user document
	var conversationMessages []bson.M
	for _, conversation := range userDoc.Conversations {
		if conversation.ConversationID == id {
			conversationMessages = conversation.Messages
			break
		}
	}

	// Return the messages in the conversation
	response.Status = "success"
	response.ConversationID = id
	response.Messages = conversationMessages // Assign the messages to the Messages field

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func PollHandlerAll(w http.ResponseWriter, r *http.Request) {
	// Set the response type
	var response models.Conversations
	w.Header().Set("Content-Type", "application/json")

	// Validate and decode JWT token
	token := r.Header.Get("Authorization")
	decodedToken, err := utils.DecodeJWT(token)
	if err != nil || !decodedToken.Valid {
		response.Status = "error"
		response.Error = "Authorization token is invalid or missing"
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Safely extract email from decoded token
	email, ok := decodedToken.Decoded["email"].(string)
	if !ok || email == "" {
		response.Status = "error"
		response.Error = "email not found in token"
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// MongoDB connection setup
	collection := Client.Database("test").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query for the specific user and fetch their conversations
	var userDoc struct {
		Conversations []struct {
			ConversationID string `bson:"conversationId"`
			Messages       []struct {
				Message   string    `bson:"message"`
				Timestamp time.Time `bson:"timestamp"`
			} `bson:"messages"`
		} `bson:"conversations"`
	}

	// Query to find user based on email
	err = collection.FindOne(ctx, bson.M{"email": email}).Decode(&userDoc)

	if err == mongo.ErrNoDocuments {
		response.Status = "error"
		response.Error = fmt.Sprintf("No conversations found for user %s", email)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	} else if err != nil {
		log.Printf("Failed to fetch user conversations: %v", err)
		response.Status = "error"
		response.Error = "Failed to fetch user conversations"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Extract all conversation IDs along with the timestamp of the first message
	var conversationDetails []bson.M
	for _, conversation := range userDoc.Conversations {
		// Get the first message's timestamp
		var firstMessageTimestamp time.Time
		if len(conversation.Messages) > 0 {
			firstMessageTimestamp = conversation.Messages[0].Timestamp
		}

		// Add conversationID and the timestamp of the first message to the response
		conversationDetails = append(conversationDetails, bson.M{
			"conversationId":        conversation.ConversationID,
			"firstMessageTimestamp": firstMessageTimestamp,
		})
	}

	// Populate response
	response.Status = "success"
	response.Conversations = conversationDetails

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
