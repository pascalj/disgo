package models

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/jmoiron/sqlx"
	"time"
)

// User is a backend user/admin.
type User struct {
	Id       int64  `form:"-"`
	Created  int64  `form:"-"`
	Email    string `binding:"required" form:"email"`
	Password string `binding:"required" form:"password"`
}

func (u *User) Save(db *sqlx.DB) error {
	stmt, err := db.Prepare(`
		INSERT INTO
		users(Created, Email, Password)
		VALUES(?, ?, ?)`)
	if err != nil {
		return err
	}

	res, err := stmt.Exec(u.Created, u.Email, u.Password)
	if err != nil {
		return err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return err
	}
	u.Id = lastId
	return nil
}

func UserByEmail(db *sqlx.DB, email string) (User, error) {
	row := db.QueryRow("SELECT Id, Email, Password FROM users WHERE Email = ?", email)
	user := User{}
	err := row.Scan(&user.Id, &user.Email, &user.Password)

	if err != nil {
		return User{}, err
	}
	return user, nil
}

func UserById(db *sqlx.DB, id int64) (User, error) {
	row := db.QueryRow("SELECT Id, Email, Password FROM users WHERE Id = ?", id)
	user := User{}
	err := row.Scan(&user.Id, &user.Email, &user.Password)

	if err != nil {
		return User{}, err
	}
	return user, nil
}

// NewUser creates a new user while automatically hashing the password.
func NewUser(email, password string) User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return User{
		Created:  time.Now().UnixNano(),
		Email:    email,
		Password: string(hashedPassword),
	}
}

// UserCount gets the number of users already in the database.
func UserCount(db *sqlx.DB) int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}
