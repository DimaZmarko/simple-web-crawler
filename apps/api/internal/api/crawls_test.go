package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/DimaZmarko/simple-web-crawler/apps/api/internal/api/types"
	"github.com/DimaZmarko/simple-web-crawler/apps/api/internal/crawl"
)

// fakeCrawlService is a hand-rolled CrawlService double driven by funcs so each
// test case can program the exact behaviour it needs.
type fakeCrawlService struct {
	createFn func(ctx context.Context, in crawl.CreateInput) (crawl.Crawl, error)
	getFn    func(ctx context.Context, id uuid.UUID) (crawl.Crawl, error)
	listFn   func(ctx context.Context, token string, limit int) (crawl.Page, error)
}

func (f fakeCrawlService) Create(ctx context.Context, in crawl.CreateInput) (crawl.Crawl, error) {
	return f.createFn(ctx, in)
}

func (f fakeCrawlService) Get(ctx context.Context, id uuid.UUID) (crawl.Crawl, error) {
	return f.getFn(ctx, id)
}

func (f fakeCrawlService) List(ctx context.Context, token string, limit int) (crawl.Page, error) {
	return f.listFn(ctx, token, limit)
}

func newTestRouter(svc CrawlService) http.Handler {
	return NewRouter(fakePinger{err: nil}, svc, zerolog.Nop())
}

func TestCreateCrawl(t *testing.T) {
	fixedID := uuid.MustParse("7d8f3b2a-1c4e-4a9b-8f6d-2e5c9a1b3d4f")
	fixedTime := time.Date(2026, 6, 14, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		body        string
		svc         CrawlService
		wantStatus  int
		wantStatusF string         // expected status field of a Crawl body, when 202
		wantFields  map[string]int // field -> expected count in a ValidationError body
	}{
		{
			name: "valid request is accepted and queued",
			body: `{"seedUrl":"https://example.com","maxDepth":2,"maxPages":500}`,
			svc: fakeCrawlService{createFn: func(_ context.Context, in crawl.CreateInput) (crawl.Crawl, error) {
				return crawl.Crawl{
					ID:        fixedID,
					SeedURL:   in.SeedURL,
					MaxDepth:  in.MaxDepth,
					MaxPages:  in.MaxPages,
					Status:    "queued",
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				}, nil
			}},
			wantStatus:  http.StatusAccepted,
			wantStatusF: "queued",
		},
		{
			name: "bad url yields a field error",
			body: `{"seedUrl":"not-a-url","maxDepth":2,"maxPages":500}`,
			svc: fakeCrawlService{createFn: func(_ context.Context, _ crawl.CreateInput) (crawl.Crawl, error) {
				return crawl.Crawl{}, &crawl.ValidationError{Errors: []crawl.FieldError{
					{Field: "seedUrl", Message: "must be an absolute http or https URL"},
				}}
			}},
			wantStatus: http.StatusBadRequest,
			wantFields: map[string]int{"seedUrl": 1},
		},
		{
			name: "maxPages zero yields a field error",
			body: `{"seedUrl":"https://example.com","maxDepth":2,"maxPages":0}`,
			svc: fakeCrawlService{createFn: func(_ context.Context, _ crawl.CreateInput) (crawl.Crawl, error) {
				return crawl.Crawl{}, &crawl.ValidationError{Errors: []crawl.FieldError{
					{Field: "maxPages", Message: "must be at least 1"},
				}}
			}},
			wantStatus: http.StatusBadRequest,
			wantFields: map[string]int{"maxPages": 1},
		},
		{
			name: "negative maxDepth yields a field error",
			body: `{"seedUrl":"https://example.com","maxDepth":-1,"maxPages":5}`,
			svc: fakeCrawlService{createFn: func(_ context.Context, _ crawl.CreateInput) (crawl.Crawl, error) {
				return crawl.Crawl{}, &crawl.ValidationError{Errors: []crawl.FieldError{
					{Field: "maxDepth", Message: "must be at least 0"},
				}}
			}},
			wantStatus: http.StatusBadRequest,
			wantFields: map[string]int{"maxDepth": 1},
		},
		{
			name:       "malformed json yields 400",
			body:       `{"seedUrl":`,
			svc:        fakeCrawlService{},
			wantStatus: http.StatusBadRequest,
			wantFields: map[string]int{"body": 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := newTestRouter(tc.svc)

			req := httptest.NewRequest(http.MethodPost, "/crawls", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tc.wantStatus, rec.Body.String())
			}

			if tc.wantStatus == http.StatusAccepted {
				var got types.Crawl
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("decode Crawl: %v", err)
				}
				if string(got.Status) != tc.wantStatusF {
					t.Errorf("status = %q, want %q", got.Status, tc.wantStatusF)
				}
				if got.Id != fixedID {
					t.Errorf("id = %v, want %v", got.Id, fixedID)
				}
				return
			}

			if tc.wantFields != nil {
				var got types.ValidationError
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("decode ValidationError: %v", err)
				}
				counts := map[string]int{}
				for _, fe := range got.Errors {
					counts[fe.Field]++
				}
				for field, want := range tc.wantFields {
					if counts[field] != want {
						t.Errorf("field %q count = %d, want %d (body: %s)", field, counts[field], want, rec.Body.String())
					}
				}
			}
		})
	}
}

