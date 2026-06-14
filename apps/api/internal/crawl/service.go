// Package crawl holds the crawl domain: validation, persistence, and the
// service layer that handlers delegate to. No HTTP concerns live here.
package crawl

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/DimaZmarko/simple-web-crawler/apps/api/internal/db"
)

// Pagination bounds for ListCrawls.
const (
	defaultLimit = 20
	maxLimit     = 100
)

// maxPagesCeiling is the upper bound on max_pages a caller may request.
const maxPagesCeiling = 10000

// Sentinel errors. Handlers map these to HTTP status codes in one place.
var (
	// ErrNotFound is returned when a crawl id does not exist.
	ErrNotFound = errors.New("crawl not found")
	// ErrInvalidCursor is returned when a pagination token cannot be decoded.
	ErrInvalidCursor = errors.New("invalid cursor")
)

// FieldError is a single field-level validation failure.
type FieldError struct {
	Field   string
	Message string
}

// ValidationError aggregates one or more field-level failures.
type ValidationError struct {
	Errors []FieldError
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %d field error(s)", len(e.Errors))
}

// Querier is the subset of the generated db.Queries the service depends on.
// Defining it here keeps the service testable and decoupled from the pool.
type Querier interface {
	CreateCrawl(ctx context.Context, arg db.CreateCrawlParams) (db.Crawl, error)
	GetCrawlByID(ctx context.Context, id pgtype.UUID) (db.Crawl, error)
	ListCrawlsFirstPage(ctx context.Context, limit int32) ([]db.Crawl, error)
	ListCrawlsAfterCursor(ctx context.Context, arg db.ListCrawlsAfterCursorParams) ([]db.Crawl, error)
}

// Service implements the crawl use cases over a Querier.
type Service struct {
	q Querier
}

// NewService builds a Service backed by the given querier.
func NewService(q Querier) *Service {
	return &Service{q: q}
}

// Crawl is the domain representation of a crawl row, with native Go types.
type Crawl struct {
	ID        uuid.UUID
	SeedURL   string
	MaxDepth  int
	MaxPages  int
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateInput is the validated-on-the-way-in payload for creating a crawl.
type CreateInput struct {
	SeedURL  string
	MaxDepth int
	MaxPages int
}

// Create validates the input and inserts a queued crawl. Returns a
// *ValidationError when the input is invalid.
func (s *Service) Create(ctx context.Context, in CreateInput) (Crawl, error) {
	if verr := validateCreate(in); verr != nil {
		return Crawl{}, verr
	}

	row, err := s.q.CreateCrawl(ctx, db.CreateCrawlParams{
		SeedUrl:  in.SeedURL,
		MaxDepth: int32(in.MaxDepth),
		MaxPages: int32(in.MaxPages),
	})
	if err != nil {
		return Crawl{}, fmt.Errorf("create crawl: %w", err)
	}

	return toDomain(row), nil
}

// Get fetches a single crawl by id, returning ErrNotFound when absent.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (Crawl, error) {
	row, err := s.q.GetCrawlByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Crawl{}, ErrNotFound
		}
		return Crawl{}, fmt.Errorf("get crawl: %w", err)
	}
	return toDomain(row), nil
}

// Page is a slice of crawls plus an optional cursor to the next page.
type Page struct {
	Items      []Crawl
	NextCursor *string
}

// List returns a page of crawls newest-first. token is the opaque cursor from a
// previous call (empty for the first page); limit is clamped to [1, maxLimit]
// and defaults to defaultLimit when <= 0. A malformed token yields
// ErrInvalidCursor.
func (s *Service) List(ctx context.Context, token string, limit int) (Page, error) {
	limit = clampLimit(limit)

	// Fetch one extra row to detect whether a further page exists.
	fetch := int32(limit + 1)

	var rows []db.Crawl
	var err error
	if token == "" {
		rows, err = s.q.ListCrawlsFirstPage(ctx, fetch)
	} else {
		var cur cursor
		cur, err = decodeCursor(token)
		if err != nil {
			return Page{}, err
		}
		rows, err = s.q.ListCrawlsAfterCursor(ctx, db.ListCrawlsAfterCursorParams{
			CursorCreatedAt: toPgTimestamptz(cur.CreatedAt),
			CursorID:        toPgUUID(cur.ID),
			PageLimit:       fetch,
		})
	}
	if err != nil {
		return Page{}, fmt.Errorf("list crawls: %w", err)
	}

	page := Page{Items: make([]Crawl, 0, len(rows))}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	for _, r := range rows {
		page.Items = append(page.Items, toDomain(r))
	}

	if hasMore && len(page.Items) > 0 {
		last := page.Items[len(page.Items)-1]
		next := encodeCursor(cursor{CreatedAt: last.CreatedAt, ID: last.ID})
		page.NextCursor = &next
	}

	return page, nil
}

func validateCreate(in CreateInput) *ValidationError {
	var fields []FieldError

	if !isAbsoluteHTTPURL(in.SeedURL) {
		fields = append(fields, FieldError{
			Field:   "seedUrl",
			Message: "must be an absolute http or https URL",
		})
	}

	if in.MaxDepth < 0 {
		fields = append(fields, FieldError{
			Field:   "maxDepth",
			Message: "must be at least 0",
		})
	}

	switch {
	case in.MaxPages < 1:
		fields = append(fields, FieldError{
			Field:   "maxPages",
			Message: "must be at least 1",
		})
	case in.MaxPages > maxPagesCeiling:
		fields = append(fields, FieldError{
			Field:   "maxPages",
			Message: fmt.Sprintf("must be at most %d", maxPagesCeiling),
		})
	}

	if len(fields) == 0 {
		return nil
	}
	return &ValidationError{Errors: fields}
}

// isAbsoluteHTTPURL reports whether raw parses as an absolute http/https URL
// with a host.
func isAbsoluteHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}

func clampLimit(limit int) int {
	switch {
	case limit <= 0:
		return defaultLimit
	case limit > maxLimit:
		return maxLimit
	default:
		return limit
	}
}

func toDomain(r db.Crawl) Crawl {
	return Crawl{
		ID:        r.ID.Bytes,
		SeedURL:   r.SeedUrl,
		MaxDepth:  int(r.MaxDepth),
		MaxPages:  int(r.MaxPages),
		Status:    r.Status,
		CreatedAt: r.CreatedAt.Time,
		UpdatedAt: r.UpdatedAt.Time,
	}
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func toPgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
