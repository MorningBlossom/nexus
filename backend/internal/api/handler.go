package api

import (
	"encoding/json"
	"net/http"

	"github.com/MorningBlossom/nexus/backend/internal/github"
)

type Handler struct {
	repository *github.Repository
}

func NewHandler(repo *github.Repository) *Handler {
	return &Handler{repository: repo}
}

func (h *Handler) GetInstallationRepositories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.repository.GetInstallationRepositories(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

}

func (h *Handler) GetPackages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	owner := "MorningBlossom"
	packageName := "ai-code-review"
	//if packageName == "" {
	//	http.Error(w, "package name is required: use /getPackages?name=ai-code-review", http.StatusBadRequest)
	//	return
	//}

	packages, err := h.repository.GetPackages(ctx, owner, packageName)

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
