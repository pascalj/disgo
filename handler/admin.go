package handler

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/pascalj/disgo/models"
	"net/http"
	"strconv"
	"strings"
)

// AdminIndex shows the overview of the admin interface with latest comments.
func AdminIndex(w http.ResponseWriter, req *http.Request, app *App) {
	qry := req.URL.Query()
	page, err := strconv.Atoi(qry.Get("page"))

	if err != nil {
		page = 0
	}
	comments := paginatedComments(app.Db, page)
	render(w, "admin/index", map[string]interface{}{"comments": comments}, app)
}

// GetRegister shows the register form.
func GetRegister(w http.ResponseWriter, req *http.Request, app *App) {
	render(w, "admin/register", nil, app)
}

// PostUser will create a new user when no other users are in the database.
// If users are present, it will redirect to the login.
func PostUser(w http.ResponseWriter, req *http.Request, app *App) {
	if models.UserCount(app.Db) == 0 {
		email, password := req.FormValue("email"), req.FormValue("password")
		user := models.NewUser(email, password)
		err := user.Save(app.Db)
		if err != nil {
			http.Redirect(w, req, app.Config.General.Prefix+"/register", http.StatusFound)
		} else {
			http.Redirect(w, req, app.Config.General.Prefix+"/login", http.StatusFound)
		}
	}
}

// GetLogin shows the login form for the backend.
func GetLogin(w http.ResponseWriter, req *http.Request, app *App) {
	if models.UserCount(app.Db) > 0 {
		render(w, "admin/login", map[string]interface{}{"hideNav": true}, app)
	} else {
		http.Redirect(w, req, app.Config.General.Prefix+"/register", http.StatusSeeOther)
	}

}

// PostLogin takes the email and password parameter and logs the user in if they are correct.
func PostSession(w http.ResponseWriter, req *http.Request, app *App) {
	var user models.User

	email, password := req.FormValue("email"), req.FormValue("password")
	user, err := models.UserByEmail(app.Db, email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		http.Redirect(w, req, app.Config.General.Prefix+"/login", http.StatusSeeOther)
		return
	}

	session, _ := app.SessionStore.Get(req, SessionName)
	session.Values["userId"] = user.Id
	session.Save(req, w)

	http.Redirect(w, req, app.Config.General.Prefix+"/admin/", http.StatusSeeOther)
}

// GetIndex shows a simple introduction.
func GetIndex(w http.ResponseWriter, req *http.Request, app *App) {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	base := []string{scheme, "://", req.Host, app.Config.General.Prefix}
	render(w, "index", map[string]interface{}{"base": strings.Join(base, ""), "hideNav": true}, app)
}

func GetSettings(w http.ResponseWriter, req *http.Request, app *App) {
	render(w, "admin/settings", map[string]interface{}{"config": app.Config}, app)
}

// PostLogout logs the user out and redirects to the login page.
func PostLogout(w http.ResponseWriter, req *http.Request, app *App) {
	session, _ := app.SessionStore.Get(req, SessionName)
	session.Values["userId"] = nil
	session.Save(req, w)
	http.Redirect(w, req, app.Config.General.Prefix+"/login", http.StatusFound)
}
