package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/deviantintegral/terminus-golang/pkg/api/models"
)

// UsersService handles user-related operations
type UsersService struct {
	client *Client
}

// NewUsersService creates a new users service
func NewUsersService(client *Client) *UsersService {
	return &UsersService{client: client}
}

// ListMachineTokens returns machine tokens for the authenticated user
func (s *UsersService) ListMachineTokens(ctx context.Context, userID string) ([]*models.MachineToken, error) {
	path := fmt.Sprintf("/users/%s/machine_tokens", userID)

	rawResults, err := s.client.GetPaged(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list machine tokens: %w", err)
	}

	tokens := make([]*models.MachineToken, 0, len(rawResults))
	for _, raw := range rawResults {
		var token models.MachineToken
		if err := json.Unmarshal(raw, &token); err != nil {
			return nil, fmt.Errorf("failed to decode machine token: %w", err)
		}
		tokens = append(tokens, &token)
	}

	return tokens, nil
}

// ListSSHKeys returns SSH keys for the authenticated user
func (s *UsersService) ListSSHKeys(ctx context.Context, userID string) ([]*models.SSHKey, error) {
	path := fmt.Sprintf("/users/%s/keys", userID)

	rawResults, err := s.client.GetPaged(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list SSH keys: %w", err)
	}

	keys := make([]*models.SSHKey, 0, len(rawResults))
	for _, raw := range rawResults {
		var key models.SSHKey
		if err := json.Unmarshal(raw, &key); err != nil {
			return nil, fmt.Errorf("failed to decode SSH key: %w", err)
		}
		keys = append(keys, &key)
	}

	return keys, nil
}

// ListPaymentMethods returns payment methods for the authenticated user
func (s *UsersService) ListPaymentMethods(ctx context.Context, userID string) ([]*models.PaymentMethod, error) {
	path := fmt.Sprintf("/users/%s/instruments", userID)

	rawResults, err := s.client.GetPaged(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list payment methods: %w", err)
	}

	methods := make([]*models.PaymentMethod, 0, len(rawResults))
	for _, raw := range rawResults {
		var method models.PaymentMethod
		if err := json.Unmarshal(raw, &method); err != nil {
			return nil, fmt.Errorf("failed to decode payment method: %w", err)
		}
		methods = append(methods, &method)
	}

	return methods, nil
}
