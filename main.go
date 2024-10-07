package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	http.Handle("/create-room", corsMiddleware(http.HandlerFunc(createRoom)))
	http.Handle("/encrypt", corsMiddleware(http.HandlerFunc(handleEncrypt)))
	http.Handle("/decrypt", corsMiddleware(http.HandlerFunc(handleDecrypt)))
	http.HandleFunc("/ws", handleConnections)

	go manageRooms(ctx)

	srv := &http.Server{
		Addr: ":8080",
	}

	go func() {
		logger.Println("Server started on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("ListenAndServe: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("Shutting down server...")

	cancel()
	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Fatalf("Server shutdown: %v", err)
	}

	logger.Println("Server exited")
}
