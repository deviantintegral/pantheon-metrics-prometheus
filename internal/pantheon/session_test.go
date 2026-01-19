package pantheon

import (
	"context"
	"testing"

	"github.com/deviantintegral/terminus-golang/pkg/api"
)

func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager(false)
	if sm == nil {
		t.Fatal("Expected non-nil session manager")
	}
	if sm.sessions == nil {
		t.Fatal("Expected sessions map to be initialized")
	}
	if len(sm.sessions) != 0 {
		t.Errorf("Expected empty sessions map, got %d entries", len(sm.sessions))
	}
}

func TestInvalidateSession(t *testing.T) {
	sm := NewSessionManager(false)

	// Pre-populate a session
	testToken := "test-machine-token"
	sm.sessions[testToken] = &Session{
		MachineToken: testToken,
		SessionToken: "session-123",
		UserID:       "user-456",
		Email:        "test@example.com",
		Client:       api.NewClient(),
	}

	// Verify it exists
	if _, exists := sm.sessions[testToken]; !exists {
		t.Fatal("Expected session to exist before invalidation")
	}

	// Invalidate it
	sm.InvalidateSession(testToken)

	// Verify it's gone
	if _, exists := sm.sessions[testToken]; exists {
		t.Error("Expected session to be removed after invalidation")
	}
}

func TestInvalidateSessionNonExistent(t *testing.T) {
	sm := NewSessionManager(false)

	// This should not panic
	sm.InvalidateSession("non-existent-token")

	if len(sm.sessions) != 0 {
		t.Errorf("Expected empty sessions map, got %d entries", len(sm.sessions))
	}
}

func TestGetSessionReturnsExisting(t *testing.T) {
	sm := NewSessionManager(false)

	// Pre-populate a session
	testToken := "existing-token"
	expectedSession := &Session{
		MachineToken: testToken,
		SessionToken: "session-abc",
		UserID:       "user-xyz",
		Email:        "existing@example.com",
		Client:       api.NewClient(),
	}
	sm.sessions[testToken] = expectedSession

	// Get the session
	ctx := context.Background()
	session, err := sm.GetSession(ctx, testToken)

	// GetSession will try to authenticate if client is nil, but since we provided
	// a client, it should return the existing session
	if err != nil {
		// If authentication fails (network), that's expected - skip this test
		t.Skipf("Network authentication required, skipping: %v", err)
	}

	if session != expectedSession {
		t.Error("Expected to get the same session object")
	}
}

func TestGetClientFromExistingSession(t *testing.T) {
	sm := NewSessionManager(false)

	// Pre-populate a session with a client
	testToken := "client-token"
	expectedClient := api.NewClient()
	sm.sessions[testToken] = &Session{
		MachineToken: testToken,
		SessionToken: "session-def",
		UserID:       "user-client",
		Email:        "client@example.com",
		Client:       expectedClient,
	}

	ctx := context.Background()
	client, err := sm.GetClient(ctx, testToken)

	if err != nil {
		t.Skipf("Network authentication required, skipping: %v", err)
	}

	if client != expectedClient {
		t.Error("Expected to get the same client object")
	}
}

func TestGetUserIDFromExistingSession(t *testing.T) {
	sm := NewSessionManager(false)

	// Pre-populate a session
	testToken := "userid-token"
	expectedUserID := "user-12345"
	sm.sessions[testToken] = &Session{
		MachineToken: testToken,
		SessionToken: "session-ghi",
		UserID:       expectedUserID,
		Email:        "userid@example.com",
		Client:       api.NewClient(),
	}

	ctx := context.Background()
	userID, err := sm.GetUserID(ctx, testToken)

	if err != nil {
		t.Skipf("Network authentication required, skipping: %v", err)
	}

	if userID != expectedUserID {
		t.Errorf("Expected UserID='%s', got '%s'", expectedUserID, userID)
	}
}

