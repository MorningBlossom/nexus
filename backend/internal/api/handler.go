package api

import (
	"encoding/json"
	"errors"
	"os"
	"time"
	"net/http"
	"strings"

	"github.com/MorningBlossom/nexus/backend/internal/github"
	"github.com/MorningBlossom/nexus/backend/internal/models"
	"github.com/MorningBlossom/nexus/backend/internal/repository"
	"github.com/MorningBlossom/nexus/backend/internal/security"
	"go.mongodb.org/mongo-driver/mongo"
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

	appID := r.URL.Query().Get("app_id")
	if appID == "" {
		http.Error(w, "app_id is required", http.StatusBadRequest)
		return
	}

	app, err := h.mongoRepo.GetApp(ctx, appID)
	if err != nil {
		http.Error(w, "app not found", http.StatusNotFound)
		return
	}

	owner := "MorningBlossom"
	repoName := app.Repo
	if repoName == "" {
		repoName = app.AppID
	}
	if parts := strings.SplitN(repoName, "/", 2); len(parts) == 2 {
		if parts[0] == owner {
			repoName = parts[1]
		} else {
			repoName = parts[1]
		}
	}

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

type environmentVariableInput struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Secret bool   `json:"secret"`
}

type environmentInput struct {
	ID        string                    `json:"id"`
	AppID     string                    `json:"appId"`
	Name      string                    `json:"name"`
	URL       string                    `json:"url"`
	Service   string                    `json:"service"`
	Artifact  string                    `json:"artifact"`
	Variables []environmentVariableInput `json:"variables"`
}

type environmentVariableResponse struct {
	Key      string `json:"key"`
	Value    string `json:"value,omitempty"`
	Secret   bool   `json:"secret"`
	HasValue bool   `json:"hasValue"`
}

type environmentResponse struct {
	ID        string                        `json:"id"`
	AppID     string                        `json:"appId"`
	Name      string                        `json:"name"`
	URL       string                        `json:"url"`
	Service   string                        `json:"service"`
	Artifact  string                        `json:"artifact"`
	Variables []environmentVariableResponse `json:"variables"`
	CreatedAt time.Time                     `json:"createdAt"`
	UpdatedAt time.Time                     `json:"updatedAt"`
}

func environmentDTO(environment *models.Environment) environmentResponse {
	out := environmentResponse{
		ID: environment.ID.Hex(), AppID: environment.AppID, Name: environment.Name,
		URL: environment.URL, Service: environment.Service, Artifact: environment.Artifact,
		CreatedAt: environment.CreatedAt, UpdatedAt: environment.UpdatedAt,
		Variables: make([]environmentVariableResponse, 0, len(environment.Variables)),
	}
	for _, variable := range environment.Variables {
		value := variable.Value
		if variable.Secret {
			value = ""
		}
		out.Variables = append(out.Variables, environmentVariableResponse{
			Key: variable.Key, Value: value, Secret: variable.Secret, HasValue: variable.HasValue,
		})
	}
	return out
}

func (h *Handler) GetEnvironments(w http.ResponseWriter, r *http.Request) {
	appID := strings.TrimSpace(r.URL.Query().Get("app_id"))
	if appID == "" {
		http.Error(w, "app_id is required", http.StatusBadRequest)
		return
	}
	if _, err := h.mongoRepo.GetApp(r.Context(), appID); err != nil {
		http.Error(w, "app not found", http.StatusNotFound)
		return
	}
	environments, err := h.mongoRepo.ListEnvironmentsByApp(r.Context(), appID)
	if err != nil {
		http.Error(w, "failed to list environments", http.StatusInternalServerError)
		return
	}
	response := make([]environmentResponse, 0, len(environments))
	for _, environment := range environments {
		response = append(response, environmentDTO(environment))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *Handler) SaveEnvironment(w http.ResponseWriter, r *http.Request) {
	var input environmentInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&input); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	input.AppID = strings.TrimSpace(input.AppID)
	input.Name = strings.TrimSpace(input.Name)
	input.URL = strings.TrimSpace(input.URL)
	input.Service = strings.TrimSpace(input.Service)
	input.Artifact = strings.TrimSpace(input.Artifact)
	if input.AppID == "" || input.Name == "" || input.Service == "" {
		http.Error(w, "appId, name, and service are required", http.StatusBadRequest)
		return
	}
	if _, err := h.mongoRepo.GetApp(r.Context(), input.AppID); err != nil {
		http.Error(w, "app not found", http.StatusNotFound)
		return
	}

	now := time.Now().UTC()
	environment := &models.Environment{
		AppID: input.AppID, Name: input.Name, URL: input.URL, Service: input.Service,
		Artifact: input.Artifact, CreatedAt: now, UpdatedAt: now,
	}
	var existing *models.Environment
	if input.ID != "" {
		var err error
		existing, err = h.mongoRepo.GetEnvironment(r.Context(), input.AppID, input.ID)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				http.Error(w, "environment not found", http.StatusNotFound)
			} else {
				http.Error(w, "invalid environment id", http.StatusBadRequest)
			}
			return
		}
		environment.ID = existing.ID
		environment.CreatedAt = existing.CreatedAt
	}

	secretKey := []byte(os.Getenv("NEXUS_ENCRYPTION_KEY"))
	oldVariables := make(map[string]models.EnvironmentVariable)
	if existing != nil {
		for _, variable := range existing.Variables {
			oldVariables[variable.Key] = variable
		}
	}
	seenKeys := make(map[string]bool)
	for _, inputVariable := range input.Variables {
		key := strings.TrimSpace(inputVariable.Key)
		if key == "" || seenKeys[key] {
			http.Error(w, "environment variable keys must be non-empty and unique", http.StatusBadRequest)
			return
		}
		seenKeys[key] = true
		variable := models.EnvironmentVariable{Key: key, Secret: inputVariable.Secret}
		if inputVariable.Secret {
			if strings.TrimSpace(inputVariable.Value) == "" {
				if old, ok := oldVariables[key]; ok && old.Secret && old.HasValue {
					variable.Value, variable.HasValue = old.Value, true
				} else {
					variable.HasValue = false
				}
			} else {
				if len(secretKey) != 16 && len(secretKey) != 24 && len(secretKey) != 32 {
					http.Error(w, "NEXUS_ENCRYPTION_KEY must be 16, 24, or 32 bytes to save secret variables", http.StatusInternalServerError)
					return
				}
				encrypted, err := security.Encrypt(inputVariable.Value, secretKey)
				if err != nil {
					http.Error(w, "failed to encrypt secret variable", http.StatusInternalServerError)
					return
				}
				variable.Value, variable.HasValue = encrypted, true
			}
		} else {
			variable.Value = inputVariable.Value
			variable.HasValue = inputVariable.Value != ""
		}
		environment.Variables = append(environment.Variables, variable)
	}
	if err := h.mongoRepo.SaveEnvironment(r.Context(), environment); err != nil {
		http.Error(w, "failed to save environment", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(environmentDTO(environment))
}

func (h *Handler) DeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	appID := strings.TrimSpace(r.URL.Query().Get("app_id"))
	environmentID := strings.TrimSpace(r.URL.Query().Get("id"))
	if appID == "" || environmentID == "" {
		http.Error(w, "app_id and id are required", http.StatusBadRequest)
		return
	}
	if err := h.mongoRepo.DeleteEnvironment(r.Context(), appID, environmentID); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			http.Error(w, "environment not found", http.StatusNotFound)
		} else {
			http.Error(w, "failed to delete environment", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
