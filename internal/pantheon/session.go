// Package pantheon provides types and client functions for interacting with Pantheon via the terminus-golang library.
package pantheon

import (
	"context"
	"fmt"
	"sync"

	"github.com/deviantintegral/terminus-golang/pkg/api"
)

// Session holds an authenticated session for one account.
type Session struct {
	MachineToken string
	SessionToken string
	UserID       string
	Email        string
	Client       *api.Client
}

// SessionManager handles authentication and client creation.
// Sessions are stored in memory only (no disk persistence).
type SessionManager struct {
	mu           sync.RWMutex
	sessions     map[string]*Session // key: machineToken
	debugEnabled bool
}

// NewSessionManager creates a new session manager.
func NewSessionManager(debug bool) *SessionManager {
	return &SessionManager{
		sessions:     make(map[string]*Session),
		debugEnabled: debug,
	}
}

// Authenticate creates a new session for a machine token.
// This always performs a fresh login, replacing any existing session.
func (sm *SessionManager) Authenticate(ctx context.Context, machineToken string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Create unauthenticated client for login with debug logging if enabled
	var client *api.Client
	if sm.debugEnabled {
		logger := api.NewLogger(api.VerbosityTrace)
		client = api.NewClient(api.WithLogger(logger))
	} else {
		client = api.NewClient()
	}

	// Authenticate with machine token
	authService := api.NewAuthService(client)
	loginResult, err := authService.Login(ctx, machineToken)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Get user email
	var email string
	user, err := authService.Whoami(ctx, loginResult.UserID)
	if err != nil {
		// Fall back to account ID from token if whoami fails
		email = GetAccountID(machineToken)
	} else {
		email = user.Email
	}

	session := &Session{
		MachineToken: machineToken,
		SessionToken: loginResult.Session,
		UserID:       loginResult.UserID,
		Email:        email,
		Client:       client,
	}

	sm.sessions[machineToken] = session
	return session, nil
}

// GetSession returns an existing session or creates a new one.
func (sm *SessionManager) GetSession(ctx context.Context, machineToken string) (*Session, error) {
	sm.mu.RLock()
	session, exists := sm.sessions[machineToken]
	sm.mu.RUnlock()

	if exists && session.Client != nil {
		return session, nil
	}

	return sm.Authenticate(ctx, machineToken)
}

// GetClient returns an authenticated API client for the given machine token.
func (sm *SessionManager) GetClient(ctx context.Context, machineToken string) (*api.Client, error) {
	session, err := sm.GetSession(ctx, machineToken)
	if err != nil {
		return nil, err
	}
	return session.Client, nil
}

// GetUserID returns the user ID for the given machine token.
func (sm *SessionManager) GetUserID(ctx context.Context, machineToken string) (string, error) {
	session, err := sm.GetSession(ctx, machineToken)
	if err != nil {
		return "", err
	}
	return session.UserID, nil
}

// GetEmail returns the email for the given machine token.
func (sm *SessionManager) GetEmail(ctx context.Context, machineToken string) (string, error) {
	session, err := sm.GetSession(ctx, machineToken)
	if err != nil {
		return "", err
	}
	return session.Email, nil
}

// InvalidateSession removes a session, forcing re-authentication on next use.
func (sm *SessionManager) InvalidateSession(machineToken string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, machineToken)
}
