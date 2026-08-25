// Package httpapi exposes the cage application service as a small JSON HTTP API.
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"example.com/animalcage/internal/model"
	"example.com/animalcage/internal/workflow"
)

// Handler routes requests to the workflow service. It has no process-global
// state, which keeps embedded deployments and deterministic tests independent.
type Handler struct{ service *workflow.Service }

func NewHandler(service *workflow.Service) *Handler { return &Handler{service: service} }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}
	if r.URL.Path == "/api/applications" {
		h.applications(w, r)
		return
	}
	if r.URL.Path == "/api/import" {
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		h.importApplications(w, r)
		return
	}
	const prefix = "/api/applications/"
	if strings.HasPrefix(r.URL.Path, prefix) {
		h.application(w, r, strings.TrimPrefix(r.URL.Path, prefix))
		return
	}
	h.error(w, http.StatusNotFound, errors.New("endpoint not found"))
}

type registerRequest struct {
	ApplicationNo string   `json:"application_no"`
	Applicant     string   `json:"applicant"`
	Facility      string   `json:"facility"`
	Species       string   `json:"species"`
	CageCount     int      `json:"cage_count"`
	Roster        []string `json:"roster"`
	Actor         string   `json:"actor"`
}

func (h *Handler) applications(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := h.service.Search(model.SearchFilter{Applicant: r.URL.Query().Get("applicant"), Facility: r.URL.Query().Get("facility"), Species: r.URL.Query().Get("species"), Status: r.URL.Query().Get("status")})
		if err != nil {
			h.error(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"applications": rows, "count": len(rows)})
	case http.MethodPost:
		var request registerRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		row, err := h.service.Register(workflow.RegistrationInput{ApplicationNo: request.ApplicationNo, Applicant: request.Applicant, Facility: request.Facility, Species: request.Species, CageCount: request.CageCount, Roster: request.Roster}, request.Actor)
		if err != nil {
			h.error(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, row)
	default:
		methodNotAllowed(w)
	}
}

func (h *Handler) application(w http.ResponseWriter, r *http.Request, path string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		h.error(w, http.StatusNotFound, errors.New("application id is required"))
		return
	}
	id := parts[0]
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		row, err := h.service.Get(id)
		if err != nil {
			h.error(w, statusForError(err), err)
			return
		}
		writeJSON(w, http.StatusOK, row)
		return
	}
	if parts[1] == "report" && r.Method == http.MethodGet {
		report, err := h.service.Report(id)
		if err != nil {
			h.error(w, statusForError(err), err)
			return
		}
		writeJSON(w, http.StatusOK, report)
		return
	}
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	switch parts[1] {
	case "submit":
		var request actorRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		row, err := h.service.Submit(id, request.Actor)
		if err != nil {
			h.error(w, statusForError(err), err)
			return
		}
		writeJSON(w, http.StatusOK, row)
	case "review":
		var request model.ReviewDecision
		if !decodeJSON(w, r, &request) {
			return
		}
		row, err := h.service.Review(id, request)
		if err != nil {
			h.error(w, statusForError(err), err)
			return
		}
		writeJSON(w, http.StatusOK, row)
	case "confirm-roster":
		var request confirmRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		row, err := h.service.ConfirmRoster(id, request.Position, request.Name, request.Actor)
		if err != nil {
			h.error(w, statusForError(err), err)
			return
		}
		writeJSON(w, http.StatusOK, row)
	case "change":
		var request model.ChangeRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		row, err := h.service.Change(id, request)
		if err != nil {
			h.error(w, statusForError(err), err)
			return
		}
		writeJSON(w, http.StatusOK, row)
	case "archive":
		var request actorRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		row, err := h.service.Archive(id, request.Actor)
		if err != nil {
			h.error(w, statusForError(err), err)
			return
		}
		writeJSON(w, http.StatusOK, row)
	case "report":
		if r.URL.Query().Get("publish") == "true" {
			var request actorRequest
			if !decodeJSON(w, r, &request) {
				return
			}
			attachment, err := h.service.PublishReport(id, request.Actor)
			if err != nil {
				h.error(w, statusForError(err), err)
				return
			}
			writeJSON(w, http.StatusOK, attachment)
			return
		}
		h.error(w, http.StatusMethodNotAllowed, errors.New("report requires GET or publish=true"))
	default:
		h.error(w, http.StatusNotFound, errors.New("application action not found"))
	}
}

type actorRequest struct {
	Actor string `json:"actor"`
}
type confirmRequest struct {
	Position int    `json:"position"`
	Name     string `json:"name"`
	Actor    string `json:"actor"`
}

type importRequest struct {
	Rows  []model.ImportRow `json:"rows"`
	Actor string            `json:"actor"`
}

func (h *Handler) importApplications(w http.ResponseWriter, r *http.Request) {
	var request importRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	result, err := h.service.Import(request.Rows, request.Actor)
	if err != nil {
		h.error(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
		return false
	}
	return true
}

func statusForError(err error) int {
	if strings.Contains(strings.ToLower(err.Error()), "not found") {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

func (h *Handler) error(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		fmt.Fprint(w, "\n")
	}
}
