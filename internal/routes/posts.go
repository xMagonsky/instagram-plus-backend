package routes

import (
	"net/http"

	"instagramplusbackend/internal/models"
	"instagramplusbackend/internal/utils"

	"github.com/gin-gonic/gin"
)

func (r *RoutesManager) RegisterPostsRoutes(router *gin.Engine) {
	postRouter := router.Group("/posts")
	postRouter.Use(r.middleware.RequireAuth())
	{
		postRouter.GET("/all", func(c *gin.Context) {
			rows, err := r.pgClient.Query(c.Request.Context(), `
				SELECT p.id, p.photo_url, p.description, p.create_timestamp, p.creator_id, u.username
				FROM posts p
				JOIN users u ON p.creator_id = u.id
				ORDER BY p.create_timestamp ASC`)
			if utils.HandleError(c, err, "database error") {
				return
			}
			defer rows.Close()

			posts := []models.Post{}
			for rows.Next() {
				var post models.Post
				err := rows.Scan(&post.ID, &post.PhotoURL, &post.Description, &post.CreateTimestamp, &post.CreatorID, &post.Author)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
					return
				}
				posts = append(posts, post)
			}

			c.JSON(http.StatusOK, posts)
		})

		postRouter.POST("/", func(c *gin.Context) {
			var req models.AddPostRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}

			_, err := r.pgClient.Exec(c.Request.Context(),
				"INSERT INTO posts (photo_url, description, creator_id) VALUES ($1, $2, $3)",
				req.Image, req.Content, 3)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}

			c.JSON(http.StatusOK, gin.H{})
		})
	}
}
