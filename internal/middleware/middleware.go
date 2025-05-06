package middleware

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type MiddlewareManager struct {
	pgClient    *pgxpool.Pool
	redisClient *redis.Client
}

func NewMiddlewareManager(pgClient *pgxpool.Pool, redisClient *redis.Client) *MiddlewareManager {
	return &MiddlewareManager{
		pgClient:    pgClient,
		redisClient: redisClient,
	}
}
