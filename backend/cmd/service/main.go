package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MorningBlossom/nexus/backend/internal/api"
	"github.com/MorningBlossom/nexus/backend/internal/config"
	"github.com/MorningBlossom/nexus/backend/internal/db"
	"github.com/MorningBlossom/nexus/backend/internal/github"
	"github.com/MorningBlossom/nexus/backend/internal/models"
	"github.com/MorningBlossom/nexus/backend/internal/repository"
)

func seedDatabase(ctx context.Context, repo *repository.MongoRepository) {
	log.Println("Seeding database...")

	apps := []models.App{
		{AppID: "auth-service", Name: "Auth Service", Mark: "A", Description: "Authentication and Authorization", Repo: "org/auth-service"},
		{AppID: "ai-code-review", Name: "AI Code Review", Mark: "R", Description: "AI powered code analysis", Repo: "org/ai-code-review"},
	}

	for _, app := range apps {
		if err := repo.SaveApp(ctx, &app); err != nil {
			log.Printf("Error seeding app %s: %v", app.AppID, err)
		}
	}

	services := []models.Service{
		{AppID: "auth-service", ServiceID: "auth-service/service", Runtime: "Go", Port: 8080, Capacity: "4 pods running", DeployedAt: time.Now()},
		{AppID: "auth-service", ServiceID: "auth-service", Runtime: "Go", Port: 8081, Capacity: "2 pods running", DeployedAt: time.Now()},
		{AppID: "ai-code-review", ServiceID: "ai-code-review", Runtime: "Go", Port: 5000, Capacity: "1 pod running", DeployedAt: time.Now()},
	}

	for _, svc := range services {
		if err := repo.SaveService(ctx, &svc); err != nil {
			log.Printf("Error seeding service %s: %v", svc.ServiceID, err)
		}
	}
	log.Println("Database seeding completed.")
}

func main() {
	cfg := config.Load()

	if cfg == nil {
		log.Fatal("Failed to load configuration")
	}

	// MongoDB Connection
	mongoClient, err := db.Connection(cfg.DBurl)
	if err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}
	mongoDB := mongoClient.Database("nexus")
	mongoRepo := repository.NewMongoRepository(mongoDB)

	// Seed data
	seedDatabase(context.Background(), mongoRepo)

	// GitHub Client
	client, err := github.NewClient()
	if err != nil {
		log.Fatalf("GitHub client initialization failed: %v", err)
	}

	repository := github.NewRepository(client.Client)

	handler := api.NewHandler(repository, mongoRepo)

	mux := http.NewServeMux()

	mux.HandleFunc("/apps", handler.GetApps)
	mux.HandleFunc("/services", handler.GetServicesByApp)
	mux.HandleFunc("/getPackages", handler.GetPackagesByName)
	mux.HandleFunc("/getPackagesByVersion", handler.GetAllPackagesByRepo)
	mux.HandleFunc("/getInstallationRepositories", handler.GetInstallationRepositories)

	// Serve static files from the frontend directory
	fs := http.FileServer(http.Dir(cfg.FrontendStaticPath))
	mux.Handle("/", fs)

	server := &http.Server{
		Addr:              ":3000",
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	log.Println("Nexus API and Frontend listening on :3000")

	err = server.ListenAndServe()
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
