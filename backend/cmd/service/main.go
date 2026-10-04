package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MorningBlossom/nexus/backend/internal/api"
	"github.com/MorningBlossom/nexus/backend/internal/config"
	"github.com/MorningBlossom/nexus/backend/internal/github"
)

func main() {
	_ = config.Load()

	client, err := github.NewClient()
	if err != nil {
		log.Fatalf("GitHub client initialization failed: %v", err)
	}

	repository := github.NewRepository(
		client.API,
		//client.HTTPClient,
	)

	handler := api.NewHandler(repository)

	mux := http.NewServeMux()

	mux.HandleFunc("/getPackages", handler.GetPackages)
	mux.HandleFunc("/getInstallationRepositories", handler.GetInstallationRepositories)

	server := &http.Server{
		Addr:              ":3000",
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	log.Println("review API listening on :8080")

	err = server.ListenAndServe()
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

	shutdownContext, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		log.Printf(
			"graceful shutdown failed: %v",
			err,
		)

		if err := server.Close(); err != nil {
			log.Printf(
				"server close failed: %v",
				err,
			)
		}
	}

	log.Println("Nexus API stopped")

}
