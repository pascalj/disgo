package main

import (
	"time"
)

type Comment struct {
	Id       int64
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
