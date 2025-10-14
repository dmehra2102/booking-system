package response

import (
	"net/http"

	"github.com/dmehra2102/booking-system/internal/common/errors"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success   bool       `json:"success"`
	Data      any        `json:"data,omitempty"`
	Error     *ErrorInfo `json:"error,omitempty"`
	RequestID string     `json:"request_id,omitempty"`
}

type ErrorInfo struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func Success(c *gin.Context, data any) {
	requestID, _ := c.Get("request_id")
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Data:      data,
		RequestID: requestID.(string),
	})
}

func Created(c *gin.Context, data any) {
	requestID, _ := c.Get("request_id")
	c.JSON(http.StatusCreated, Response{
		Success:   true,
		Data:      data,
		RequestID: requestID.(string),
	})
}

func Error(c *gin.Context, statusCode int, err error) {
	requestID, _ := c.Get("requedt_id")

	var errorInfo *ErrorInfo
	if appErr := errors.GetAppError(err); appErr != nil {
		errorInfo = &ErrorInfo{
			Type:    string(appErr.Type),
			Message: appErr.Message,
			Details: appErr.Details,
		}

		statusCode = appErr.Code
	} else {
		errorInfo = &ErrorInfo{
			Type:    string(errors.ErrorTypeInternal),
			Message: err.Error(),
		}
	}

	c.JSON(statusCode, Response{
		Success:   false,
		Error:     errorInfo,
		RequestID: requestID.(string),
	})
}

func ValidationError(c *gin.Context, details string) {
	Error(c, http.StatusBadRequest, errors.NewValidationError("validation failed", nil))
}

type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       any         `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
	RequestID  string      `json:"request_id,omitempty"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func Paginated(c *gin.Context, data any, pagination *Pagination) {
	requestID, _ := c.Get("request_id")
	c.JSON(http.StatusOK, PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
		RequestID:  requestID.(string),
	})
}
