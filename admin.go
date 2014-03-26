package main

import (
	"github.com/codegangsta/martini-contrib/render"
	"github.com/coopernurse/gorp"
)

func AdminIndex(ren render.Render, dbmap *gorp.DbMap) {
	var comments []Comment
	dbmap.Select(&comments, "select * from comments order by created desc limit 10")
	ren.HTML(200, "admin/index", comments, render.HTMLOptions{
		Layout: "admin/layout",
	})
}
