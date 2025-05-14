package routes

import (
	"net/http"

	"instagramplusbackend/internal/models"
	"instagramplusbackend/internal/utils"

	"github.com/gin-gonic/gin"
)

func (r *RoutesManager) RegisterAuthRoutes(router *gin.Engine) {
	authRouter := router.Group("/auth")
	{
		authRouter.POST("/register", func(c *gin.Context) {
			var req models.RegisterRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				println(err.Error())
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}

			userID, token, err := r.auth.Register(c.Request.Context(), req.Username, req.Password, req.Email)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			_, err = r.pgClient.Exec(c.Request.Context(), `
				INSERT INTO user_profiles (user_id, name, surname, description, profile_image_url, gender, birth)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				userID, req.Name, req.Surname, req.Description, req.ProfileImage, req.Gender, req.BirthDate)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user profile"})
				return
			}

			c.SetCookie("AUTH", token, 0, "/", "", false, true)

			c.JSON(http.StatusOK, gin.H{"username": req.Username})
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

			c.JSON(http.StatusOK, gin.H{"username": req.Username})
		})

		authRouter.POST("/validate", func(c *gin.Context) {
			var req models.TokenRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}

			_, err := r.auth.ValidateToken(c.Request.Context(), req.Token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{})
		})

		authRouter.POST("/logout", func(c *gin.Context) {
			token, err := c.Cookie("AUTH")
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged in"})
				return
			}

			if err := r.auth.Logout(c.Request.Context(), token); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.SetCookie("AUTH", "", -1, "/", "", false, true)

			c.JSON(http.StatusOK, gin.H{})
		})
	}
}
