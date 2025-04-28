package utils

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"runtime/debug"
)

// SendErrorBody sends a structured error response with code, message, and the corresponding HTTP status
func SendErrorBody(c *gin.Context, code int, body interface{}, err error) {
	errorDetails, exists := ErrorMessages[code]
	if !exists {
		errorDetails = ErrorDetails{
			HTTPStatus: http.StatusInternalServerError,
			Message:    "Unknown error: " + err.Error(),
		}
		log.Printf("Unknown error code: %d\nError details: %v\nStack trace: %s", code, err, debug.Stack())
	}

	log.Printf("Error %d: %s\n%s", code, errorDetails.Message, err.Error())
	sendResponse(c, errorDetails.HTTPStatus, code, errorDetails.Message, body)
	c.Abort()
}

// SendError sends a structured error response without a body
func SendError(c *gin.Context, code int, err error) {
	SendErrorBody(c, code, nil, err)
}

// SendSuccessBody sends a successful 200 response with a body
func SendSuccessBody(c *gin.Context, body interface{}) {
	sendResponse(c, http.StatusOK, SUCCESS, ErrorMessages[SUCCESS].Message, body)
}

// SendSuccess sends a successful 200 response without a body
func SendSuccess(c *gin.Context) {
	SendSuccessBody(c, nil)
}

// Helper function to send JSON responses
func sendResponse(c *gin.Context, httpStatus int, code int, message string, body interface{}) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Body:    body,
	})
}
