package errors

import "fmt"

// ErrorType represents the type of error that occurred
type ErrorType string

const (
	// Error types
	ErrTypeDatabase   ErrorType = "database_error"
	ErrTypeCommand    ErrorType = "command_error"
	ErrTypeValidation ErrorType = "validation_error"
	ErrTypeBot        ErrorType = "bot_error"
	ErrTypeEmail      ErrorType = "email_error"
	ErrTypeFeature    ErrorType = "feature_error"
	ErrTypeUnknown    ErrorType = "unknown_error"
)

// BotError is a custom error type for the bot
type BotError struct {
	Type    ErrorType
	Message string
	Err     error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *BotError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *BotError) Unwrap() error {
	return e.Err
}

// WithContext adds context to the error
func (e *BotError) WithContext(key string, value interface{}) *BotError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewDatabaseError creates a new database error
func NewDatabaseError(message string, err error) *BotError {
	return &BotError{
		Type:    ErrTypeDatabase,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// NewCommandError creates a new command error
func NewCommandError(message string, err error) *BotError {
	return &BotError{
		Type:    ErrTypeCommand,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string, err error) *BotError {
	return &BotError{
		Type:    ErrTypeValidation,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// NewBotError creates a new bot error
func NewBotError(message string, err error) *BotError {
	return &BotError{
		Type:    ErrTypeBot,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// NewEmailError creates a new email error
func NewEmailError(message string, err error) *BotError {
	return &BotError{
		Type:    ErrTypeEmail,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// NewFeatureError creates a new feature error
func NewFeatureError(message string, err error) *BotError {
	return &BotError{
		Type:    ErrTypeFeature,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}