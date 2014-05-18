package handler

import (
	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/pascalj/disgo/models"
	"net/http"
	"strings"
	"time"
)

// GetComments will display all comments for a given URL-parameter. If configured, it only
// displays approved comments.
func GetComments(w http.ResponseWriter, req *http.Request, app *App) {
	qry := req.URL.Query()
	ses, _ := app.SessionStore.Get(req, "disgo")
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
	_ = ctx

	if len(comments) > 0 {
		renderComments(w, "comments", ctx, app)
	}
}

// ApproveComment allows admins to approve a comment by id.
func ApproveComment(ren render.Render, params martini.Params, dbmap *gorp.DbMap) {
	obj, err := dbmap.Get(models.Comment{}, params["id"])
	if err != nil || obj == nil {
		ren.Error(404)
	} else {
		comment := obj.(*models.Comment)
		comment.Approved = true
		dbmap.Update(comment)
		ren.Redirect("/admin")
	}
}

// CreateComment validates and creates a new comment. It also saves the client's IP-adress
// to reduce spam.
func CreateComment(w http.ResponseWriter, req *http.Request, app *App) {
	comment := models.NewComment("email", "name", "title", "body", "url", "ip", "id")
	comment.Created = time.Now().Unix()
	comment.ClientIp = strings.Split(req.RemoteAddr, ":")[0]
	// err := dbmap.Insert(&comment)
	// if err != nil {
	// 	ren.JSON(400, err.Error())
	// } else {
	// 	session.Set("email", comment.Email)
	// 	session.Set("name", comment.Name)
	// 	go notifier.CommentCreated(&comment)
	// 	view.RenderComment(comment, nil, ren)
	// }
}

// DestroyComment deletes a comment from the database by id.
func DestroyComment(ren render.Render, params martini.Params, dbmap *gorp.DbMap) {
	obj, err := dbmap.Get(models.Comment{}, params["id"])
	if err != nil || obj == nil {
		ren.JSON(404, nil)
	} else {
		comment := obj.(*models.Comment)
		count, err := dbmap.Delete(comment)
		if count != 1 || err != nil {
			ren.JSON(500, err.Error())
		} else {
			ren.Redirect("/admin")
		}
	}
}

func renderComment(w http.ResponseWriter, tmpl string, comment models.Comment, app *App) {
	err := app.Templates.ExecuteTemplate(w, tmpl+".tmpl", comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderComments(w http.ResponseWriter, tmpl string, ctx map[string]interface{}, app *App) {
	err := app.Templates.ExecuteTemplate(w, tmpl+".tmpl", ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
