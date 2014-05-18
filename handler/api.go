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
		"email": email,
		"name":  name,
	}
	_ = ctx

	if len(comments) > 0 {
		w.Write([]byte("success"))
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

// MapView maps the View type for martini depending on the accept header. It is used
// to generate the appropriate output for html and json.
func MapView(ctx martini.Context, res http.ResponseWriter, req *http.Request) {
	accept := req.Header["Accept"]
	if accept[0] != "" {
		accept = strings.Split(accept[0], ",")
	}
	switch accept[0] {
	case "text/html":
		ctx.MapTo(models.HtmlView{}, (*models.View)(nil))
		res.Header().Set("Content-Type", "text/html")
	default:
		ctx.MapTo(models.JsonView{}, (*models.View)(nil))
		res.Header().Set("Content-Type", "application/json")
	}
}
