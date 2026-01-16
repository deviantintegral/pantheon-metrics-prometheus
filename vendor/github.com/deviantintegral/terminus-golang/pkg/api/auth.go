// Package api provides a client for the Pantheon API.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/deviantintegral/terminus-golang/pkg/api/models"
)

// AuthService handles authentication operations
type AuthService struct {
	client *Client
}

// NewAuthService creates a new auth service
func NewAuthService(client *Client) *AuthService {
	return &AuthService{client: client}
}

// LoginRequest represents a machine token login request
type LoginRequest struct {
	MachineToken string `json:"machine_token"`
	Client       string `json:"client"`
}

// SessionResponse represents the session response from login
type SessionResponse struct {
	Session   string `json:"session"`
	UserID    string `json:"user_id"`
	ExpiresAt int64  `json:"expires_at"`
}

// Login authenticates using a machine token and returns a session
func (s *AuthService) Login(ctx context.Context, machineToken string) (*SessionResponse, error) {
	req := LoginRequest{
		MachineToken: machineToken,
		Client:       "terminus-golang",
	}

	resp, err := s.client.Post(ctx, "/authorize/machine-token", req)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	var session SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode session response: %w", err)
	}

	// Update client token
	s.client.SetToken(session.Session)

	return &session, nil
}

// Whoami returns information about the current user
func (s *AuthService) Whoami(ctx context.Context, userID string) (*models.User, error) {
	path := fmt.Sprintf("/users/%s", userID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("whoami request failed: %w", err)
	}

	var user models.User
	if err := DecodeResponse(resp, &user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

// ValidateSession checks if the current session is valid
func (s *AuthService) ValidateSession(ctx context.Context, userID string) (bool, error) {
	if s.client.token == "" {
		return false, nil
	}

	_, err := s.Whoami(ctx, userID)
	if err != nil {
		if IsNotFound(err) || (err != nil && err.Error() == "API error 401") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
