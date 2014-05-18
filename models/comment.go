package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

func ApprovedComments(db *sql.DB, url string, email string) []Comment {
	comments := make([]Comment, 0)
	rows, err := db.Query("SELECT * FROM Comments WHERE (Approved = 1 OR Email = ?) AND Url = ?", email, url)
	if err != nil {
		logErr(err, "Could not load comments:")
		return comments
	}
	defer rows.Close()
	for rows.Next() {
		comment := Comment{}
		err := rows.Scan(
			&comment.Id,
			&comment.Created,
			&comment.Email,
			&comment.Name,
			&comment.Body,
			&comment.Url,
			&comment.ClientIp,
			&comment.Approved)
		if err != nil {
			logErr(err, "Error mapping comment:")
		}
		comments = append(comments, comment)
	}
	err = rows.Err()
	if err != nil {
		logErr(err, "Could not read comments")
	}
	return comments
}

func AllComments(db *sql.DB, url string) []Comment {
	comments := make([]Comment, 0)
	rows, err := db.Query("SELECT * FROM Comments WHERE Url = ?", url)
	if err != nil {
		logErr(err, "Could not load comments:")
		return comments
	}
	defer rows.Close()
	for rows.Next() {
		comment := Comment{}
		err := rows.Scan(
			&comment.Id,
			&comment.Created,
			&comment.Email,
			&comment.Name,
			&comment.Body,
			&comment.Url,
			&comment.ClientIp,
			&comment.Approved)
		if err != nil {
			logErr(err, "Error mapping comment:")
		}
		comments = append(comments, comment)
	}
	err = rows.Err()
	if err != nil {
		logErr(err, "Could not read comments")
	}
	return comments
}

func logErr(err error, description string) {
	fmt.Println(description, err)
}
