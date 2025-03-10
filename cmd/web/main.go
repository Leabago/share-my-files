package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"share-my-files/pkg/models/operation"
	"time"

	"github.com/go-redis/redis"
)

type application struct {
	logger *AppLogger

	redisClient *redis.Client

	files interface {
		Insert()
		Get()
		Latest()
	}

	templateCache map[string]*template.Template

	maxFileSize int
}

type AppLogger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
}

func main() {
	fmt.Println("start share-my-files! ddns")

	// create logger
	logFormat := log.Ldate | log.Ltime | log.Lshortfile
	infoLog := log.New(os.Stdout, "INFO\t", logFormat)
	debugLog := log.New(os.Stdout, "DEBUG\t", logFormat)
	errorLog := log.New(os.Stderr, "ERROR\t", logFormat)

	logger := &AppLogger{
		infoLog:  infoLog,
		debugLog: debugLog,
		errorLog: errorLog,
	}

	// connect to redis
	redisAddr := getEnv("REDIS_ADDR", logger)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	pong, err := redisClient.Ping().Result()
	fmt.Println("ping redis", pong, err)

	// create folders
	createFolderForFiles(folderPath, logger)
	// createFolderForFiles(configFolderPath, logger)

	// create file with ddns address for javascript
	writeDdnsAddress(getEnv("DDNS_ADDRESS", logger), logger)

	app := &application{
		logger:      logger,
		files:       &operation.FileModel{},
		redisClient: redisClient,
		maxFileSize: writeFileSize(logger),
	}

	// delete files every 10 seconds
	go app.deleteFileEveryNsec(10)

	templateCache, err := app.newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}
	app.templateCache = templateCache

	srv := &http.Server{
		Handler: app.routes(),
		Addr:    APP_PORT,
		// Good practice: enforce timeouts for servers you create!
		IdleTimeout:  time.Minute,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	infoLog.Printf("Starting server on %s", APP_PORT)

	err = srv.ListenAndServe()
	// err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}
