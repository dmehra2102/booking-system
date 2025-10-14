package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/dmehra2102/booking-system/internal/common/errors"
	"github.com/dmehra2102/booking-system/internal/common/logger"
	"github.com/dmehra2102/booking-system/pkg/response"
	"github.com/gin-gonic/gin"
)

func Recovery(logger *logger.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				logger.WithContext(ctx.Request.Context()).
					WithFields(map[string]any{
						"panic": err,
						"stack": string(stack),
						"path":  ctx.Request.URL.Path,
					}).
					Error("panic recovered")

				response.Error(ctx, http.StatusInternalServerError, errors.NewInternalError("internal server error", nil))
				ctx.Abort()
			}
		}()

		ctx.Next()
	}
}
