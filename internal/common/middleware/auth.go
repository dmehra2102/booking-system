package middleware

import (
	"net/http"
	"strings"

	"github.com/dmehra2102/booking-system/internal/common/errors"
	"github.com/dmehra2102/booking-system/pkg/auth"
	"github.com/dmehra2102/booking-system/pkg/response"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(ctx, http.StatusUnauthorized, errors.NewUnauthorizedError("missing authorization header"))
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if  tokenString == authHeader {
			response.Error(ctx, http.StatusUnauthorized, errors.NewUnauthorizedError("invalid authorization format"))
			ctx.Abort()
			return
		}

		claims,err := auth.ValidateToken(tokenString,jwtSecret)
		if err != nil {
			response.Error(ctx, http.StatusUnauthorized, errors.NewUnauthorizedError("invalid token"))
			ctx.Abort()
			return
		}

		ctx.Set("user_id", claims.UserID)
		ctx.Set("user_email", claims.Email)
		ctx.Next()
	}
}

func OptionalAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString != authHeader {
			claims, err := auth.ValidateToken(tokenString, jwtSecret)
			if err == nil {
				c.Set("user_id", claims.UserID)
				c.Set("user_email", claims.Email)
			}
		}

		c.Next()
	}
}