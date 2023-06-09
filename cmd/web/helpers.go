package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getEnv(name string, errorLog *log.Logger) string {
	varEnv := os.Getenv(name)
	if varEnv == "" {
		ErrDuplicateEmail := fmt.Errorf("empty environment variable %s", name)
		errorLog.Fatal(ErrDuplicateEmail)
	}
	return varEnv
}

func openDB(errorLog *log.Logger) (*gorm.DB, error) {
	DB_NAME := getEnv("DB_NAME", errorLog)
	DB_USER := getEnv("DB_USER", errorLog)
	DB_PASSWORD := getEnv("DB_PASSWORD", errorLog)

	dsn := fmt.Sprintf(
		"host=localhost user=%s password=%s dbname=%s port=5432 sslmode=disable",
		DB_USER,
		DB_PASSWORD,
		DB_NAME,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		errorLog.Panic("failed to connect database")
	}

	return db, err
}
