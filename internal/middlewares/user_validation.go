package middlewares

import "github.com/gin-gonic/gin"

func UserValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user-uuid", "0f856bef-70d1-433e-8c28-840ab5d04b7b")
		c.Next()
	}
}
