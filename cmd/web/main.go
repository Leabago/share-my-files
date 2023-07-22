package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"share-my-file/pkg/models/operation"
	"time"
)

func home(w http.ResponseWriter, r *http.Request) {

	asd := []byte("asd")
	w.Write(asd)
}

type application struct {
	// infoLog  *log.Logger
	// errorLog *log.Logger
	// debugLog *log.Logger

	logger AppLogger

	files interface {
		Insert()
		Get()
		Latest()
	}

	templateCache map[string]*template.Template
}

type AppLogger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
}

func main() {
	fmt.Println("start app")

	// create folder

	logFormat := log.Ldate | log.Ltime | log.Lshortfile
	infoLog := log.New(os.Stdout, "INFO\t", logFormat)
	debugLog := log.New(os.Stdout, "DEBUG\t", logFormat)
	errorLog := log.New(os.Stderr, "ERROR\t", logFormat)

	logger := AppLogger{
		infoLog:  infoLog,
		debugLog: debugLog,
		errorLog: errorLog,
	}

	app := &application{
		logger: logger,
		files:  &operation.FileModel{},
	}

	templateCache, err := app.newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}
	app.templateCache = templateCache
	// APP_PORT := getEnv("APP_PORT", logger)
	APP_PORT := ":8080"
	createFolderForFiles(logger)

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
