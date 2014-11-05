package handler

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"os"
)

func Setup(cfgPath string) {
	router := mux.NewRouter()
	router.HandleFunc("/setup/database", TestDatabase).Methods("GET")
	// router.Handle("/setup/database", CreateComment).Methods("POST")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("public/setup/")))
	http.Handle("/", router)
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	http.ListenAndServe(host+":"+port, nil)
}

// CheckDatabase tries to connect to the database
func TestDatabase(w http.ResponseWriter, req *http.Request) {
	qry := req.URL.Query()
	driver, host, port := qry.Get("driver"), qry.Get("host"), qry.Get("port")
	user, password, database := qry.Get("username"), qry.Get("password"), qry.Get("database")
	path := qry.Get("path")

	var valid bool
	switch driver {
	case "sqlite":
		valid = validateSqlite(path)
	case "postgresql":
		valid = validatePostgresql(host, port, user, password, database)
	case "mysql":
		valid = validateMysql(host, port, user, password, database)
	}

	w.Header().Set("Content-Type", "application/json")
	out, _ := json.Marshal(valid)
	w.Write(out)
}

func validateSqlite(path string) bool {
	return true
}

func validatePostgresql(host, port, user, password, database string) bool {
	db, err := sqlx.Connect("postgres",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, database, password))
	if err != nil {
		return false
	}
	if err = db.Ping(); err != nil {
		return false
	}
	return true
}

func validateMysql(host, port, user, password, database string) bool {
	db, err := sqlx.Connect("mysql",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, database, password))
	if err != nil {
		return false
	}
	if err = db.Ping(); err != nil {
		return false
	}
	return true
}
