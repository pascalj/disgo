package handler

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"net/http"
	"os"
)

var listener net.Listener

type dbConfig struct {
	Driver   string
	Host     string
	Port     string
	Username string
	Password string
	Database string
	Path     string
}

func Setup(cfgPath string) {
	router := mux.NewRouter()
	router.HandleFunc("/setup/database", testDatabase).Methods("GET")
	router.HandleFunc("/setup/database", writeConfigHandler(cfgPath)).Methods("POST")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("public/setup/")))
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	var err error
	listener, err = net.Listen("tcp", host+":"+port)
	if err != nil {
		return
	}
	http.Serve(listener, router)
}

// testDatabase tries to connect to the database
func testDatabase(w http.ResponseWriter, req *http.Request) {
	qry := req.URL.Query()
	cfg := dbConfig{
		qry.Get("driver"),
		qry.Get("host"),
		qry.Get("port"),
		qry.Get("username"),
		qry.Get("password"),
		qry.Get("database"),
		qry.Get("path"),
	}
	w.Header().Set("Content-Type", "application/json")
	out, _ := json.Marshal(validateDatabase(cfg))
	w.Write(out)
}

// setDatabase writes the minimal config file to disk
func writeConfigHandler(cfgPath string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cfg := dbConfig{
			req.FormValue("driver"),
			req.FormValue("host"),
			req.FormValue("port"),
			req.FormValue("username"),
			req.FormValue("password"),
			req.FormValue("database"),
			req.FormValue("path"),
		}
		if validateDatabase(cfg) {
			if err := writeConfig(cfgPath, cfg); err == nil {
				// success writing the config
				out, _ := json.Marshal(true)
				w.Write(out)
				listener.Close()
			} else {
				out, _ := json.Marshal(err.Error())
				w.Write(out)
			}
		} else {
			out, _ := json.Marshal("Invalid credentials.")
			w.Write(out)
		}
	}
}

func writeConfig(cfgPath string, cfg dbConfig) error {
	cfgYaml, err := yaml.Marshal(map[string]dbConfig{"database": cfg})
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cfgPath, cfgYaml, 0644)
}

func validateDatabase(cfg dbConfig) bool {
	switch cfg.Driver {
	case "sqlite3":
		return validateSqlite3(cfg)
	case "postgresql":
		return validatePostgresql(cfg)
	case "mysql":
		return validateMysql(cfg)
	default:
		return false
	}
}

func validateSqlite3(cfg dbConfig) bool {
	return true
}

func validatePostgresql(cfg dbConfig) bool {
	db, err := sqlx.Connect("postgres",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Password))
	if err != nil {
		return false
	}
	if err = db.Ping(); err != nil {
		return false
	}
	return true
}

func validateMysql(cfg dbConfig) bool {
	db, err := sqlx.Connect("mysql",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Password))
	if err != nil {
		return false
	}
	if err = db.Ping(); err != nil {
		return false
	}
	return true
}
