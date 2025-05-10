package middleware

import (
	"instagramplusbackend/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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

func (m *MiddlewareManager) RequireUserOwnership(userParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		paramUserID := c.Param(userParam)
		if paramUserID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			c.Abort()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		paramUserIDInt, err := strconv.Atoi(paramUserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			c.Abort()
			return
		}

		if userID.(int) != paramUserIDInt {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *MiddlewareManager) RequirePostOwnership(postParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		paramPostID := c.Param(postParam)
		if paramPostID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
			c.Abort()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var ownerID int
		err := m.pgClient.QueryRow(c.Request.Context(), `SELECT creator_id FROM posts WHERE id = $1`, paramPostID).Scan(&ownerID)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
				c.Abort()
				return
			}
			utils.LogError(c, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			c.Abort()
			return
		}

		if ownerID != userID.(int) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}
