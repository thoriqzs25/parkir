package response

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorDetail provides structured error information.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// Response is the standard API envelope.
type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error interface{} `json:"error,omitempty"`
	Meta  interface{} `json:"meta,omitempty"`
}

// Meta holds pagination or other metadata.
type Meta struct {
	Page   int `json:"page,omitempty"`
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
	Total  int `json:"total,omitempty"`
}

// OK sends a 200 response with data.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Data: data})
}

// Created sends a 201 response with data.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Data: data})
}

// NoContent sends a 204 response.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error response.
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, Response{
		Error: ErrorDetail{Code: code, Message: message},
	})
}

// ValidationError sends a 422 response with field errors.
func ValidationError(c *gin.Context, errors []ErrorDetail) {
	c.JSON(http.StatusUnprocessableEntity, Response{Error: errors})
}

// BadRequest sends a 400 response.
func BadRequest(c *gin.Context, code, message string) {
	Error(c, http.StatusBadRequest, code, message)
}

// Unauthorized sends a 401 response.
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a 403 response.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound sends a 404 response.
func NotFound(c *gin.Context, resource string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", resource+" not found")
}

// Conflict sends a 409 response.
func Conflict(c *gin.Context, code, message string) {
	Error(c, http.StatusConflict, code, message)
}

// InternalServerError sends a 500 response and logs the underlying errors.
func InternalServerError(c *gin.Context) {
	for _, err := range c.Errors {
		// Log via gin's default error writer; in production this should go to structured logger
		fmt.Fprintf(gin.DefaultErrorWriter, "[ERROR] %s: %v\n", c.Request.URL.Path, err.Err)
	}
	Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "an unexpected error occurred")
}
