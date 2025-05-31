package models

type SearchResponse struct {
	Users []Profile `json:"users"`
	Posts []Post    `json:"posts"`
}
