package utils

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

func LogError(c *gin.Context, err error) {
	log.Print(c.Request.URL.String() + ": " + err.Error())
}

func IsDuplicatePgxError(err error) bool {
	var pgErr *pgconn.PgError
	if err != nil && errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}
	return false
}

func IsForeignKeyViolationPgxError(err error, constraint ...string) bool {
	var pgErr *pgconn.PgError
	if err != nil && errors.As(err, &pgErr) {
		if pgErr.Code != "23503" { // foreign_key_violation
			return false
		}
		if len(constraint) == 0 {
			return true
		}
		return pgErr.ConstraintName == constraint[0]
	}
	return false
}
