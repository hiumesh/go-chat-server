package utils

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HTTPError struct {
	Code            int    `json:"code"`
	Message         string `json:"msg"`
	InternalError   error  `json:"-"`
	InternalMessage string `json:"-"`
	ErrorID         string `json:"error_id,omitempty"`
}

func (e *HTTPError) Error() string {
	if e.InternalMessage != "" {
		return e.InternalMessage
	}
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func (e *HTTPError) Is(target error) bool {
	return e.Error() == target.Error()
}

// Cause returns the root cause error
func (e *HTTPError) Cause() error {
	if e.InternalError != nil {
		return e.InternalError
	}
	return e
}

func HandleHttpError(err *HTTPError, c *gin.Context) {
	c.AbortWithStatusJSON(err.Code, err)
}

// WithInternalError adds internal error information to the error
func (e *HTTPError) WithInternalError(err error) *HTTPError {
	e.InternalError = err
	return e
}

// WithInternalMessage adds internal message information to the error
func (e *HTTPError) WithInternalMessage(fmtString string, args ...interface{}) *HTTPError {
	e.InternalMessage = fmt.Sprintf(fmtString, args...)
	return e
}

func httpError(code int, fmtString string, args ...interface{}) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: fmt.Sprintf(fmtString, args...),
	}
}

func UnauthorizedError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusUnauthorized, fmtString, args...)
}

func BadRequestError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusBadRequest, fmtString, args...)
}

func InternalServerError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusInternalServerError, fmtString, args...)
}

func NotFoundError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusNotFound, fmtString, args...)
}

func ExpiredTokenError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusUnauthorized, fmtString, args...)
}

func ForbiddenError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusForbidden, fmtString, args...)
}

func UnprocessableEntityError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusUnprocessableEntity, fmtString, args...)
}

func TooManyRequestsError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusTooManyRequests, fmtString, args...)
}

func ConflictError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusConflict, fmtString, args...)
}
