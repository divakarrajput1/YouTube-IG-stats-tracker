package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"video-stats-tracker/internal/endpoint"
	"video-stats-tracker/internal/repository"
	"video-stats-tracker/internal/service"
	httpTransport "video-stats-tracker/internal/transport/http"
	"video-stats-tracker/internal/worker"
)

func main() {
    // Configuration
    youtubeAPIKey := getEnv("YOUTUBE_API_KEY", "AIzaSyDzce9GZEMoL15w5Xv7j_x3E6gUbapWHyA")
    instagramToken := getEnv("INSTAGRAM_TOKEN", "EAALI2aRc08wBPnURGNzA5dS0pJkvkWZBKDLLN4W3aJGZCuua5d9fOJ3Uw5RFhuOij1yRKwNn3BL9RlL0jV5ERZA3ebEbzQLKMEUZAJKGOFXQcD754USCDYFKpLZB4AOZATkPJJAntGUzrhQa48FqOZAc9sraH3eUCrEvlqyntNew05CGlNfpIiRn7MPwl6VSB7VBmSfXxvUQrzEjLiSmg8nsYtKZBSnzMJG2uaggJEBkH9pzQwZDZD")
    instagramID := getEnv("INSTAGRAM_ID", "17841477784603001")
    httpPort := getEnv("HTTP_PORT", "8080")

    // Initialize repository - Using SQLite
    repo, err := repository.NewSQLiteRepository()
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    // Initialize service
    svc := service.NewService(repo, youtubeAPIKey, instagramToken, instagramID)

    // Initialize endpoints
    endpoints := endpoint.MakeEndpoints(svc)

    // Initialize HTTP handler
    handler := httpTransport.NewHTTPHandler(endpoints)

    // Initialize polling worker
    poller := worker.NewPoller(svc)
    poller.Start()
    defer poller.Stop()

    // Start HTTP server
    server := &http.Server{
        Addr:         ":" + httpPort,
        Handler:      handler,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    // Start server in goroutine
    go func() {
        log.Printf("Server starting on port %s", httpPort)
        //log.Printf("Tracking Instagram account: %s", instagramUsername)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")

    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server exited")
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}