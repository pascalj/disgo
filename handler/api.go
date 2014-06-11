package handler

import (
	"github.com/gorilla/mux"
	"github.com/pascalj/disgo/models"
	"net/http"
	"strconv"
	"time"
)

// GetComments will display all comments for a given URL-parameter. If configured, it only
// displays approved comments.
func GetComments(w http.ResponseWriter, req *http.Request, app *App) {
	qry := req.URL.Query()
	ses, _ := app.SessionStore.Get(req, SessionName)
	email := ""
	name := ""
	if val := ses.Values["email"]; val != nil {
		email = val.(string)
	}
	if val := ses.Values["name"]; val != nil {
		name = val.(string)
	}
	comments := make([]models.Comment, 0)
	if qry["url"] == nil {
		return
	}

	if app.Config.General.Approval {
		comments = models.ApprovedComments(app.Db, qry["url"][0], email)
	} else {
		comments = models.AllComments(app.Db, qry["url"][0])
	}

	ctx := map[string]interface{}{
		"email":    email,
		"name":     name,
		"comments": comments,
	}

	render(w, "partial/comments", ctx, app)
}

// ApproveComment allows admins to approve a comment by id.
func ApproveComment(w http.ResponseWriter, req *http.Request, app *App) {
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(422)
		return
	}
	comment := models.GetComment(app.Db, id)
	if comment == nil {
		w.WriteHeader(404)
	} else {
		comment.Approved = true
		comment.Save(app.Db)
		http.Redirect(w, req, "/admin", 303)
	}
}

// CreateComment validates and creates a new comment. It also saves the client's IP-adress
// to reduce spam.
func CreateComment(w http.ResponseWriter, req *http.Request, app *App) {
	comment := models.NewComment(
		req.FormValue("email"),
		req.FormValue("name"),
		req.FormValue("title"),
		req.FormValue("body"),
		req.FormValue("url"),
		req.RemoteAddr)
	comment.Created = time.Now().Unix()

	ip, err := relevantIpBytes(req.RemoteAddr)
	if err != nil {
		ip = req.RemoteAddr
	}
	comment.ClientIp = ip

	valid, valErrors := comment.Validate()
	if valid {
		err := comment.Save(app.Db)
		if err != nil {
			w.WriteHeader(500)
		} else {
			session, _ := app.SessionStore.Get(req, SessionName)
			session.Values["email"] = comment.Email
			session.Values["name"] = comment.Name
			session.Save(req, w)
			renderComment(w, "partial/comment", comment, app)
		}
	} else {
		renderErrors(w, valErrors, 422)
	}

	go app.Notifier.CommentCreated(&comment)
}

// DestroyComment deletes a comment from the database by id.
func DestroyComment(w http.ResponseWriter, req *http.Request, app *App) {
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(422)
		return
	}
	comment := models.GetComment(app.Db, id)
	if comment == nil {
		w.WriteHeader(404)
	} else {
		err := comment.Delete(app.Db)
		if err != nil {
			w.WriteHeader(500)
		} else {
			http.Redirect(w, req, "/admin", 303)
		}
	}
}
