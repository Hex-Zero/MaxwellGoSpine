package errs

import (
	"errors"
	"github.com/hex-zero/MaxwellGoSpine/internal/core"
	"net/http"
)

func HTTPStatus(err error) int {
	switch {
	case errors.Is(err, core.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, core.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, core.ErrValidation):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
