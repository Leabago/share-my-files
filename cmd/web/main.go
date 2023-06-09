package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

func home(w http.ResponseWriter, r *http.Request) {

	asd := []byte("asd")
	w.Write(asd)
}

type application struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger

	files interface {
		Insert()
		Get()
		Latest()
	}

	templateCache map[string]*template.Template
}

func main() {
	fmt.Println("start app")

	logFormat := log.Ldate | log.Ltime | log.Lshortfile
	infoLog := log.New(os.Stdout, "INFO\t", logFormat)
	debugLog := log.New(os.Stdout, "DEBUG\t", logFormat)
	errorLog := log.New(os.Stderr, "ERROR\t", logFormat)

	app := &application{
		infoLog:  infoLog,
		debugLog: debugLog,
		errorLog: errorLog,

		// files: &operation.FileModel{DB: db},
	}
	templateCache, err := app.newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}
	app.templateCache = templateCache

	APP_PORT := getEnv("APP_PORT", errorLog)

	srv := &http.Server{
		Handler: app.routes(),
		Addr:    APP_PORT,
		// Good practice: enforce timeouts for servers you create!
		IdleTimeout:  time.Minute,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	infoLog.Printf("Starting server on %s", APP_PORT)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}
