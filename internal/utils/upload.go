package utils

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

var nginxPath = "c://nginx/"

var postImagePathPrefix = "/images/posts/"
var profileImagePathPrefix = "/images/profiles/"

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

	imageURL := postImagePathPrefix + fileName

	return imageURL, nil
}

func RemovePostImage(imagePath string) error {
	uploadPath := filepath.Join(nginxPath, "data/posts", filepath.Base(imagePath))

	if err := os.Remove(uploadPath); err != nil {
		return err
	}

	return nil
}

func UploadProfileImage(c *gin.Context) (string, error) {
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // Limit to 10MB
		return "", err
	}

	file, err := c.FormFile("image")
	if err != nil {
		return "", err
	}

	uploadPath := filepath.Join(nginxPath, "data/profiles")
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

	imageURL := profileImagePathPrefix + fileName

	return imageURL, nil
}

func RemoveProfileImage(imagePath string) error {
	uploadPath := filepath.Join(nginxPath, "data/profiles", filepath.Base(imagePath))

	if err := os.Remove(uploadPath); err != nil {
		return err
	}

	return nil
}
