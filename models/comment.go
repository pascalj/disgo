package models

import (
	"encoding/json"
	"github.com/ungerik/go-gravatar"
	"time"
)

// Comment represents a single comment with all its associated data.
type Comment struct {
	Id       int64  `form:"-"`
	Created  int64  `form:"-"`
	Email    string `binding:"required" form:"email"`
	Name     string `binding:"required" form:"name"`
	Body     string `binding:"required" form:"body"`
	Url      string `binding:"required" form:"url"`
	ClientIp string `form:"-"`
	ClientId string `form:"-"`
	Approved bool   `form:"-"`
}

// PaginatedComments helps to easily divide a set of Comments into TotalPages/PerPage slices.
type PaginatedComments struct {
	TotalPages int
	Page       int
	PerPage    int
	Comments   []Comment
}

// NewComment creates a comment and sets the current time as its Created date.
func NewComment(email, name, title, body, url, ip, id string) Comment {
	return Comment{
		Created:  time.Now().Unix(),
		Email:    email,
		Name:     name,
		Body:     body,
		Url:      url,
		ClientIp: ip,
		ClientId: id,
		Approved: false,
	}
}

// MarshalJSON implements the Marshal interface to serialize the comment.
func (c *Comment) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{
		"id":         c.Id,
		"avatar":     gravatar.Url(c.Email),
		"name":       c.Name,
		"body":       c.Body,
		"created_at": c.Created,
		"url":        c.Url,
		"approved":   c.Approved,
	}
	return json.Marshal(data)
}
