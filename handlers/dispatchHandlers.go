package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mohamedjawady/dispatcher-prime/models"
	"github.com/mohamedjawady/dispatcher-prime/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func DispatchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["category"]
	id := vars["id"]

	var response models.Response
	w.Header().Set("Content-Type", "application/json")

	// Parse the request body
	var requestBody struct {
		Message  string `json:"message"`
		Language string `json:"language"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		response.Status = "error"
		response.Error = "Failed to parse request body"
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

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

	// Get the current timestamp
	timestamp := time.Now().Unix()

	// If conversationId is not provided (i.e., create new conversation)
	if id == "" {
		// Create a new conversation ID
		newConversationID := uuid.New().String()
		response.Status = "success"
		response.ConversationID = newConversationID
		response.Message = fmt.Sprintf("No ID provided. Created new conversation with ID: %s", newConversationID)

		// Append the new conversation to the user's conversations with a timestamp
		_, err := collection.UpdateOne(
			ctx,
			bson.M{"email": email},
			bson.M{"$push": bson.M{
				"conversations": bson.M{
					"name":           category,
					"conversationId": newConversationID,
					"messages": []bson.M{{
						"message":   requestBody.Message,
						"language":  requestBody.Language,
						"timestamp": timestamp, // Add timestamp here
					}},
				},
			}})

		if err != nil {
			log.Printf("Failed to add new conversation: %v", err)
			response.Status = "error"
			response.Error = "Failed to add new conversation"
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
	} else {
		// Check if the conversation exists for this user by ID
		var userDoc struct {
			Conversations []struct {
				ConversationID string `bson:"conversationId"`
				Messages       []bson.M
			} `bson:"conversations"`
		}

		// Query for the specific user and conversation ID in the "conversations" array
		err := collection.FindOne(ctx, bson.M{
			"email": email,
		}).Decode(&userDoc)

		fmt.Println(err)

		if err == mongo.ErrNoDocuments {
			// If no document found, create a new conversation
			newConversationID := uuid.New().String()
			response.Status = "success"
			response.ConversationID = newConversationID
			response.Message = fmt.Sprintf("Conversation with ID %s not found. Created new conversation with ID: %s", id, newConversationID)

			// Add the new conversation to the user's conversations
			_, err := collection.UpdateOne(
				ctx,
				bson.M{"email": email},
				bson.M{"$push": bson.M{
					"conversations": bson.M{
						"name":           category,
						"conversationId": newConversationID,
						"messages": []bson.M{{
							"message":   requestBody.Message,
							"language":  requestBody.Language,
							"timestamp": timestamp, // Add timestamp here
						}},
					},
				}})

			if err != nil {
				log.Printf("Failed to add new conversation: %v", err)
				response.Status = "error"
				response.Error = "Failed to add new conversation"
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}
		} else if err != nil {
			log.Printf("Failed to fetch conversation: %v", err)
			response.Status = "error"
			response.Error = "Failed to fetch conversation"
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		} else {
			// If the conversation exists, append the message to the existing conversation
			_, err = collection.UpdateOne(
				ctx,
				bson.M{
					"email":                        email,
					"conversations.conversationId": id,
				},
				bson.M{
					"$push": bson.M{
						"conversations.$.messages": bson.M{
							"message":   requestBody.Message,
							"language":  requestBody.Language,
							"timestamp": timestamp, // Add timestamp here
						},
					},
				})

			if err != nil {
				log.Printf("Failed to update conversation %s: %v", id, err)
				response.Status = "error"
				response.Error = "Failed to update conversation"
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}

			response.Status = "success"
			response.ConversationID = id
			response.Message = fmt.Sprintf("Message appended to conversation with ID: %s", id)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
