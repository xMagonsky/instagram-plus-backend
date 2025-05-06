package routes

import (
	"instagramplusbackend/internal/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type RoutesManager struct {
	pgClient    *pgxpool.Pool
	redisClient *redis.Client
	middleware  *middleware.MiddlewareManager
}

func NewRoutesManager(pgClient *pgxpool.Pool, redisClient *redis.Client, middleware *middleware.MiddlewareManager) *RoutesManager {
	return &RoutesManager{
		pgClient:    pgClient,
		redisClient: redisClient,
		middleware:  middleware,
	}
}
