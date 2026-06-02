package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shashisharma307703/vedantam/internal/onboarder"
	"github.com/shashisharma307703/vedantam/internal/repository"
)

type OnboardHandler struct {
	repo   *repository.Repository
	logger onboarder.Logger
}

func NewOnboardHandler(repo *repository.Repository, logger onboarder.Logger) *OnboardHandler {
	return &OnboardHandler{repo: repo, logger: logger}
}

func (h *OnboardHandler) OnboardExisting(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenantId")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "invalid tenant id", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	tmpDir, err := os.MkdirTemp("", "onboard-*")
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)

	for key, headers := range r.MultipartForm.File {
		for _, header := range headers {
			f, err := header.Open()
			if err != nil {
				http.Error(w, "cannot open file", http.StatusBadRequest)
				return
			}
			defer f.Close()
			out, err := os.Create(filepath.Join(tmpDir, key))
			if err != nil {
				http.Error(w, "server error", http.StatusInternalServerError)
				return
			}
			io.Copy(out, f)
			out.Close()
		}
	}

	upsert := r.URL.Query().Get("upsert") == "true"
	dryRun := r.URL.Query().Get("dryRun") == "true"

	loader := onboarder.NewLoader(h.repo.Pool, h.repo.Queries, upsert, dryRun, h.logger)
	if err := loader.LoadAllFromDirectory(r.Context(), tenantID, tmpDir); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}