func TestGetCrawlNotFound(t *testing.T) {
	tests := []struct {
		name string
		path string
		svc  CrawlService
	}{
		{
			name: "unknown id returns 404",
			path: "/crawls/" + uuid.NewString(),
			svc: fakeCrawlService{getFn: func(_ context.Context, _ uuid.UUID) (crawl.Crawl, error) {
				return crawl.Crawl{}, crawl.ErrNotFound
			}},
		},
		{
			name: "non-uuid id returns 404",
			path: "/crawls/not-a-uuid",
			svc:  fakeCrawlService{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := newTestRouter(tc.svc)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404 (body: %s)", rec.Code, rec.Body.String())
			}

			var got types.NotFoundError
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode NotFoundError: %v", err)
			}
			if got.Message == "" {
				t.Error("NotFoundError.message is empty")
			}
		})
	}
}

func TestListCrawls(t *testing.T) {
	id := uuid.MustParse("7d8f3b2a-1c4e-4a9b-8f6d-2e5c9a1b3d4f")
	next := "next-cursor-token"

	t.Run("returns items and next cursor", func(t *testing.T) {
		svc := fakeCrawlService{listFn: func(_ context.Context, token string, limit int) (crawl.Page, error) {
			if token != "" {
				t.Errorf("token = %q, want empty for first page", token)
			}
			return crawl.Page{
				Items: []crawl.Crawl{{
					ID:        id,
					SeedURL:   "https://example.com",
					Status:    "queued",
					CreatedAt: time.Now().UTC(),
				}},
				NextCursor: &next,
			}, nil
		}}

		router := newTestRouter(svc)
		req := httptest.NewRequest(http.MethodGet, "/crawls", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
		}
		var got types.CrawlList
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode CrawlList: %v", err)
		}
		if len(got.Items) != 1 {
			t.Fatalf("items = %d, want 1", len(got.Items))
		}
		if got.NextCursor == nil || *got.NextCursor != next {
			t.Errorf("nextCursor = %v, want %q", got.NextCursor, next)
		}
	})

	t.Run("forwards cursor and limit", func(t *testing.T) {
		svc := fakeCrawlService{listFn: func(_ context.Context, token string, limit int) (crawl.Page, error) {
			if token != "abc" {
				t.Errorf("token = %q, want abc", token)
			}
			if limit != 50 {
				t.Errorf("limit = %d, want 50", limit)
			}
			return crawl.Page{Items: nil}, nil
		}}

		router := newTestRouter(svc)
		req := httptest.NewRequest(http.MethodGet, "/crawls?cursor=abc&limit=50", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		var got types.CrawlList
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode CrawlList: %v", err)
		}
		// Empty page must serialise items as [] (not null) to match the contract.
		if got.Items == nil {
			t.Error("items is null, want empty array")
		}
	})

	t.Run("invalid limit returns 400", func(t *testing.T) {
		router := newTestRouter(fakeCrawlService{})
		req := httptest.NewRequest(http.MethodGet, "/crawls?limit=abc", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("over-max limit returns 400", func(t *testing.T) {
		router := newTestRouter(fakeCrawlService{})
		req := httptest.NewRequest(http.MethodGet, "/crawls?limit=500", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("invalid cursor returns 400", func(t *testing.T) {
		svc := fakeCrawlService{listFn: func(_ context.Context, _ string, _ int) (crawl.Page, error) {
			return crawl.Page{}, crawl.ErrInvalidCursor
		}}
		router := newTestRouter(svc)
		req := httptest.NewRequest(http.MethodGet, "/crawls?cursor=@@bad@@", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})
}
