package main

import (
  "github.com/codegangsta/martini-contrib/render"
)

type View interface {
  RenderComment(Comment, render.Render)
  RenderComments([]Comment, render.Render)
}

type HtmlView struct {}
type JsonView struct {}

func (view HtmlView) RenderComment(comment Comment, ren render.Render) {
  ren.HTML(200, "comment", &comment)
}

func (view JsonView) RenderComment(comment Comment, ren render.Render) {
  ren.JSON(200, &comment)
}

func (view HtmlView) RenderComments(comments []Comment, ren render.Render) {
  ren.HTML(200, "comments", comments)
}

func (view JsonView) RenderComments(comments []Comment, ren render.Render) {
  ren.JSON(200, comments)
}
