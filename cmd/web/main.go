package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"share-my-files/pkg/models"
	"share-my-files/pkg/models/operation"
	"strconv"
	"syscall"
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

	maxFileSize int64

	healthCheck *models.HealthCheck
}

type AppLogger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
}

func main() {
	fmt.Println("start share-my-files! helm 1")

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

	// get enviroment variables
	// get redist address
	redisAddr := getEnv("REDIS_ADDR", logger)
	// get variable defining maximum file size
	maxFileSize := getEnv("MAX_FILE_SIZE", logger)
	// port for aplication
	appPort := getEnv("APP_PORT", logger)

	// connect to redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	pong, err := redisClient.Ping().Result()
	fmt.Println("ping redis", pong, err)

	// create folders
	createFolderForFiles(folderPath, logger)

	// create file with maxFileSize for javascript
	writeVariable("var MAX_FILE_SIZE = "+maxFileSize+";", maxFileSizeFileName, logger)
	maxFileSizeInt64, err := strconv.ParseInt(maxFileSize, 10, 64)
	if err != nil {
		logger.errorLog.Fatal(err)
	}

	// health check for readiness probe
	healthCheck := models.NewHealthCheck()

	// Add Redis checker
	healthCheck.AddChecker("redis", &models.RedisChecker{RedisClient: redisClient})

	app := &application{
		logger:      logger,
		files:       &operation.FileModel{},
		redisClient: redisClient,
		maxFileSize: maxFileSizeInt64,
		healthCheck: healthCheck,
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
		Addr:    appPort,
		// Good practice: enforce timeouts for servers you create!
		IdleTimeout:  time.Minute,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	infoLog.Printf("Starting server on %s", appPort)

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServeTLS("./tls/certificate.crt", "./tls/private.key"); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

}
