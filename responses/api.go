package responses

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Api represents a standard response structure for the API.
type Api struct {
	TotalRows     int64      `json:"count,omitempty"`         // Total number of rows (for pagination)
	Relationships any        `json:"relationships,omitempty"` // Related data for the response
	QueryParams   any        `json:"queryParams,omitempty"`   // Query parameters sent in the request
	Meta          any        `json:"meta,omitempty"`          // Additional metadata
	Links         any        `json:"links,omitempty"`         // Links for navigation (pagination or others)
	UserData      any        `json:"userData,omitempty"`      // Additional user-related data
	Token         string     `json:"token,omitempty"`         // Authorization token if applicable
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`     // Expiration time for the token or session
	Data          any        `json:"data"`                    // Main data returned in the response
	Message       string     `json:"message"`                 // Message describing the result
	Status        int        `json:"status"`                  // HTTP status code
	Success       bool       `json:"success"`                 // Indicates if the request was successful
}

// Error sends an error response using the provided Error object.
// It sets the HTTP status code and aborts the current request.
func (api *Api) Error(ctx *gin.Context, e Error) {
	// Load error details if not already loaded
	e.ErrorDetail.LoadDetail()

	// Default to status code 500 if not explicitly set
	if e.Code == 0 {
		e.Code = http.StatusInternalServerError
	}

	// Populate the error response details
	e.Status = e.Code
	e.Success = false

	// Send the error response as JSON and abort the request
	ctx.JSON(e.Code, e)
	ctx.Abort()
}

// OK sends a successful response.
// It sets the status code to 200 (or 204 if no data) and returns the response as JSON.
func (api *Api) OK(ctx *gin.Context, a Api) {
	// Set status to 200 or 204 if no data is present
	status := http.StatusOK
	if a.Data == nil {
		status = http.StatusNoContent
	}

	// Populate the success response details
	a.Status = status
	a.Success = true

	// Send the success response as JSON
	ctx.JSON(http.StatusOK, a)
}
