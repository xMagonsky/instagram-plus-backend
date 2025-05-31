package routes

import (
	"net/http"
	"strings"

	"instagramplusbackend/internal/models"
	"instagramplusbackend/internal/utils"

	"github.com/gin-gonic/gin"
)

func (r *RoutesManager) RegisterSearchRoutes(router *gin.Engine) {
	searchRouter := router.Group("/search")
	searchRouter.Use(r.middleware.RequireAuth())
	{
		searchRouter.GET("", func(c *gin.Context) {
			query := c.Query("q")
			if strings.TrimSpace(query) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
				return
			}

			// Search users (Profile)
			userRows, err := r.pgClient.Query(c.Request.Context(), `
				SELECT u.username, p.name, p.surname, p.description, p.profile_image_url, p.gender, p.birth, u.creation_timestamp,
					(SELECT COUNT(*) FROM follows WHERE profile_id = u.id) AS followers_count,
					(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) AS following_count,
					FALSE as already_followed
				FROM users u
				JOIN user_profiles p ON u.id = p.user_id
				WHERE u.username ILIKE '%' || $1 || '%'
				   OR p.name ILIKE '%' || $1 || '%'
				   OR p.surname ILIKE '%' || $1 || '%'
				LIMIT 10`, query)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
			defer userRows.Close()
			users := []models.Profile{}
			for userRows.Next() {
				var u models.Profile
				if err := userRows.Scan(&u.Username, &u.Name, &u.Surname, &u.Description, &u.ProfileImageURL, &u.Gender, &u.BirthDate, &u.CreationTimestamp, &u.FollowersCount, &u.FollowingCount, &u.AlreadyFollowed); err != nil {
					utils.LogError(c, err)
					continue
				}
				users = append(users, u)
			}

			// Search posts (Post)
			postRows, err := r.pgClient.Query(c.Request.Context(), `
				SELECT p.id, u.username, p.image_url, p.description, p.creation_timestamp, up.name, up.surname, up.profile_image_url,
				   (SELECT COUNT(*) FROM posts_likes l WHERE l.post_id = p.id) AS likes_count,
				   (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id) AS comments_count,
				   FALSE as user_liked
				FROM posts p
				JOIN users u ON p.creator_id = u.id
				JOIN user_profiles up ON up.user_id = u.id
				WHERE p.description ILIKE '%' || $1 || '%'
				   OR u.username ILIKE '%' || $1 || '%'
				ORDER BY p.creation_timestamp DESC
				LIMIT 10`, query)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
			defer postRows.Close()
			posts := []models.Post{}
			for postRows.Next() {
				var p models.Post
				if err := postRows.Scan(&p.ID, &p.AuthorUsername, &p.ImageURL, &p.Description, &p.CreationTimestamp, &p.AuthorName, &p.AuthorSurname, &p.AuthorProfileImageURL, &p.LikesCount, &p.CommentsCount, &p.AlreadyLiked); err != nil {
					utils.LogError(c, err)
					continue
				}
				posts = append(posts, p)
			}

			c.JSON(http.StatusOK, models.SearchResponse{
				Users: users,
				Posts: posts,
			})
		})
	}
}
