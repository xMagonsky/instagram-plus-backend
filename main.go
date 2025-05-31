package main

import (
	"context"
	"instagramplusbackend/internal/middleware"
	"instagramplusbackend/internal/routes"
	"net/http"
	"os"
	"strconv"

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
	r.RedirectTrailingSlash = false

	middlewareManager := middleware.NewMiddlewareManager(pgClient, redisClient)
	r.Use(middlewareManager.CORS())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, Backend!",
		})
	})

	routesManager := routes.NewRoutesManager(pgClient, redisClient, middlewareManager)
	routesManager.RegisterAuthRoutes(r)
	routesManager.RegisterPostsRoutes(r)
	routesManager.RegisterUserRoutes(r)
	routesManager.RegisterCommentsRoutes(r)
	routesManager.RegisterSearchRoutes(r)
	routesManager.RegisterReportsRoutes(r)

	r.Run(":5069")
}
