package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"runtime"
	"time"

	"github.com/deviantintegral/terminus-golang/pkg/version"
	"github.com/google/uuid"
)

const (
	// DefaultBaseURL is the base URL for the Pantheon API
	DefaultBaseURL = "https://terminus.pantheon.io:443/api"

	// DefaultTimeout is the default timeout for HTTP requests
	DefaultTimeout = 86400 * time.Second

	// MaxRetries is the maximum number of retry attempts
	MaxRetries = 5

	// InitialBackoff is the initial backoff duration for retries
	InitialBackoff = 1 * time.Second
)

// Client is the HTTP client for the Pantheon API
type Client struct {
	baseURL        string
	httpClient     *http.Client
	userAgent      string
	token          string
	logger         Logger
	tokenRefresher TokenRefresher
}

// Logger is an interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// TokenRefresher is an interface for refreshing authentication tokens
// It is used to automatically renew session tokens when they expire or
// when the API returns a 401 Unauthorized error.
type TokenRefresher interface {
	// RefreshToken attempts to refresh the authentication token.
	// It should return the new token on success, or an error if the refresh fails.
	// The implementation is responsible for using the machine token to obtain
	// a new session token.
	RefreshToken(ctx context.Context) (string, error)
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// NewClient creates a new API client
func NewClient(options ...ClientOption) *Client {
	c := &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		userAgent: fmt.Sprintf("Terminus-Go/%s (go_version=%s; os=%s; arch=%s)",
			version.String(), runtime.Version(), runtime.GOOS, runtime.GOARCH),
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

// WithBaseURL sets a custom base URL
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithToken sets the authentication token
func WithToken(token string) ClientOption {
	return func(c *Client) {
		c.token = token
	}
}

// WithLogger sets a custom logger
func WithLogger(logger Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithUserAgent sets a custom User-Agent header.
// This allows downstream applications using this library to identify themselves.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// WithTokenRefresher sets a token refresher for automatic token renewal
func WithTokenRefresher(refresher TokenRefresher) ClientOption {
	return func(c *Client) {
		c.tokenRefresher = refresher
	}
}

// SetToken updates the authentication token
func (c *Client) SetToken(token string) {
	c.token = token
}

// SetTokenRefresher sets a token refresher for automatic token renewal
func (c *Client) SetTokenRefresher(refresher TokenRefresher) {
	c.tokenRefresher = refresher
}

// Request makes an HTTP request to the API with retry logic
func (c *Client) Request(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	fullURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add trace ID
	traceID := uuid.New().String()
	req.Header.Set("X-Pantheon-Trace-Id", traceID)

	if c.logger != nil {
		c.logger.Debug("API Request: %s %s (trace: %s)", method, fullURL, traceID)

		// Log detailed HTTP request at trace level
		if httpLogger, ok := AsHTTPLogger(c.logger); ok && httpLogger.IsTraceEnabled() {
			headers := make(map[string][]string)
			for k, v := range req.Header {
				headers[k] = v
			}
			bodyStr := ""
			if len(bodyBytes) > 0 {
				bodyStr = string(bodyBytes)
			}
			httpLogger.LogHTTPRequest(method, fullURL, headers, bodyStr)
		}
	}

	// Execute request with retry logic
	return c.doWithRetry(req)
}

// doWithRetry executes an HTTP request with exponential backoff retry logic
// It also handles 401 Unauthorized errors by attempting to refresh the token
// and retrying the request once.
func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	tokenRefreshAttempted := false

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		// Clone the request body for retries
		bodyReader, cloneErr := c.cloneRequestBody(req)
		if cloneErr != nil {
			return nil, cloneErr
		}

		resp, err = c.httpClient.Do(req)

		if c.shouldStopRetrying(resp, err) {
			c.logResponse(resp)
			// Check if the response indicates an error (4XX or 5XX)
			if resp.StatusCode >= 400 {
				body, _ := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				apiErr := &Error{
					StatusCode: resp.StatusCode,
					Message:    string(body),
				}

				// Handle 401 Unauthorized with token refresh
				if resp.StatusCode == http.StatusUnauthorized &&
					c.tokenRefresher != nil &&
					!tokenRefreshAttempted {
					tokenRefreshAttempted = true
					if c.logger != nil {
						c.logger.Debug("Received 401 Unauthorized, attempting token refresh")
					}

					// Get context from request
					ctx := req.Context()
					newToken, refreshErr := c.tokenRefresher.RefreshToken(ctx)
					if refreshErr != nil {
						if c.logger != nil {
							c.logger.Warn("Token refresh failed: %v", refreshErr)
						}
						return nil, apiErr
					}

					// Update client token and retry
					c.token = newToken
					req.Header.Set("Authorization", "Bearer "+newToken)
					c.restoreRequestBody(req, bodyReader)
					if c.logger != nil {
						c.logger.Debug("Token refreshed successfully, retrying request")
					}
					continue
				}

				return nil, apiErr
			}
			return resp, nil
		}

		c.logRetryAttempt(err, resp, attempt)

		if attempt < MaxRetries {
			c.sleepWithBackoff(attempt)
			c.restoreRequestBody(req, bodyReader)
		}

		c.closeResponseBody(resp)
	}

	return c.formatRetryError(err, resp)
}

// cloneRequestBody clones the request body for retry attempts
func (c *Client) cloneRequestBody(req *http.Request) (io.Reader, error) {
	if req.Body == nil {
		return nil, nil
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	_ = req.Body.Close()

	bodyReader := bytes.NewReader(bodyBytes)
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return bodyReader, nil
}

// shouldStopRetrying determines if we should stop retrying
func (c *Client) shouldStopRetrying(resp *http.Response, err error) bool {
	return err == nil && !shouldRetry(resp.StatusCode)
}

// logResponse logs the response if trace logging is enabled
func (c *Client) logResponse(resp *http.Response) {
	if httpLogger, ok := AsHTTPLogger(c.logger); ok && httpLogger.IsTraceEnabled() {
		c.logHTTPResponse(resp, httpLogger)
	}
}

// logRetryAttempt logs a retry attempt
func (c *Client) logRetryAttempt(err error, resp *http.Response, attempt int) {
	if c.logger == nil {
		return
	}

	if err != nil {
		c.logger.Warn("Request failed (attempt %d/%d): %v", attempt+1, MaxRetries+1, err)
	} else {
		c.logger.Warn("Request returned %d (attempt %d/%d)", resp.StatusCode, attempt+1, MaxRetries+1)
	}
}

// sleepWithBackoff sleeps with exponential backoff
func (c *Client) sleepWithBackoff(attempt int) {
	backoff := time.Duration(math.Pow(2, float64(attempt))) * InitialBackoff
	if c.logger != nil {
		c.logger.Debug("Retrying after %v", backoff)
	}
	time.Sleep(backoff)
}

// restoreRequestBody restores the request body for retry
func (c *Client) restoreRequestBody(req *http.Request, bodyReader io.Reader) {
	if bodyReader != nil {
		req.Body = io.NopCloser(bodyReader)
	}
}

// closeResponseBody safely closes a response body
func (c *Client) closeResponseBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}

// formatRetryError formats the final error after all retries are exhausted
func (c *Client) formatRetryError(err error, resp *http.Response) (*http.Response, error) {
	if err != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w", MaxRetries+1, err)
	}
	return resp, fmt.Errorf("request failed with status %d after %d attempts", resp.StatusCode, MaxRetries+1)
}

