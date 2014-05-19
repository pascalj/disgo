package main

import (
	"flag"
	"github.com/go-martini/martini"
	"github.com/pascalj/disgo/handler"
	"github.com/pascalj/disgo/models"
	"log"
	"net/http"
	"os"
)

var (
	m          *martini.ClassicMartini
	cfg        models.Config
	cfgPath    string
	importPath string
	help       bool
	app        *handler.App
)

func init() {
	flag.StringVar(&cfgPath, "config", "disgo.gcfg", "path to the config file")
	flag.StringVar(&importPath, "import", "", "Disqus XML file to import")
	flag.Parse()
}

func main() {
	// if importPath != "" {
	// 	file, err := os.Open(importPath)
	// 	checkErr(err, "Could not open Disqus XML file:")
	// 	reader := bufio.NewReader(file)
	// 	return
	// }
	app, err := handler.NewApp(cfgPath)
	checkErr(err, "Unable to start Disgo:")
	http.Handle("/", app.Router)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}

func checkErr(err error, description string) {
	if err != nil {
		log.Fatalln(description, err)
	}
}
