package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/phuangpheth/rolePermission/database"
	mdw "github.com/phuangpheth/rolePermission/middleware"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
		os.Exit(1)
	}
}

func main() {
	dbDriver := getEnv("DB_DRIVER", "postgres")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5455")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	timeZone := getEnv("TZ", "Asia/Vientiane")

	dbConn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=%s", dbHost, dbPort, dbUser, dbPass, dbName, timeZone)
	db, err := database.Open(dbDriver, dbConn)
	failOnError(err, "failed to connect to database")
	defer func() {
		if err := db.Close(); err != nil {
			failOnError(err, "failed to close database")
		}
	}()

	err = db.Ping(context.Background())
	failOnError(err, "failed to ping database")

	e := echo.New()
	e.Use(middleware.Logger())

	e.POST("/v1/enrollments", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	}, mdw.RoleMiddleware(mdw.Config{
		DB: db,
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/greeting"
		},
	}))

	e.GET("/v1/enrollments", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	}, mdw.RoleMiddleware(mdw.Config{
		DB: db,
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/greeting"
		},
	}))

	go func() {
		if err := e.Start(fmt.Sprintf(":%s", getEnv("PORT", "8080"))); err != nil {
			e.Logger.Fatal("Shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutdown in progress...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal("Failed to shutdown the server", err)
	}

}
