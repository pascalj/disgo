package models

import (
	"encoding/json"
	"github.com/martini-contrib/binding"
	"github.com/ungerik/go-gravatar"
	"net/http"
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
	TotalPages int64
	Page       int64
	PerPage    int64
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

// Validate takes a comment and sets errors if any information is missing or wrong.
func (comment Comment) Validate(errors *binding.Errors, req *http.Request) {
	if len(comment.Name) == 0 {
		errors.Fields["name"] = "Please enter a name."
	}
	if len(comment.Body) == 0 {
		errors.Fields["body"] = "You must enter a comment text."
	}
	if len(comment.Email) == 0 {
		errors.Fields["email"] = "Please enter an email address."
	}
}
