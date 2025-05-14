package main

import (
	"context"
	"instagramplusbackend/internal/middleware"
	"instagramplusbackend/internal/routes"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := godotenv.Load(); err != nil {
		println("Error loading .env file: ", err)
	}

	pgClient, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	defer pgClient.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB: func() int {
			db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
			if err != nil {
				panic("invalid REDIS_DB value: " + err.Error())
			}
			return db
		}(),
	})
	defer redisClient.Close()

	r := gin.Default()

	middlewareManager := middleware.NewMiddlewareManager(pgClient, redisClient)
	//r.Use(middlewareManager.CORS())

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}                   // Allow frontend origin
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"} // Allowed HTTP methods
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"} // Allowed headers
	config.AllowCredentials = true                                            // Allow cookies or credentials if needed

	// Apply CORS middleware
	r.Use(cors.New(config))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, Backend!",
		})
	})

	routesManager := routes.NewRoutesManager(pgClient, redisClient, middlewareManager)
	routesManager.RegisterAuthRoutes(r)
	routesManager.RegisterPostsRoutes(r)
	routesManager.RegisterUserRoutes(r)

	r.Run(":5069")
}
