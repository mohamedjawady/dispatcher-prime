package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// DecodedJWT represents the structure of the decoded JWT token.
type DecodedJWT struct {
	Valid   bool
	Decoded map[string]interface{}
}

// TokenVerificationResponse is the response structure from the external token verification service.
type TokenVerificationResponse struct {
	Valid   bool                   `json:"valid"`
	Message string                 `json:"message,omitempty"`
	Decoded map[string]interface{} `json:"decoded,omitempty"`
}

// DecodeJWT sends the JWT token to an external service for validation.
func DecodeJWT(tokenString string) (*DecodedJWT, error) {
	// Prepare the payload to send to the verification endpoint
	payload := map[string]string{
		"token": tokenString,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error encoding token payload: %v", err)
	}

	// Send the POST request to the token verification service
	resp, err := http.Post("http://172.20.10.4:3000/verifyTokens", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("error sending token verification request: %v", err)
	}
	defer resp.Body.Close()

	// Decode the response
	var verificationResponse TokenVerificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&verificationResponse); err != nil {
		return nil, fmt.Errorf("error decoding verification response: %v", err)
	}

	fmt.Println(verificationResponse)
	if verificationResponse.Valid {
		return &DecodedJWT{
			Valid:   true,
			Decoded: verificationResponse.Decoded,
		}, nil
	}

	// Return invalid if not valid
	return &DecodedJWT{Valid: false}, nil
}
