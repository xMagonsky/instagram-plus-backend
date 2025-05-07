package routes

import (
	"net/http"

	"instagramplusbackend/internal/models"

	"github.com/gin-gonic/gin"
)

func (r *RoutesManager) RegisterAuthRoutes(router *gin.Engine) {
	authRouter := router.Group("/auth")
	{
		authRouter.POST("/register", func(c *gin.Context) {
			var req models.RegisterRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}

			token, err := r.auth.Register(c.Request.Context(), req.Username, req.Password, req.Email)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.SetCookie("AUTH", token, 0, "/", "", false, true)

			c.JSON(http.StatusOK, gin.H{})
		})

		authRouter.POST("/login", func(c *gin.Context) {
			var req models.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}

			token, err := r.auth.Login(c.Request.Context(), req.Username, req.Password)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}

			c.SetCookie("AUTH", token, 0, "/", "", false, true)

			c.JSON(http.StatusOK, gin.H{})
		})

		authRouter.POST("/validate", func(c *gin.Context) {
			var req models.TokenRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}

			userID, err := r.auth.ValidateToken(c.Request.Context(), req.Token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		})

		authRouter.POST("/logout", func(c *gin.Context) {
			var req models.TokenRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}

			if err := r.auth.Logout(c.Request.Context(), req.Token); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{})
		})
	}
}
