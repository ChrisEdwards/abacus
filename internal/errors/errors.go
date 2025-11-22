package errors

import "errors"

// Code identifies a structured error type used across the application.
type Code string

const (
	// Generic codes
	CodeUnknown Code = "unknown"

	// CLI errors
	CodeCLINotFound Code = "cli_not_found"
	CodeCLIFailed   Code = "cli_failed"
	CodeParseFailed Code = "parse_failed"
	CodeNotFound    Code = "not_found"

	// Domain/graph errors
	CodeInvalidTransition  Code = "invalid_transition"
	CodeInvalidPriority    Code = "invalid_priority"
	CodeInvalidStatus      Code = "invalid_status"
	CodeCyclicDependency   Code = "cyclic_dependency"
	CodeInvalidIssueData   Code = "invalid_issue_data"
	CodeGraphConstruction  Code = "graph_construction_failed"
	CodeConfigurationError Code = "configuration_error"
)

// Error represents a structured error with a machine-readable code plus message.
type Error struct {
	Code    Code
	Message string
	Err     error
}

// Error implements the error interface.
func (e Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

// Unwrap returns the wrapped error.
func (e Error) Unwrap() error {
	return e.Err
}

// New wraps an error with a code/message.
func New(code Code, msg string, err error) Error {
	return Error{Code: code, Message: msg, Err: err}
}

// CodeOf walks the error chain and returns the first structured code found.
func CodeOf(err error) Code {
	var structured Error
	if errors.As(err, &structured) {
		return structured.Code
	}
	return CodeUnknown
}

// IsCode reports whether the error (or its unwrap chain) matches the provided code.
func IsCode(err error, code Code) bool {
	return CodeOf(err) == code
}
