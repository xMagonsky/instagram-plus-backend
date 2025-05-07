package utils

import (
	"log"

	"github.com/gin-gonic/gin"
)

func LogError(c *gin.Context, err error) {
	log.Print(c.Request.URL.String() + ": " + err.Error())
}
