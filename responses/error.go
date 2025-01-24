package responses

import (
	"errors"

	"gorm.io/gorm"
)

// Constants for error types
const (
	TypeDB      = "DB"      // Represents database-related errors
	TypeBind    = "BIND"    // Represents binding-related errors
	TypeUnknown = "UNKNOWN" // Represents unknown error types
)

// Struct representing detailed error information
type ErrorDetail struct {
	Error   error  `json:"-"`                 // Actual error (not serialized in JSON)
	Type    string `json:"type,omitempty"`    // Type of error (e.g., DB, BIND, UNKNOWN)
	Message string `json:"message,omitempty"` // Error message for the client
	Details string `json:"details,omitempty"` // Translated or detailed error message
}

// Struct representing a complete error response
type Error struct {
	ErrorDetail ErrorDetail `json:"error"`   // Detailed error information
	Status      int         `json:"status"`  // HTTP status code
	Code        int         `json:"-"`       // Internal HTTP code (not serialized in JSON)
	Success     bool        `json:"success"` // Indicates whether the request was successful
}

// Method to load or translate the error details if not already defined
func (e *ErrorDetail) LoadDetail() {
	if e.Details == "" {
		var translatedError string
		if e.Error != nil {
			switch e.Type {
			case TypeDB:
				// Translate database-related errors
				translatedError = TranslateDBError(e.Error)
			case TypeBind:
				// Translate binding-related errors
				translatedError = TranslateBindError(e.Error)
			default:
				// Use the default error message
				translatedError = e.Error.Error()
			}
		} else {
			// Default message when no specific error is defined
			translatedError = "Undefined error"
		}
		e.Details = translatedError
	}

	// If the type is neither DB nor BIND, set it to UNKNOWN
	if e.Type != TypeDB && e.Type != TypeBind {
		e.Type = TypeUnknown
	}
}

// Function to translate database-related errors
func TranslateDBError(err error) string {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "Record not found"
	} else if errors.Is(err, gorm.ErrInvalidTransaction) {
		return "Invalid transaction"
	} else if errors.Is(err, gorm.ErrNotImplemented) {
		return "Function not implemented"
	} else if errors.Is(err, gorm.ErrMissingWhereClause) {
		return "Missing WHERE clause"
	} else if errors.Is(err, gorm.ErrUnsupportedRelation) {
		return "Unsupported relation"
	} else if errors.Is(err, gorm.ErrPrimaryKeyRequired) {
		return "Primary key required"
	} else if errors.Is(err, gorm.ErrModelValueRequired) {
		return "Model value required"
	} else if errors.Is(err, gorm.ErrInvalidData) {
		return "Invalid data"
	} else if errors.Is(err, gorm.ErrDuplicatedKey) {
		return "Duplicated key"
	}
	// Default to returning the original error message
	return err.Error()
}

// Function to translate binding-related errors
func TranslateBindError(err error) string {
	// Custom logic for translating binding-related errors can be added here
	// Example:
	// if errGin, ok := err.(gin.Error); !ok {
	// }
	return err.Error()
}
