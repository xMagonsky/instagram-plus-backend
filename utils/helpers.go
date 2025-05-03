package utils

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, err error, error_msg string) bool {
	if err != nil {
		log.Print(c.Request.URL.String() + ": " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": error_msg})
		return true
	}
	return false
}
