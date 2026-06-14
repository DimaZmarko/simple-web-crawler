package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog"

	"github.com/DimaZmarko/simple-web-crawler/apps/api/internal/api/types"
	"github.com/DimaZmarko/simple-web-crawler/apps/api/internal/crawl"
)

// CrawlService is the subset of the crawl service the handlers depend on.
type CrawlService interface {
	Create(ctx context.Context, in crawl.CreateInput) (crawl.Crawl, error)
	Get(ctx context.Context, id uuid.UUID) (crawl.Crawl, error)
	List(ctx context.Context, token string, limit int) (crawl.Page, error)
}

// CrawlHandler holds the crawl HTTP handlers.
type CrawlHandler struct {
	svc    CrawlService
	logger zerolog.Logger
}

// NewCrawlHandler builds a CrawlHandler over the given service.
func NewCrawlHandler(svc CrawlService, logger zerolog.Logger) *CrawlHandler {
	return &CrawlHandler{svc: svc, logger: logger}
}

// Create handles POST /crawls.
//
// @Summary  Submit a crawl job
// @Tags     crawls
// @Accept   json
// @Produce  json
// @Param    request body types.CreateCrawlRequest true "Crawl configuration"
// @Success  202 {object} types.Crawl
// @Failure  400 {object} types.ValidationError
// @Router   /crawls [post]
func (h *CrawlHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body types.CreateCrawlRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, types.ValidationError{
			Message: "Request validation failed",
			Errors: []types.FieldError{
				{Field: "body", Message: "must be a valid JSON object matching the schema"},
			},
		})
		return
	}

	c, err := h.svc.Create(r.Context(), crawl.CreateInput{
		SeedURL:  body.SeedUrl,
		MaxDepth: body.MaxDepth,
		MaxPages: body.MaxPages,
	})
	if err != nil {
		var verr *crawl.ValidationError
		if errors.As(err, &verr) {
			writeJSON(w, http.StatusBadRequest, toValidationError(verr))
			return
		}
		h.writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusAccepted, toAPICrawl(c))
}

// List handles GET /crawls.
//
// @Summary  List crawl jobs (cursor-paginated, newest first)
// @Tags     crawls
// @Produce  json
// @Param    cursor query string false "Opaque base64url pagination token"
// @Param    limit  query int    false "Maximum number of items to return"
// @Success  200 {object} types.CrawlList
// @Failure  400 {object} types.ValidationError
// @Router   /crawls [get]
func (h *CrawlHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := 0
	if raw := q.Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > crawl.MaxLimit {
			writeJSON(w, http.StatusBadRequest, types.ValidationError{
				Message: "Request validation failed",
				Errors: []types.FieldError{
					{Field: "limit", Message: fmt.Sprintf("must be an integer between 1 and %d", crawl.MaxLimit)},
				},
			})
			return
		}
		limit = n
	}

	page, err := h.svc.List(r.Context(), q.Get("cursor"), limit)
	if err != nil {
		if errors.Is(err, crawl.ErrInvalidCursor) {
			writeJSON(w, http.StatusBadRequest, types.ValidationError{
				Message: "Request validation failed",
				Errors: []types.FieldError{
					{Field: "cursor", Message: "is not a valid pagination token"},
				},
			})
			return
		}
		h.writeInternalError(w, err)
		return
	}

	items := make([]types.CrawlSummary, 0, len(page.Items))
	for _, c := range page.Items {
		items = append(items, types.CrawlSummary{
			Id:        openapi_types.UUID(c.ID),
			SeedUrl:   c.SeedURL,
			Status:    types.CrawlStatus(c.Status),
			CreatedAt: c.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, types.CrawlList{
		Items:      items,
		NextCursor: page.NextCursor,
	})
}

// Get handles GET /crawls/{id}.
//
// @Summary  Fetch a single crawl job by id
// @Tags     crawls
// @Produce  json
// @Param    id path string true "Crawl id" format(uuid)
// @Success  200 {object} types.Crawl
// @Failure  404 {object} types.NotFoundError
// @Router   /crawls/{id} [get]
func (h *CrawlHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, types.NotFoundError{Message: "crawl not found"})
		return
	}

	c, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, crawl.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, types.NotFoundError{Message: "crawl not found"})
			return
		}
		h.writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toAPICrawl(c))
}

func toAPICrawl(c crawl.Crawl) types.Crawl {
	return types.Crawl{
		Id:        openapi_types.UUID(c.ID),
		SeedUrl:   c.SeedURL,
		MaxDepth:  c.MaxDepth,
		MaxPages:  c.MaxPages,
		Status:    types.CrawlStatus(c.Status),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func toValidationError(verr *crawl.ValidationError) types.ValidationError {
	fields := make([]types.FieldError, 0, len(verr.Errors))
	for _, f := range verr.Errors {
		fields = append(fields, types.FieldError{Field: f.Field, Message: f.Message})
	}
	return types.ValidationError{
		Message: "Request validation failed",
		Errors:  fields,
	}
}

// writeInternalError logs the cause and emits an opaque 500 so internals never
// leak to clients.
func (h *CrawlHandler) writeInternalError(w http.ResponseWriter, err error) {
	h.logger.Error().Err(err).Msg("crawl handler internal error")
	writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal server error"})
}
