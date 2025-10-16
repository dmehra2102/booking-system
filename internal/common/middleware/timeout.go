package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/dmehra2102/booking-system/internal/common/errors"
	"github.com/dmehra2102/booking-system/pkg/response"
	"github.com/gin-gonic/gin"
)

func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})
		go func() {
			defer func() {
				if err := recover(); err != nil {
					// TODO : Handle panic
				}
			}()
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				response.Error(c, http.StatusGatewayTimeout,
					errors.NewInternalError("request timeout", ctx.Err()))
				c.Abort()
			}
		}
	}
}
