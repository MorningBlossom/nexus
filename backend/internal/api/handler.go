package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MorningBlossom/nexus/backend/internal/github"
	"github.com/MorningBlossom/nexus/backend/internal/models"
	"github.com/MorningBlossom/nexus/backend/internal/repository"
	"github.com/MorningBlossom/nexus/backend/internal/security"
	"go.mongodb.org/mongo-driver/bson/primitive"
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


// environmentVariableResponse never exposes a secret's encrypted value.
type environmentVariableResponse struct {
	Key      string `json:"key"`
	Value    string `json:"value,omitempty"`
	IsSecret bool   `json:"is_secret"`
	HasValue bool   `json:"has_value"`
}

// environmentResponse deliberately excludes legacy and variable secret values.
type environmentResponse struct {
	ID        string    `json:"id"`
	AppID     string    `json:"app_id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	HasSecret bool                         `json:"has_secret"`
	Variables []environmentVariableResponse `json:"variables"`
	Service   string                       `json:"service"`
	Artifact  string    `json:"artifact"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toEnvironmentResponse(environment *models.Environment) environmentResponse {
	variables := make([]environmentVariableResponse, 0, len(environment.Variables))
	for _, variable := range environment.Variables {
		responseVariable := environmentVariableResponse{
			Key: variable.Key, IsSecret: variable.IsSecret, HasValue: variable.Value != "",
		}
		if !variable.IsSecret {
			responseVariable.Value = variable.Value
		}
		variables = append(variables, responseVariable)
	}
	return environmentResponse{
		ID: environment.ID.Hex(), AppID: environment.AppID, Name: environment.Name,
		URL: environment.URL, HasSecret: environment.Secret != "", Variables: variables,
		Service: environment.Service, Artifact: environment.Artifact,
		CreatedAt: environment.CreatedAt, UpdatedAt: environment.UpdatedAt,
	}
}

func (h *Handler) GetEnvironments(w http.ResponseWriter, r *http.Request) {
	appID := strings.TrimSpace(r.URL.Query().Get("app_id"))
	if appID == "" {
		http.Error(w, "app_id is required", http.StatusBadRequest)
		return
	}

	environments, err := h.mongoRepo.ListEnvironmentsByApp(r.Context(), appID)
	if err != nil {
		http.Error(w, "failed to list environments", http.StatusInternalServerError)
		return
	}

	response := make([]environmentResponse, 0, len(environments))
	for _, environment := range environments {
		response = append(response, toEnvironmentResponse(environment))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

type saveEnvironmentRequest struct {
	ID       string `json:"id"`
	AppID    string `json:"app_id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Secret    string                    `json:"secret"`
	Variables []models.EnvironmentVariable `json:"variables"`
	Service   string                    `json:"service"`
	Artifact string `json:"artifact"`
}

func (h *Handler) SaveEnvironment(w http.ResponseWriter, r *http.Request) {
	var input saveEnvironmentRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	input.AppID = strings.TrimSpace(input.AppID)
	input.Name = strings.TrimSpace(input.Name)
	input.URL = strings.TrimSpace(input.URL)
	input.Service = strings.TrimSpace(input.Service)
	if input.AppID == "" || input.Name == "" || input.Service == "" {
		http.Error(w, "app_id, name, and service are required", http.StatusBadRequest)
		return
	}
	if _, err := h.mongoRepo.GetApp(r.Context(), input.AppID); err != nil {
		http.Error(w, "app not found", http.StatusNotFound)
		return
	}

	environment := &models.Environment{
		AppID: input.AppID, Name: input.Name, URL: input.URL,
		Service: input.Service, Artifact: input.Artifact,
	}
	if input.ID != "" {
		id, err := primitive.ObjectIDFromHex(input.ID)
		if err != nil {
			http.Error(w, "invalid environment id", http.StatusBadRequest)
			return
		}
		environment.ID = id
		existing, err := h.mongoRepo.GetEnvironment(r.Context(), input.AppID, id)
		if err != nil {
			http.Error(w, "environment not found", http.StatusNotFound)
			return
		}
		environment.CreatedAt = existing.CreatedAt
		environment.Secret = existing.Secret
	}

	// Preserve existing secret variable values when the editor submits them blank.
	if input.ID != "" {
		existing, err := h.mongoRepo.GetEnvironment(r.Context(), input.AppID, environment.ID)
		if err != nil {
			http.Error(w, "environment not found", http.StatusNotFound)
			return
		}
		existingVariables := make(map[string]models.EnvironmentVariable, len(existing.Variables))
		for _, variable := range existing.Variables {
			existingVariables[variable.Key] = variable
		}
		key := []byte(os.Getenv("NEXUS_ENCRYPTION_KEY"))
		for i := range input.Variables {
			input.Variables[i].Key = strings.TrimSpace(input.Variables[i].Key)
			if input.Variables[i].Key == "" {
				if strings.TrimSpace(input.Variables[i].Value) != "" {
					http.Error(w, "variable key is required when a value is provided", http.StatusBadRequest)
					return
				}
				continue
			}
			if input.Variables[i].IsSecret && input.Variables[i].Value == "" {
				if previous, ok := existingVariables[input.Variables[i].Key]; ok && previous.IsSecret {
					input.Variables[i].Value = previous.Value
				}
			}
		}
	}

	// Encrypt only populated secret values; ordinary configuration does not need a key.
	key := []byte(os.Getenv("NEXUS_ENCRYPTION_KEY"))
	for i := range input.Variables {
		variable := &input.Variables[i]
		variable.Key = strings.TrimSpace(variable.Key)
		if variable.Key == "" && strings.TrimSpace(variable.Value) == "" {
			continue
		}
		if variable.Key == "" {
			http.Error(w, "variable key is required when a value is provided", http.StatusBadRequest)
			return
		}
		if variable.IsSecret && variable.Value != "" {
			if len(key) != 16 && len(key) != 24 && len(key) != 32 {
				http.Error(w, "NEXUS_ENCRYPTION_KEY must be 16, 24, or 32 bytes to save a secret", http.StatusInternalServerError)
				return
			}
			encrypted, err := security.Encrypt(variable.Value, key)
			if err != nil {
				http.Error(w, "failed to encrypt environment variable secret", http.StatusInternalServerError)
				return
			}
			variable.Value = encrypted
		}
	}
	environment.Variables = make([]models.EnvironmentVariable, 0, len(input.Variables))
	for _, variable := range input.Variables {
		if variable.Key != "" {
			environment.Variables = append(environment.Variables, variable)
		}
	}

	if input.Secret != "" {
		key := []byte(os.Getenv("NEXUS_ENCRYPTION_KEY"))
		if len(key) != 16 && len(key) != 24 && len(key) != 32 {
			http.Error(w, "NEXUS_ENCRYPTION_KEY must be 16, 24, or 32 bytes to save a secret", http.StatusInternalServerError)
			return
		}
		encrypted, err := security.Encrypt(input.Secret, key)
		if err != nil {
			http.Error(w, "failed to encrypt environment secret", http.StatusInternalServerError)
			return
		}
		environment.Secret = encrypted
	}

	if err := h.mongoRepo.SaveEnvironment(r.Context(), environment); err != nil {
		http.Error(w, "failed to save environment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toEnvironmentResponse(environment))
}

func (h *Handler) DeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	appID := strings.TrimSpace(r.URL.Query().Get("app_id"))
	idValue := strings.TrimSpace(r.URL.Query().Get("id"))
	if appID == "" || idValue == "" {
		http.Error(w, "app_id and id are required", http.StatusBadRequest)
		return
	}
	id, err := primitive.ObjectIDFromHex(idValue)
	if err != nil {
		http.Error(w, "invalid environment id", http.StatusBadRequest)
		return
	}
	if err := h.mongoRepo.DeleteEnvironment(r.Context(), appID, id); err != nil {
		http.Error(w, "environment not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RestoreEnvironment(w http.ResponseWriter, r *http.Request) {
	appID := strings.TrimSpace(r.URL.Query().Get("app_id"))
	idValue := strings.TrimSpace(r.URL.Query().Get("id"))
	if appID == "" || idValue == "" {
		http.Error(w, "app_id and id are required", http.StatusBadRequest)
		return
	}
	id, err := primitive.ObjectIDFromHex(idValue)
	if err != nil {
		http.Error(w, "invalid environment id", http.StatusBadRequest)
		return
	}
	if err := h.mongoRepo.RestoreEnvironment(r.Context(), appID, id); err != nil {
		http.Error(w, "environment not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
