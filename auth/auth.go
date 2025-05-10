package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthModule struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewAuthModule(db *pgxpool.Pool, redis *redis.Client) *AuthModule {
	return &AuthModule{
		db:    db,
		redis: redis,
	}
}

func generateSecureToken(length int) (string, error) {
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(randomBytes), nil
}

func (a *AuthModule) Register(ctx context.Context, username, password string, email string) (int, string, error) {
	var exists bool
	err := a.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	if err != nil {
		return 0, "", err
	}
	if exists {
		return 0, "", errors.New("username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, "", err
	}

	var userID int
	err = a.db.QueryRow(ctx,
		"INSERT INTO users (username, password, email) VALUES ($1, $2, $3) RETURNING id",
		username, string(hashedPassword), email,
	).Scan(&userID)
	if err != nil {
		return 0, "", err
	}

	token, err := generateSecureToken(32)
	if err != nil {
		return 0, "", err
	}

	key := "session:" + token
	err = a.redis.Set(ctx, key, userID, 24*time.Hour).Err()
	if err != nil {
		return 0, "", err
	}

	return userID, token, nil
}

func (a *AuthModule) Login(ctx context.Context, username, password string) (string, error) {
	var userID int
	var passwordHash string
	err := a.db.QueryRow(ctx, "SELECT id, password FROM users WHERE username = $1", username).Scan(&userID, &passwordHash)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token, err := generateSecureToken(32)
	if err != nil {
		return "", err
	}

	key := "session:" + token
	err = a.redis.Set(ctx, key, userID, 24*time.Hour).Err()
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *AuthModule) ValidateToken(ctx context.Context, token string) (string, error) {
	key := "session:" + token
	userID, err := a.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", errors.New("invalid token")
	} else if err != nil {
		return "", err
	}

	// Check the expiration time of the token
	ttl, err := a.redis.TTL(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return "", err
	}

	// Update expiration only after some time
	if ttl < 20*time.Hour {
		err = a.redis.Expire(ctx, key, 24*time.Hour).Err()
		if err != nil {
			return "", err
		}
	}
	return userID, nil
}

func (a *AuthModule) Logout(ctx context.Context, token string) error {
	key := "session:" + token
	return a.redis.Del(ctx, key).Err()
}

//
// TOKEN SIGNING FOR HASHED TOKENS IN REDIS
//
// func signToken(token, secretKey string) string {
// 	h := hmac.New(sha256.New, []byte(secretKey))
// 	h.Write([]byte(token))
// 	signature := h.Sum(nil)

// 	return base64.URLEncoding.EncodeToString(signature)
// }
