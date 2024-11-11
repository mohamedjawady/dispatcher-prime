package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mohamedjawady/dispatcher-prime/handlers"
	"github.com/rs/cors"
)

func main() {
	// Initialize MongoDB client with new URI
	mongoURI := "mongodb+srv://chatBotUser:chatBotUser@chatbotisi.ylsnk.mongodb.net/?retryWrites=true&w=majority&appName=ChatBotIsi"
	if err := handlers.InitializeMongoClient(mongoURI); err != nil {
		log.Fatalf("Could not initialize MongoDB client: %v", err)
	}
	defer handlers.Client.Disconnect(context.Background())

	// Set up routes and server
	r := mux.NewRouter()
	dispatch := r.PathPrefix("/dispatch/").Subrouter()
	dispatch.HandleFunc("/{category}/{id:(?:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})?}", handlers.DispatchHandler)

	poll := r.PathPrefix("/poll/").Subrouter()
	poll.HandleFunc("/{category}/", handlers.PollHandlerAll)
	poll.HandleFunc("/{category}/{id:(?:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})?}", handlers.PollHandler)

	// Enable CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(r)

	fmt.Println("Listening on 0.0.0.0:8000")

	srv := &http.Server{
		Handler:      corsHandler, // Use the CORS handler
		Addr:         ":8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
