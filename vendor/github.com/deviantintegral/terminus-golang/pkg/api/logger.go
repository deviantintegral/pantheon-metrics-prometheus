package api

import (
	"io"
	"log"
	"os"
	"regexp"
)

// VerbosityLevel represents the logging verbosity level
type VerbosityLevel int

const (
	// VerbosityNone disables all logging
	VerbosityNone VerbosityLevel = 0
	// VerbosityInfo enables info level logging (-v)
	VerbosityInfo VerbosityLevel = 1
	// VerbosityDebug enables debug level logging (-vv)
	VerbosityDebug VerbosityLevel = 2
	// VerbosityTrace enables trace level logging including HTTP details (-vvv)
	VerbosityTrace VerbosityLevel = 3
)

// sensitiveFieldPatterns contains regex patterns for sensitive data that should be redacted in logs.
// The pattern (?:[^"\\]|\\.)* properly handles JSON escape sequences like \" and \\
// to prevent tokens containing escaped characters from bypassing redaction.
var sensitiveFieldPatterns = []*regexp.Regexp{
	// Machine token in request bodies (snake_case and PascalCase)
	regexp.MustCompile(`"machine_token"\s*:\s*"((?:[^"\\]|\\.){20,})"`),
	regexp.MustCompile(`"MachineToken"\s*:\s*"((?:[^"\\]|\\.){20,})"`),
	// Session token in response bodies (snake_case and PascalCase)
	regexp.MustCompile(`"session"\s*:\s*"((?:[^"\\]|\\.){20,})"`),
	regexp.MustCompile(`"Session"\s*:\s*"((?:[^"\\]|\\.){20,})"`),
	// Session token alternate field name (snake_case and PascalCase)
	regexp.MustCompile(`"session_token"\s*:\s*"((?:[^"\\]|\\.){20,})"`),
	regexp.MustCompile(`"SessionToken"\s*:\s*"((?:[^"\\]|\\.){20,})"`),
	// User IDs (UUIDs)
	regexp.MustCompile(`"user_id"\s*:\s*"[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}"`),
	regexp.MustCompile(`"UserID"\s*:\s*"[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}"`),
	regexp.MustCompile(`"id"\s*:\s*"[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}"`),
	// Email addresses
	regexp.MustCompile(`"email"\s*:\s*"[^"]+@[^"]+\.[^"]+"`),
	regexp.MustCompile(`"Email"\s*:\s*"[^"]+@[^"]+\.[^"]+"`),
}

// sensitiveFieldReplacements contains the replacement strings for each pattern
var sensitiveFieldReplacements = []string{
	`"machine_token": "REDACTED"`,
	`"MachineToken": "REDACTED"`,
	`"session": "REDACTED"`,
	`"Session": "REDACTED"`,
	`"session_token": "REDACTED"`,
	`"SessionToken": "REDACTED"`,
	`"user_id": "REDACTED-USER-ID"`,
	`"UserID": "REDACTED-USER-ID"`,
	`"id": "REDACTED-ID"`,
	`"email": "redacted@example.com"`,
	`"Email": "redacted@example.com"`,
}

// RedactSensitiveData redacts sensitive tokens from a string (typically a JSON body).
// It handles machine tokens, session tokens, user IDs, and email addresses.
// This function is also used by test fixtures to redact sensitive data before saving.
func RedactSensitiveData(data string) string {
	result := data
	for i, pattern := range sensitiveFieldPatterns {
		result = pattern.ReplaceAllString(result, sensitiveFieldReplacements[i])
	}
	return result
}

// DefaultLogger is a default implementation of the Logger interface
type DefaultLogger struct {
	verbosity VerbosityLevel
	logger    *log.Logger
}

// NewLogger creates a new logger with the specified verbosity level
func NewLogger(verbosity VerbosityLevel) *DefaultLogger {
	return &DefaultLogger{
		verbosity: verbosity,
		logger:    log.New(os.Stderr, "", log.LstdFlags),
	}
}

// NewLoggerWithWriter creates a new logger with a custom writer
func NewLoggerWithWriter(verbosity VerbosityLevel, w io.Writer) *DefaultLogger {
	return &DefaultLogger{
		verbosity: verbosity,
		logger:    log.New(w, "", log.LstdFlags),
	}
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	if l.verbosity >= VerbosityDebug {
		l.logger.Printf("[DEBUG] "+msg, args...)
	}
}

// Info logs an info message
func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	if l.verbosity >= VerbosityInfo {
		l.logger.Printf("[INFO] "+msg, args...)
	}
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	if l.verbosity >= VerbosityInfo {
		l.logger.Printf("[WARN] "+msg, args...)
	}
}

// Error logs an error message
func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+msg, args...)
}

// Trace logs a trace message (only at highest verbosity)
func (l *DefaultLogger) Trace(msg string, args ...interface{}) {
	if l.verbosity >= VerbosityTrace {
		l.logger.Printf("[TRACE] "+msg, args...)
	}
}

// LogHTTPRequest logs HTTP request details at trace level
func (l *DefaultLogger) LogHTTPRequest(method, url string, headers map[string][]string, body string) {
	if l.verbosity < VerbosityTrace {
		return
	}

	l.logger.Printf("[TRACE] HTTP Request:")
	l.logger.Printf("[TRACE]   Method: %s", method)
	l.logger.Printf("[TRACE]   URL: %s", url)
	l.logger.Printf("[TRACE]   Headers:")
	for key, values := range headers {
		for _, value := range values {
			// Redact sensitive headers
			if key == "Authorization" {
				l.logger.Printf("[TRACE]     %s: REDACTED", key)
			} else {
				l.logger.Printf("[TRACE]     %s: %s", key, value)
			}
		}
	}
	if body != "" {
		l.logger.Printf("[TRACE]   Body: %s", RedactSensitiveData(body))
	}
}

// LogHTTPResponse logs HTTP response details at trace level
func (l *DefaultLogger) LogHTTPResponse(statusCode int, status string, headers map[string][]string, body string) {
	if l.verbosity < VerbosityTrace {
		return
	}

	l.logger.Printf("[TRACE] HTTP Response:")
	l.logger.Printf("[TRACE]   Status: %d %s", statusCode, status)
	l.logger.Printf("[TRACE]   Headers:")
	for key, values := range headers {
		for _, value := range values {
			l.logger.Printf("[TRACE]     %s: %s", key, value)
		}
	}
	if body != "" {
		l.logger.Printf("[TRACE]   Body: %s", RedactSensitiveData(body))
	}
}

// GetVerbosity returns the current verbosity level
func (l *DefaultLogger) GetVerbosity() VerbosityLevel {
	return l.verbosity
}

// IsTraceEnabled returns true if trace logging is enabled
func (l *DefaultLogger) IsTraceEnabled() bool {
	return l.verbosity >= VerbosityTrace
}

// HTTPLogger is an interface for logging HTTP requests and responses
type HTTPLogger interface {
	Logger
	LogHTTPRequest(method, url string, headers map[string][]string, body string)
	LogHTTPResponse(statusCode int, status string, headers map[string][]string, body string)
	IsTraceEnabled() bool
}

// AsHTTPLogger safely converts a Logger to HTTPLogger if possible
func AsHTTPLogger(logger Logger) (HTTPLogger, bool) {
	if logger == nil {
		return nil, false
	}
	httpLogger, ok := logger.(HTTPLogger)
	return httpLogger, ok
}
