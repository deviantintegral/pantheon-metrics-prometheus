package api

import (
	"context"
	"fmt"
)

// SessionTokenRefresher implements TokenRefresher using machine token authentication.
// It refreshes the session token by calling the authentication endpoint with the
// stored machine token.
type SessionTokenRefresher struct {
	// getMachineToken is a callback that returns the machine token to use for refresh.
	// This allows the caller to load the token from storage (e.g., token files).
	getMachineToken func() (string, error)
	// client is the API client used to make the refresh request
	// Note: This creates a circular reference, but it's necessary for the refresh flow.
	// Infinite loops are prevented by using PostOnlyOnce for auth requests, which
	// bypasses the retry logic and token refresh mechanism.
	client *Client
	// onTokenRefreshed is an optional callback invoked when a token is successfully refreshed.
	// This allows the caller to persist the new session information.
	onTokenRefreshed func(session *SessionResponse) error
	// logger is an optional logger for debug output
	logger Logger
}

// SessionTokenRefresherOption is a function that configures a SessionTokenRefresher
type SessionTokenRefresherOption func(*SessionTokenRefresher)

// WithRefreshLogger sets the logger for the token refresher
func WithRefreshLogger(logger Logger) SessionTokenRefresherOption {
	return func(r *SessionTokenRefresher) {
		r.logger = logger
	}
}

// WithOnTokenRefreshed sets the callback invoked when a token is refreshed
func WithOnTokenRefreshed(callback func(session *SessionResponse) error) SessionTokenRefresherOption {
	return func(r *SessionTokenRefresher) {
		r.onTokenRefreshed = callback
	}
}

// NewSessionTokenRefresher creates a new SessionTokenRefresher.
// The getMachineToken callback is called when a token refresh is needed to get
// the current machine token. This allows loading tokens from storage dynamically.
// The client is the API client to use for the refresh request.
func NewSessionTokenRefresher(getMachineToken func() (string, error), client *Client, opts ...SessionTokenRefresherOption) *SessionTokenRefresher {
	r := &SessionTokenRefresher{
		getMachineToken: getMachineToken,
		client:          client,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// RefreshToken implements the TokenRefresher interface.
// It uses the stored machine token to obtain a new session token from the API.
func (r *SessionTokenRefresher) RefreshToken(ctx context.Context) (string, error) {
	if r.getMachineToken == nil {
		return "", fmt.Errorf("no machine token provider configured")
	}

	machineToken, err := r.getMachineToken()
	if err != nil {
		return "", fmt.Errorf("failed to get machine token: %w", err)
	}

	if machineToken == "" {
		return "", fmt.Errorf("no machine token available for token refresh")
	}

	if r.logger != nil {
		r.logger.Debug("Refreshing session token using machine token")
	}

	// Create a temporary auth service to perform the login
	// Note: We use the existing client which will use the same base URL and settings.
	// The Login method uses PostOnlyOnce which bypasses retry/refresh logic,
	// preventing infinite recursion if the auth endpoint returns 401.
	authService := NewAuthService(r.client)
	session, err := authService.Login(ctx, machineToken)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	if r.logger != nil {
		r.logger.Debug("Session token refreshed successfully")
	}

	// Call the callback if set (to persist the new session)
	if r.onTokenRefreshed != nil {
		if err := r.onTokenRefreshed(session); err != nil {
			if r.logger != nil {
				r.logger.Warn("Failed to save refreshed session: %v", err)
			}
			// Continue anyway since we have the new token
		}
	}

	return session.Session, nil
}
