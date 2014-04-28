package models

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/coopernurse/gorp"
	"time"
)

// User is a backend user/admin.
type User struct {
	Id       int64  `form:"-"`
	Created  int64  `form:"-"`
	Email    string `binding:"required" form:"email"`
	Password string `binding:"required" form:"password"`
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
func UserCount(dbmap *gorp.DbMap) int {
	var ids []string
	dbmap.Select(&ids, "select id from users")
	return len(ids)
}
