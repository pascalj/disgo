package main

import (
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/coopernurse/gorp"
	"net/http"
)

func GetComments(ren render.Render, params martini.Params, dbmap *gorp.DbMap) {
	var comments []Comment
	dbmap.Select(&comments, "select * from comments order by created asc")
	ren.JSON(200, comments)
}

func GetComment(ren render.Render, params martini.Params, dbmap *gorp.DbMap) {
	obj, err := dbmap.Get(Comment{}, params["id"])
	if err != nil || obj == nil {
		ren.JSON(404, nil)
	} else {
		comment := obj.(*Comment)
		ren.JSON(200, comment)
	}
}

func UpdateComment(ren render.Render, params martini.Params, req *http.Request, dbmap *gorp.DbMap) {
	obj, err := dbmap.Get(Comment{}, params["id"])
	if err != nil || obj == nil {
		ren.JSON(404, nil)
	} else {
		comment := obj.(*Comment)
		comment.Email = req.FormValue("email")
		comment.Title = req.FormValue("title")
		comment.Body = req.FormValue("body")
		comment.Url = req.FormValue("url")
		ren.JSON(200, comment)
	}
}

func CreateComment(ren render.Render, req *http.Request, dbmap *gorp.DbMap) {
	comment := NewComment(req.FormValue("email"), req.FormValue("title"), req.FormValue("body"), req.FormValue("url"), "0")
	err := dbmap.Insert(&comment)
	if err != nil {
		ren.JSON(400, err.Error())
	} else {
		ren.JSON(200, &comment)
	}
}

func DestroyComment(ren render.Render, params martini.Params, dbmap *gorp.DbMap) {
	obj, err := dbmap.Get(Comment{}, params["id"])
	if err != nil || obj == nil {
		ren.JSON(404, nil)
	} else {
		comment := obj.(*Comment)
		count, err := dbmap.Delete(comment)
		if count != 1 || err != nil {
			ren.JSON(500, err.Error())
		}
	}
}
