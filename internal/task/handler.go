package task

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repo *Repo
}

func NewHandler(repo *Repo) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)          // GET /tasks
	r.Post("/", h.create)       // POST /tasks
	r.Get("/{id}", h.get)       // GET /tasks/{id}
	r.Put("/{id}", h.update)    // PUT /tasks/{id}
	r.Delete("/{id}", h.delete) // DELETE /tasks/{id}
	return r
}

func parsePositiveInt(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return def
	}
	return n
}

type createReq struct {
	Title string `json:"title"`
}

type updateReq struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func validateTitle(title string) error {
	length := utf8.RuneCountInString(title)
	log.Println("title:", title, "length:", length)
	if length < 3 || length > 100 {
		return fmt.Errorf("title must be 3..100 chars")
	}
	return nil
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := parsePositiveInt(q.Get("page"), 1)
	limit := parsePositiveInt(q.Get("limit"), 0) // 0 => без лимита
	var (
		filterDone *bool
	)
	if raw := q.Get("done"); raw != "" {
		if b, err := strconv.ParseBool(raw); err == nil {
			filterDone = &b
		} else {
			httpError(w, http.StatusBadRequest, "bad 'done' query param")
			return
		}
	}

	all := h.repo.List()

	// фильтр
	filtered := all[:0]
	for _, t := range all {
		if filterDone != nil && t.Done != *filterDone {
			continue
		}
		filtered = append(filtered, t)
	}

	//сортировка
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ID < filtered[j].ID
	})

	// пагинация
	if limit > 0 {
		if page < 1 {
			page = 1
		}
		start := (page - 1) * limit
		if start >= len(filtered) {
			writeJSON(w, http.StatusOK, []Task{}) // пустой список
			return
		}
		end := start + limit
		if end > len(filtered) {
			end = len(filtered)
		}
		filtered = filtered[start:end]
	}

	writeJSON(w, http.StatusOK, filtered)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, bad := parseID(w, r)
	if bad {
		return
	}
	t, err := h.repo.Get(id)
	if err != nil {
		httpError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		httpError(w, http.StatusBadRequest, "invalid json: require non-empty title")
		return
	}
	if err := validateTitle(req.Title); err != nil {
		httpError(w, http.StatusUnprocessableEntity, err.Error()) // 422
		return
	}
	t := h.repo.Create(req.Title)
	writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, bad := parseID(w, r)
	if bad {
		return
	}
	var req updateReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		httpError(w, http.StatusBadRequest, "invalid json: require non-empty title")
		return
	}
	if err := validateTitle(req.Title); err != nil {
		httpError(w, http.StatusUnprocessableEntity, err.Error()) // 422
		return
	}
	t, err := h.repo.Update(id, req.Title, req.Done)
	if err != nil {
		httpError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, bad := parseID(w, r)
	if bad {
		return
	}
	if err := h.repo.Delete(id); err != nil {
		httpError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// helpers

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	raw := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		httpError(w, http.StatusBadRequest, "invalid id")
		return 0, true
	}
	return id, false
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func httpError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