// logHTTPResponse logs HTTP response details while preserving the response body
func (c *Client) logHTTPResponse(resp *http.Response, httpLogger HTTPLogger) {
	if resp == nil || resp.Body == nil {
		return
	}

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		httpLogger.Error("Failed to read response body for logging: %v", err)
		return
	}
	_ = resp.Body.Close()

	// Restore the response body for the caller
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	// Convert headers to map
	headers := make(map[string][]string)
	for k, v := range resp.Header {
		headers[k] = v
	}

	// Log the response
	httpLogger.LogHTTPResponse(resp.StatusCode, resp.Status, headers, string(bodyBytes))
}

// shouldRetry determines if a status code should trigger a retry
func shouldRetry(statusCode int) bool {
	// Retry on 5xx server errors and 429 rate limiting
	return statusCode >= 500 || statusCode == 429
}

// Get makes a GET request
func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.Request(ctx, http.MethodGet, path, nil)
}

// Post makes a POST request
func (c *Client) Post(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPost, path, body)
}

// Put makes a PUT request
func (c *Client) Put(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPut, path, body)
}

// Delete makes a DELETE request
func (c *Client) Delete(ctx context.Context, path string) (*http.Response, error) {
	return c.Request(ctx, http.MethodDelete, path, nil)
}

// Patch makes a PATCH request
func (c *Client) Patch(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPatch, path, body)
}

