package api

import (
	"encoding/json"
	"net/http"

	"github.com/MorningBlossom/nexus/backend/internal/models"

	"github.com/MorningBlossom/nexus/backend/internal/github"
	"github.com/MorningBlossom/nexus/backend/internal/repository"
)

type Handler struct {
	githubRepo *github.Repository
	mongoRepo  *repository.MongoRepository
}

func NewHandler(ghRepo *github.Repository, mongoRepo *repository.MongoRepository) *Handler {
	return &Handler{
		githubRepo: ghRepo,
		mongoRepo:  mongoRepo,
	}
}

func (h *Handler) GetApps(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	apps, err := h.mongoRepo.ListApps(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if apps == nil {
		apps = []*models.App{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apps)
}

func (h *Handler) GetInstallationRepositories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.githubRepo.GetInstallationRepositories(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetAllPackagesByRepo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	owner := "MorningBlossom"
	repoName := "auth-service/service"

	services, err := h.githubRepo.GetPackagesByPackageName(ctx, owner, repoName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(
		services,
	)
}

func (h *Handler) GetPackagesByName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	owner := "MorningBlossom"
	repoName := "auth-service"

	packages, err := h.githubRepo.GetPackagesByName(ctx, owner, repoName)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(
	    packages,
	)
}

func (h *Handler) GetPackageByVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	owner := "MorningBlossom"
	packageName := "ai-code-review"

	packages, err := h.githubRepo.GetPackageByVersion(ctx, owner, packageName)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(
		packages,
	)
}

func (h *Handler) GetServicesByApp(w http.ResponseWriter, r *http.Request) {
	appID := r.URL.Query().Get("app_id")
	if appID == "" {
		http.Error(w, "app_id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	services, err := h.mongoRepo.ListServicesByApp(ctx, appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(services)
}