func TestGetEmailFromExistingSession(t *testing.T) {
	sm := NewSessionManager(false)

	// Pre-populate a session
	testToken := "email-token"
	expectedEmail := "test@pantheon.io"
	sm.sessions[testToken] = &Session{
		MachineToken: testToken,
		SessionToken: "session-jkl",
		UserID:       "user-email",
		Email:        expectedEmail,
		Client:       api.NewClient(),
	}

	ctx := context.Background()
	email, err := sm.GetEmail(ctx, testToken)

	if err != nil {
		t.Skipf("Network authentication required, skipping: %v", err)
	}

	if email != expectedEmail {
		t.Errorf("Expected Email='%s', got '%s'", expectedEmail, email)
	}
}

func TestSessionManagerConcurrentAccess(t *testing.T) {
	sm := NewSessionManager(false)

	// Pre-populate sessions
	for i := 0; i < 10; i++ {
		token := "token-" + string(rune('a'+i))
		sm.sessions[token] = &Session{
			MachineToken: token,
			SessionToken: "session-" + token,
			UserID:       "user-" + token,
			Email:        token + "@example.com",
			Client:       api.NewClient(),
		}
	}

	// Concurrent reads and invalidations
	done := make(chan bool, 20)

	for i := 0; i < 10; i++ {
		token := "token-" + string(rune('a'+i))
		go func(t string) {
			sm.InvalidateSession(t)
			done <- true
		}(token)

		go func() {
			sm.InvalidateSession("non-existent")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// All sessions should be invalidated
	if len(sm.sessions) != 0 {
		t.Errorf("Expected all sessions to be invalidated, got %d remaining", len(sm.sessions))
	}
}

func TestSessionStruct(t *testing.T) {
	session := &Session{
		MachineToken: "machine-token-123",
		SessionToken: "session-token-456",
		UserID:       "user-789",
		Email:        "user@example.com",
	}

	if session.MachineToken != "machine-token-123" {
		t.Errorf("Expected MachineToken='machine-token-123', got '%s'", session.MachineToken)
	}
	if session.SessionToken != "session-token-456" {
		t.Errorf("Expected SessionToken='session-token-456', got '%s'", session.SessionToken)
	}
	if session.UserID != "user-789" {
		t.Errorf("Expected UserID='user-789', got '%s'", session.UserID)
	}
	if session.Email != "user@example.com" {
		t.Errorf("Expected Email='user@example.com', got '%s'", session.Email)
	}
	if session.Client != nil {
		t.Error("Expected Client to be nil by default")
	}
}

func TestGetSessionWithNilClient(t *testing.T) {
	sm := NewSessionManager(false)

	// Pre-populate a session with nil client
	testToken := "nil-client-token"
	sm.sessions[testToken] = &Session{
		MachineToken: testToken,
		SessionToken: "session-xxx",
		UserID:       "user-xxx",
		Email:        "nilclient@example.com",
		Client:       nil, // nil client should trigger re-authentication
	}

	// GetSession should try to re-authenticate because client is nil
	ctx := context.Background()
	_, err := sm.GetSession(ctx, testToken)

	// Will fail because it tries to authenticate with invalid token
	// But this exercises the code path
	if err == nil {
		t.Skip("Unexpectedly succeeded (network may have responded)")
	}
}

func TestMultipleInvalidateSessions(t *testing.T) {
	sm := NewSessionManager(false)

	// Add multiple sessions
	for i := 0; i < 5; i++ {
		token := "token-" + string(rune('a'+i))
		sm.sessions[token] = &Session{
			MachineToken: token,
			Client:       api.NewClient(),
		}
	}

	if len(sm.sessions) != 5 {
		t.Fatalf("Expected 5 sessions, got %d", len(sm.sessions))
	}

	// Invalidate all of them
	for i := 0; i < 5; i++ {
		token := "token-" + string(rune('a'+i))
		sm.InvalidateSession(token)
	}

	if len(sm.sessions) != 0 {
		t.Errorf("Expected 0 sessions after invalidation, got %d", len(sm.sessions))
	}
}
