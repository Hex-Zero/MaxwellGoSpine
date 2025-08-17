package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/hex-zero/MaxwellGoSpine/internal/core"
	"github.com/hex-zero/MaxwellGoSpine/internal/errs"
	"github.com/hex-zero/MaxwellGoSpine/internal/http/render"
	"io"
	"net/http"
	"strconv"
	"time"
)

type UserHandler struct {
	svc      core.UserService
	validate *validator.Validate
}

func NewUserHandler(svc core.UserService) *UserHandler {
	return &UserHandler{svc: svc, validate: validator.New()}
}

func (h *UserHandler) Register(r chi.Router) {
	r.Get("/users", h.list)
	r.Post("/users", h.create)
	r.Get("/users/{id}", h.get)
	r.Patch("/users/{id}", h.update)
	r.Delete("/users/{id}", h.delete)
}

type userDTO struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type createUserReq struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type updateUserReq struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createUserReq
	if err := decodeJSON(w, r, &req); err != nil {
		render.Problem(w, r, errs.HTTPStatus(err), "Invalid JSON", err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		render.Problem(w, r, http.StatusBadRequest, "Validation Error", err.Error())
		return
	}
	u, err := h.svc.Create(r.Context(), req.Name, req.Email)
	if err != nil {
		render.Problem(w, r, errs.HTTPStatus(err), "Create Failed", err.Error())
		return
	}
	render.JSON(w, r, http.StatusCreated, toDTO(u))
}

func (h *UserHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		render.Problem(w, r, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}
	u, err := h.svc.Get(r.Context(), id)
	if err != nil {
		render.Problem(w, r, errs.HTTPStatus(err), "Get Failed", err.Error())
		return
	}
	render.JSON(w, r, http.StatusOK, toDTO(u))
}

func (h *UserHandler) list(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	users, total, err := h.svc.List(r.Context(), page, pageSize)
	if err != nil {
		render.Problem(w, r, errs.HTTPStatus(err), "List Failed", err.Error())
		return
	}
	out := make([]userDTO, 0, len(users))
	for _, u := range users {
		out = append(out, toDTO(u))
	}
	render.JSON(w, r, http.StatusOK, map[string]any{"data": out, "total": total, "page": page, "page_size": pageSize})
}

func (h *UserHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		render.Problem(w, r, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}
	var req updateUserReq
	if r.Header.Get("Content-Type") == "application/merge-patch+json" {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))
		if err != nil {
			render.Problem(w, r, http.StatusBadRequest, "Read Error", err.Error())
			return
		}
		if err := json.Unmarshal(body, &req); err != nil {
			render.Problem(w, r, http.StatusBadRequest, "Invalid JSON", err.Error())
			return
		}
	} else {
		if err := decodeJSON(w, r, &req); err != nil {
			render.Problem(w, r, http.StatusBadRequest, "Invalid JSON", err.Error())
			return
		}
	}
	u, err := h.svc.Update(r.Context(), id, req.Name, req.Email)
	if err != nil {
		render.Problem(w, r, errs.HTTPStatus(err), "Update Failed", err.Error())
		return
	}
	render.JSON(w, r, http.StatusOK, toDTO(u))
}

func (h *UserHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		render.Problem(w, r, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		render.Problem(w, r, errs.HTTPStatus(err), "Delete Failed", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, name))
}

func parsePagination(r *http.Request) (int, int) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(q.Get("page_size"))
	if size <= 0 || size > 100 {
		size = 20
	}
	return page, size
}

func toDTO(u *core.User) userDTO {
	return userDTO{ID: u.ID, Name: u.Name, Email: u.Email, CreatedAt: u.CreatedAt.Format(time.RFC3339), UpdatedAt: u.UpdatedAt.Format(time.RFC3339)}
}

const maxBody = 1 << 20 // 1MB

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if dec.More() {
		return fmt.Errorf("unexpected data")
	}
	return nil
}
