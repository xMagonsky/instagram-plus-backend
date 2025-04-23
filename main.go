package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"instagramplusbackend/auth"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type AddPostRequest struct {
	Author  string `json:"author" binding:"required"`
	Image   string `json:"image" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		println("Error loading .env file: ", err)
	}

	postgreClient, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	defer postgreClient.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       1,
	})
	defer redisClient.Close()

	auth := auth.NewAuthModule(postgreClient, redisClient)

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}                   // Allow frontend origin
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"} // Allowed HTTP methods
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"} // Allowed headers
	config.AllowCredentials = true                                            // Allow cookies or credentials if needed
	r.Use(cors.New(config))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, Backend!",
		})
	})

	r.POST("/register", func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		ctx := context.Background()
		if _, err := auth.Register(ctx, req.Username, req.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//
		// LOGIN!
		//

		c.JSON(http.StatusOK, gin.H{"message": "user registered successfully"})
	})

	r.POST("/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		ctx := context.Background()
		token, err := auth.Login(ctx, req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		println("Number of cookies: ", len(c.Request.Cookies()))
		c.SetCookie("session_token", token, 3600, "/", "localhost:5173", false, true)

		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	r.POST("/validate", func(c *gin.Context) {
		var req TokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		ctx := context.Background()
		userID, err := auth.ValidateToken(ctx, req.Token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	r.POST("/logout", func(c *gin.Context) {
		var req TokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		ctx := context.Background()
		if err := auth.Logout(ctx, req.Token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
	})

	r.GET("/post/all", func(c *gin.Context) {
		c.JSON(http.StatusOK, []gin.H{
			{
				"author":  "Test Author1",
				"image":   "https://picsum.photos/500",
				"content": "Test content of a post.2",
			},
			{
				"author":  "Test Author2",
				"image":   "https://picsum.photos/500",
				"content": "Test content of a post.2",
			},
		})
	})

	r.POST("/post", func(c *gin.Context) {
		var req AddPostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.Run(":5069")
}
