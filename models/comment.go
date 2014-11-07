package models

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/ungerik/go-gravatar"
	"net/mail"
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
	ClientIp string `form:"-" db:"clientIp"`
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
func NewComment(email, name, title, body, url, ip string) Comment {
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

func (c *Comment) Validate() (bool, map[string]string) {
	errors := map[string]string{}
	if _, err := mail.ParseAddress(c.Email); err != nil {
		errors["email"] = "A valid Email address is required."
	}
	if c.Name == "" {
		errors["name"] = "Name is required."
	}
	if c.Body == "" {
		errors["body"] = "A message is required."
	}
	if c.Url == "" {
		errors["url"] = "A URL is required. Something went wrong."
	}
	return (len(errors) == 0), errors
}

func (c *Comment) Save(db *sqlx.DB) error {
	var err error
	if c.Id != 0 {
		_, err = db.Exec(`
			UPDATE comments SET
			Email = ?, Name = ?, Body = ?, Created = ?, Url = ?, ClientIp = ?, Approved = ? WHERE Id = ?`,
			c.Email, c.Name, c.Body, c.Created, c.Url, c.ClientIp, c.Approved, c.Id)
		if err != nil {
			return err
		}
	} else {
		res, err := db.Exec(`
			INSERT INTO
			comments(Email, Name, Body, Created, Url, ClientIp, Approved)
			VALUES(?, ?, ?, ?, ?, ?, ?)`,
			c.Email, c.Name, c.Body, c.Created, c.Url, c.ClientIp, c.Approved)
		if err != nil {
			return err
		}

		lastId, err := res.LastInsertId()
		if err != nil {
			return err
		}
		c.Id = lastId
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *Comment) Delete(db *sqlx.DB) error {
	_, err := db.Exec(`
		DELETE FROM
		comments
		WHERE id = ?`, c.Id)
	if err != nil {
		return err
	}
	return nil
}

func GetComment(db *sqlx.DB, id int) *Comment {
	comment := &Comment{}
	err := db.Get(comment, "SELECT * FROM comments WHERE Id = ?", id)
	if err != nil {
		return nil
	}
	return comment
}

func UnapprovedComments(db *sqlx.DB) []Comment {
	comments := []Comment{}
	err := db.Select(&comments, "SELECT * FROM comments WHERE (Approved = 0)")
	if err != nil {
		logErr(err, "Could not load comments:")
		return nil
	}
	return comments
}

func ApprovedComments(db *sqlx.DB, url string, email string) []Comment {
	comments := []Comment{}
	err := db.Select(&comments, "SELECT * FROM comments WHERE (Approved = 1 OR Email = ?) AND Url = ?", email, url)
	if err != nil {
		logErr(err, "Could not load comments:")
		return nil
	}
	return comments
}

func UnapprovedCommentsCount(db *sqlx.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT * FROM comments WHERE Approved<>1").Scan(&count)
	return count, err
}

func AllComments(db *sqlx.DB, url string) []Comment {
	comments := []Comment{}
	err := db.Get(&comments, "SELECT * FROM comments WHERE Url = ?", url)
	if err != nil {
		logErr(err, "Could not load comments:")
	}
	return comments
}

func AllCommentsPaginated(db *sqlx.DB, page int) ([]Comment, int) {
	comments := []Comment{}
	err := db.Select(&comments, "SELECT * FROM comments ORDER BY Created DESC LIMIT 10 OFFSET ?", page*10)
	if err != nil {
		logErr(err, "Could not load comments:")
		return nil, 0
	}

	row := db.QueryRow("SELECT COUNT(*)/10 FROM comments")
	var count int
	err = row.Scan(&count)
	if err != nil {
		logErr(err, "Could not find comment pages count:")
		return comments, 0
	}

	if err != nil {
		logErr(err, "Could not find comment pages count:")
		return comments, 0
	}
	return comments, count
}

func logErr(err error, description string) {
	fmt.Println(description, err)
}
