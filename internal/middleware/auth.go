package middleware

import (
	"instagramplusbackend/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (m *MiddlewareManager) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("AUTH")
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var userID string
		userID, err = m.auth.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userIDInt, err := strconv.Atoi(userID)
		if err != nil {
			utils.LogError(c, err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Set("user_id", userIDInt)

		c.Next()
	}
}
