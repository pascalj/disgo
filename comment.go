package main

import (
	"encoding/json"
	"github.com/ungerik/go-gravatar"
	"time"
)

type Comment struct {
	Id       int64 `form:"-"`
	Created  int64 `form:"-"`
	Email    string `binding:"required" form:"email"`
	Body     string `binding:"required" form:"body"`
	Url      string `binding:"required" form:"url"`
}

func NewComment(email, title, body, url string) Comment {
	return Comment{
		Created:  time.Now().Unix(),
		Email:    email,
		Body:     body,
		Url:      url,
	}
}

func (c *Comment) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{
		"id":         c.Id,
		"avatar":     gravatar.Url(c.Email),
		"body":       c.Body,
		"created_at": c.Created,
		"url":        c.Url,
	}
	return json.Marshal(data)
}
