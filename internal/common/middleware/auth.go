package middleware

import "github.com/gin-gonic/gin"

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			// response
		}
	}
}
