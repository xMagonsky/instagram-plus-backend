package models

type ReportedUser struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id"`
	ReporterID int    `json:"reporter_id"`
	Reason     string `json:"reason"`
}

type ReportedPost struct {
	ID         int    `json:"id"`
	PostID     int    `json:"post_id"`
	ReporterID int    `json:"reporter_id"`
	Reason     string `json:"reason"`
}

type ReportedComment struct {
	ID         int    `json:"id"`
	CommentID  int    `json:"comment_id"`
	ReporterID int    `json:"reporter_id"`
	Reason     string `json:"reason"`
}

type ReportedRequest struct {
	Reason string `json:"reason" binding:"required,max=255"`
}
