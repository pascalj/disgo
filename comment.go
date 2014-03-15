package main

import (
	"encoding/json"
	"github.com/ungerik/go-gravatar"
	"time"
)

type Comment struct {
	Id       int64 `json:"id"`
	Created  int64
	Email    string
	Title    string
	Body     string
	Url      string
	ParentId int64
}

func NewComment(email, title, body, url, parentId string) Comment {
	return Comment{
		Created:  time.Now().Unix(),
		Email:    email,
		Title:    title,
		Body:     body,
		Url:      url,
		ParentId: 0,
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
