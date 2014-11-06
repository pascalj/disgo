package main

import (
	"bufio"
	"flag"
	"github.com/pascalj/disgo/handler"
	"github.com/pascalj/disgo/models"
	"github.com/pascalj/disgo/service"
	"log"
	"net/http"
	"os"
)

var (
	cfg        models.Config
	cfgPath    string
	importPath string
	app        *handler.App
)

func init() {
	flag.StringVar(&cfgPath, "config", "disgo.yml", "path to the config file")
	flag.StringVar(&importPath, "import", "", "Disqus XML file to import")
	flag.Parse()
}

func main() {
	if _, err := os.Stat(cfgPath); err != nil {
		handler.Setup(cfgPath)
	}
	app, err := handler.NewApp(cfgPath)
	checkErr(err, "Unable to start Disgo:")
	if importPath != "" {
		file, err := os.Open(importPath)
		checkErr(err, "Could not open Disqus XML file:")
		reader := bufio.NewReader(file)
		service.Import(app.Db, reader)
		return
	}
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	err = http.ListenAndServe(host+":"+port, app.Router)
	if err != nil {
		log.Fatal("Unable to start Disgo:", err)
	}
}

func checkErr(err error, description string) {
	if err != nil {
		log.Fatalln(description, err)
	}
}