// buildPagedPath constructs a URL with pagination parameters
func buildPagedPath(basePath string, limit int, start string) string {
	separator := "?"
	// Check if basePath already has query parameters
	for i := len(basePath) - 1; i >= 0; i-- {
		if basePath[i] == '?' {
			separator = "&"
			break
		}
	}

	path := fmt.Sprintf("%s%slimit=%d", basePath, separator, limit)
	if start != "" {
		path = fmt.Sprintf("%s&start=%s", path, start)
	}
	return path
}

// processPageResults processes a page of results, extracting IDs and detecting duplicates
func processPageResults(results []json.RawMessage, seenIDs map[string]bool, allResults *[]json.RawMessage) (lastID string, foundDuplicate bool) {
	for _, result := range results {
		// Try to extract ID from the result
		var item struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(result, &item); err == nil && item.ID != "" {
			// Check if we've seen this ID before (duplicate detection)
			if seenIDs[item.ID] {
				// We've received a duplicate, pagination is complete
				return "", true
			}
			seenIDs[item.ID] = true
			*allResults = append(*allResults, result)
			lastID = item.ID
		} else {
			// If we can't extract an ID, just add the result
			*allResults = append(*allResults, result)
		}
	}
	return lastID, false
}

// GetPaged makes paginated GET requests using cursor-based pagination and returns all results
// The Pantheon API uses cursor-based pagination with 'start' parameter (ID of last item)
// rather than page-based pagination
func (c *Client) GetPaged(ctx context.Context, basePath string) ([]json.RawMessage, error) {
	var allResults []json.RawMessage
	seenIDs := make(map[string]bool) // Track IDs to detect duplicates
	limit := 100
	var start string // Cursor for pagination (ID of last item from previous page)

	for {
		path := buildPagedPath(basePath, limit, start)

		resp, err := c.Get(ctx, path)
		if err != nil {
			return nil, err
		}

		var results []json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		_ = resp.Body.Close()

		if len(results) == 0 {
			break
		}

		// Process results and check for duplicates
		lastID, foundDuplicate := processPageResults(results, seenIDs, &allResults)
		if foundDuplicate {
			break
		}

		// If we got fewer results than the limit, we're done
		if len(results) < limit {
			break
		}

		// If we couldn't extract any IDs, we can't paginate further
		if lastID == "" {
			break
		}

		// Update cursor to the last ID for next page
		start = lastID
	}

	return allResults, nil
}

// DecodeResponse decodes a JSON response into a target struct
func DecodeResponse(resp *http.Response, target interface{}) error {
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return &Error{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// Error represents an API error response
type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// IsNotFound returns true if the error is a 404 Not Found
func IsNotFound(err error) bool {
	if apiErr, ok := err.(*Error); ok {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsConflict returns true if the error is a 409 Conflict
func IsConflict(err error) bool {
	if apiErr, ok := err.(*Error); ok {
		return apiErr.StatusCode == http.StatusConflict
	}
	return false
}

// IsUnauthorized returns true if the error is a 401 Unauthorized
func IsUnauthorized(err error) bool {
	if apiErr, ok := err.(*Error); ok {
		return apiErr.StatusCode == http.StatusUnauthorized
	}
	return false
}
