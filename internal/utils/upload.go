package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

var nginxPath = "c://nginx/"

func UploadPostImage(c *gin.Context) (string, error) {
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // Limit to 10MB
		return "", err
	}

	file, err := c.FormFile("image")
	if err != nil {
		return "", err
	}

	uploadPath := filepath.Join(nginxPath, "data/posts")
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		return "", err
	}

	randomBytes := make([]byte, 16)
	if _, err = rand.Read(randomBytes); err != nil {
		return "", err
	}

	fileName := hex.EncodeToString(randomBytes) + filepath.Ext(file.Filename)

	dst := filepath.Join(uploadPath, fileName)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return "", err
	}

	imageURL := fmt.Sprintf("/images/posts/%s", fileName)

	return imageURL, nil
}
