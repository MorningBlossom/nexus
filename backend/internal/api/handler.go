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


// Secret values are decrypted only for the authenticated environment-management API response.
type environmentVariableResponse struct {
	Key      string `json:"key"`
	Value    string `json:"value,omitempty"`
	IsSecret bool   `json:"is_secret"`
	HasValue bool   `json:"has_value"`
}

type environmentResponse struct {
	ID        string    `json:"id"`
	AppID     string    `json:"app_id"`
	Name      string    `json:"name"`
	Secret    string    `json:"secret,omitempty"`
	HasSecret bool      `json:"has_secret"`
	Variables []environmentVariableResponse `json:"variables"`
	Service   string    `json:"service"`
	Artifact  string    `json:"artifact"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toEnvironmentResponse(environment *models.Environment, key []byte) (environmentResponse, error) {
	variables := make([]environmentVariableResponse, 0, len(environment.Variables))
	hasSecret := environment.Secret != ""
	response := environmentResponse{}
	if environment.Secret != "" {
		secret, err := security.Decrypt(environment.Secret, key)
		if err != nil {
			return response, err
		}
		response.Secret = secret
	}
	for _, variable := range environment.Variables {
		responseVariable := environmentVariableResponse{
			Key: variable.Key, IsSecret: variable.IsSecret, HasValue: variable.Value != "",
		}
		if variable.IsSecret {
			if variable.Value != "" {
				value, err := security.Decrypt(variable.Value, key)
				if err != nil {
					return response, err
				}
				responseVariable.Value = value
				hasSecret = true
			}
		} else {
			responseVariable.Value = variable.Value
		}
		variables = append(variables, responseVariable)
	}
	return environmentResponse{
		ID: environment.ID.Hex(), AppID: environment.AppID, Name: environment.Name,
		Secret: response.Secret, HasSecret: hasSecret, Variables: variables,
		Service: environment.Service, Artifact: environment.Artifact,
		CreatedAt: environment.CreatedAt, UpdatedAt: environment.UpdatedAt,
	}, nil
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
	key := []byte(os.Getenv("NEXUS_ENCRYPTION_KEY"))
	for _, environment := range environments {
		item, err := toEnvironmentResponse(environment, key)
		if err != nil {
			http.Error(w, "failed to decrypt environment secret", http.StatusInternalServerError)
			return
		}
		response = append(response, item)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

type saveEnvironmentRequest struct {
	ID       string `json:"id"`
	AppID    string `json:"app_id"`
	Name     string `json:"name"`
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
		AppID: input.AppID, Name: input.Name,
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

	// Keep existing secret values when the editor submits them blank.
	existingVariables := make(map[string]models.EnvironmentVariable)
	if input.ID != "" {
		existing, err := h.mongoRepo.GetEnvironment(r.Context(), input.AppID, environment.ID)
		if err != nil {
			http.Error(w, "environment not found", http.StatusNotFound)
			return
		}
		for _, variable := range existing.Variables {
			existingVariables[variable.Key] = variable
		}
	}

	// Encrypt only newly supplied secret values; plain configuration needs no key.
	key := []byte(os.Getenv("NEXUS_ENCRYPTION_KEY"))
	environment.Variables = make([]models.EnvironmentVariable, 0, len(input.Variables))
	for _, variable := range input.Variables {
		variable.Key = strings.TrimSpace(variable.Key)
		if variable.Key == "" && strings.TrimSpace(variable.Value) == "" {
			continue
		}
		if variable.Key == "" {
			http.Error(w, "variable key is required when a value is provided", http.StatusBadRequest)
			return
		}
		if variable.IsSecret {
			if variable.Value == "" {
				if previous, ok := existingVariables[variable.Key]; ok && previous.IsSecret {
					// Reuse ciphertext as-is; do not encrypt it a second time.
					variable.Value = previous.Value
				}
			} else {
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
		environment.Variables = append(environment.Variables, variable)
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
	response, err := toEnvironmentResponse(environment, []byte(os.Getenv("NEXUS_ENCRYPTION_KEY")))
	if err != nil {
		http.Error(w, "failed to decrypt environment secret", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(response)
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
