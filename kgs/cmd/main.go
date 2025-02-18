package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/bikraj2/url_shortener/kgs/internal/controller"
	httphandler "github.com/bikraj2/url_shortener/kgs/internal/handler/http"

	"github.com/bikraj2/url_shortener/kgs/internal/repository"
	"github.com/redis/go-redis/v9"
)

func main() {

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	repo := repository.New(client)

	fmt.Println("Repository initialized successfully!", repo)

	ctrl := controller.New(repo)

	fmt.Println("controller initialized successfully!", repo)

	h := httphandler.New(ctrl)
	// Set up HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.Handle) // Assuming you have a `ShortenURL` method

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
