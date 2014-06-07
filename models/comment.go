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
	if c.Email == "" {
		errors["email"] = "Email address is required."
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

func (c *Comment) Save(db *sql.DB) error {
	var stmt *sql.Stmt
	var err error
	if c.Id != 0 {
		stmt, err = db.Prepare(`
			UPDATE comments SET
			Email = ?, Name = ?, Body = ?, Created = ?, Url = ?, ClientIp = ?, Approved = ? WHERE Id = ?`)
		if err != nil {
			return err
		}
		_, err = stmt.Exec(c.Email, c.Name, c.Body, c.Created, c.Url, c.ClientIp, c.Approved, c.Id)
	} else {
		stmt, err = db.Prepare(`
			INSERT INTO
			comments(Email, Name, Body, Created, Url, ClientIp, Approved)
			VALUES(?, ?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		res, err := stmt.Exec(c.Email, c.Name, c.Body, c.Created, c.Url, c.ClientIp, c.Approved)
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

func (c *Comment) Delete(db *sql.DB) error {
	stmt, err := db.Prepare(`
		DELETE FROM
		comments
		WHERE id = ?`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(c.Id)
	return err
}

func GetComment(db *sql.DB, id int) *Comment {
	row := db.QueryRow("SELECT * FROM Comments WHERE Id = ?", id)
	comment, err := scanComment(row)
	if err != nil {
		return nil
	} else {
		return comment
	}
}

func UnapprovedComments(db *sql.DB) []Comment {
	comments := make([]Comment, 0)
	rows, err := db.Query("SELECT * FROM Comments WHERE (Approved = 0)")
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

func ApprovedComments(db *sql.DB, url string, email string) []Comment {
	rows, err := db.Query("SELECT * FROM Comments WHERE (Approved = 1 OR Email = ?) AND Url = ?", email, url)
	if err != nil {
		logErr(err, "Could not load comments:")
		return nil
	}
	defer rows.Close()
	comments, err := scanComments(rows)

	err = rows.Err()
	if err != nil {
		logErr(err, "Could not read comments")
	}
	return comments
}

func UnapprovedCommentsCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT * FROM Comments WHERE Approved<>1").Scan(&count)
	return count, err
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

func AllCommentsPaginated(db *sql.DB, page int) ([]Comment, int) {
	rows, err := db.Query("SELECT * FROM COMMENTS ORDER BY Created DESC LIMIT 10 OFFSET ?", page*10)
	if err != nil {
		logErr(err, "Could not load comments:")
		return nil, 0
	}
	defer rows.Close()
	comments, err := scanComments(rows)

	err = rows.Err()
	if err != nil {
		logErr(err, "Could not read comments")
	}
	countRows, err := db.Query("SELECT CEIL(COUNT(*)/10) FROM comments")
	defer countRows.Close()
	if err != nil {
		logErr(err, "Could not find comment pages count:")
		return comments, 0
	}
	pages := 1
	if countRows.Next() {
		countRows.Scan(&pages)
	}
	return comments, pages
}

func scanComment(row *sql.Row) (*Comment, error) {
	comment := Comment{}
	err := row.Scan(
		&comment.Id,
		&comment.Created,
		&comment.Email,
		&comment.Name,
		&comment.Body,
		&comment.Url,
		&comment.ClientIp,
		&comment.Approved)
	return &comment, err
}

func scanComments(rows *sql.Rows) ([]Comment, error) {
	comments := make([]Comment, 0)
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
			return comments, err
		} else {
			comments = append(comments, comment)
		}
	}

	return comments, nil
}
