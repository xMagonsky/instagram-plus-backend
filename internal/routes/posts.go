package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"instagramplusbackend/internal/models"
	"instagramplusbackend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (r *RoutesManager) RegisterPostsRoutes(router *gin.Engine) {
	postRouter := router.Group("/posts")
	postRouter.Use(r.middleware.RequireAuth())
	{
		postRouter.GET("", func(c *gin.Context) {

			rows, err := r.pgClient.Query(c.Request.Context(), `
				SELECT p.id, p.creator_id, p.image_url, p.description, p.creation_timestamp, u.username,
					   (SELECT COUNT(*) FROM posts_likes l WHERE l.post_id = p.id) AS likes_count
				FROM posts p
				JOIN users u ON p.creator_id = u.id
				ORDER BY p.creation_timestamp DESC`)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
			defer rows.Close()

			posts := []models.Post{}
			for rows.Next() {
				var post models.Post
				err := rows.Scan(&post.ID, &post.AuthorID, &post.ImageURL, &post.Description, &post.CreationTimestamp, &post.AuthorName, &post.LikesCount)
				if err != nil {
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
					return
				}
				posts = append(posts, post)
			}

			c.JSON(http.StatusOK, posts)
		})

		postRouter.POST("", func(c *gin.Context) {
			var req models.AddPostRequest
			data := c.Request.FormValue("data")
			if err := json.Unmarshal([]byte(data), &req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}

			imageURL, err := utils.UploadPostImage(c)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upload image"})
				return
			}

			_, err = r.pgClient.Exec(c.Request.Context(),
				"INSERT INTO posts (image_url, description, creator_id) VALUES ($1, $2, $3)",
				imageURL, req.Description, c.GetInt("user_id"))
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}

			c.JSON(http.StatusOK, gin.H{})
		})

		postRouter.GET("/:post_id", func(c *gin.Context) {
			postID := c.Param("post_id")
			if postID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
				return
			}

			row := r.pgClient.QueryRow(c.Request.Context(), `
				SELECT p.id, p.creator_id, p.image_url, p.description, p.creation_timestamp, u.username,
					(SELECT COUNT(*) FROM posts_likes l WHERE l.post_id = p.id) AS likes_count
				FROM posts p
				JOIN users u ON p.creator_id = u.id
				WHERE p.id = $1`, postID)

			var post models.Post
			err := row.Scan(&post.ID, &post.AuthorID, &post.ImageURL, &post.Description, &post.CreationTimestamp, &post.AuthorName, &post.LikesCount)
			if err != nil {
				if err == pgx.ErrNoRows {
					c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
					return
				}
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}

			c.JSON(http.StatusOK, post)
		})

		postRouter.DELETE("/:post_id", r.middleware.RequirePostOwnership("post_id"), func(c *gin.Context) {
			postID := c.Param("post_id")

			_, err := r.pgClient.Exec(c.Request.Context(), "DELETE FROM posts WHERE id = $1", postID)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}

			c.JSON(http.StatusOK, gin.H{})
		})

		postRouter.PATCH("/:post_id", r.middleware.RequirePostOwnership("post_id"), func(c *gin.Context) {
			postID := c.Param("post_id")

			var req models.UpdatePostRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}

			_, err := r.pgClient.Exec(c.Request.Context(),
				"UPDATE posts SET description = $1 WHERE id = $2",
				req.Description, postID)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}

			c.JSON(http.StatusOK, gin.H{})
		})

		postRouter.POST("/:post_id/like", func(c *gin.Context) {
			postID, err := strconv.Atoi(c.Param("post_id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
				return
			}

			likerID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
				return
			}

			_, err = r.pgClient.Exec(c.Request.Context(),
				"INSERT INTO posts_likes (post_id, user_id) VALUES ($1, $2)",
				postID, likerID)
			if err != nil {
				if utils.IsDuplicatePgxError(err) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "you already liked this post"})
					return
				}
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}

			c.JSON(http.StatusOK, gin.H{})
		})

		postRouter.DELETE("/:post_id/like", func(c *gin.Context) {
			postID, err := strconv.Atoi(c.Param("post_id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
				return
			}

			unlikerID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
				return
			}

			_, err = r.pgClient.Exec(c.Request.Context(),
				"DELETE FROM posts_likes WHERE post_id = $1 AND user_id = $2",
				postID, unlikerID)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}

			c.JSON(http.StatusOK, gin.H{})
		})
	}
}
