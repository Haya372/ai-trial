package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Haya372/ai-trial/backend/domain"
	"github.com/Haya372/ai-trial/backend/interface/handler/response"
)

// domainErrorResponse is the HTTP status/code/message a domain error code
// maps to.
type domainErrorResponse struct {
	status  int
	code    string
	message string
}

// writeDomainError maps a *domain.ValidationError to a 400 with field
// details, and a *domain.DomainError to whatever mapping registers for its
// Code(); anything else, including an unmapped domain error code, is logged
// and reported as a 500. Shared by every handler's usecase-error translation
// so a new domain error code needs updating in exactly one mapping.
func writeDomainError(
	w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error, mapping map[string]domainErrorResponse,
) {
	var ve *domain.ValidationError
	var de *domain.DomainError

	switch {
	case errors.As(err, &ve):
		details := make([]response.ErrorDetail, len(ve.Details))
		for i, d := range ve.Details {
			details[i] = response.ErrorDetail{Field: d.Field, Code: d.Code, Message: d.Message}
		}
		response.WriteValidationError(w, details)
	case errors.As(err, &de):
		if m, ok := mapping[de.Code()]; ok {
			response.WriteError(w, m.status, m.code, m.message)
			return
		}
		logger.Error("unexpected domain error", "error", err, "path", r.URL.Path)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
	default:
		logger.Error("internal error", "error", err, "path", r.URL.Path)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
	}
}
