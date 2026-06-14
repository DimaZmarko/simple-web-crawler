package crawl_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/DimaZmarko/simple-web-crawler/apps/api/internal/crawl"
	"github.com/DimaZmarko/simple-web-crawler/apps/api/internal/db"
)

// TestMain makes the testcontainers-backed integration tests run under
// `make test-api` (plain `go test -race ./...`) without per-developer env vars.
//
// On a non-default Docker provider such as Colima, the daemon socket is not at
// /var/run/docker.sock, so testcontainers cannot find it. When DOCKER_HOST is
// unset we discover the active `docker context` endpoint and export it. We also
// disable the Ryuk reaper, which is unreliable on Colima. CI that already sets
// DOCKER_HOST or uses the default socket is left untouched.
func TestMain(m *testing.M) {
	if os.Getenv("DOCKER_HOST") == "" {
		if endpoint := activeDockerEndpoint(); endpoint != "" {
			_ = os.Setenv("DOCKER_HOST", endpoint)
			if _, ok := os.LookupEnv("TESTCONTAINERS_RYUK_DISABLED"); !ok {
				_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
			}
		}
	}
	os.Exit(m.Run())
}

// activeDockerEndpoint returns the daemon endpoint of the active docker context,
// or "" when the docker CLI is unavailable or reports the default socket.
func activeDockerEndpoint() string {
	out, err := exec.Command("docker", "context", "inspect", "--format", "{{.Endpoints.docker.Host}}").Output()
	if err != nil {
		return ""
	}
	endpoint := strings.TrimSpace(string(out))
	if endpoint == "" || endpoint == "unix:///var/run/docker.sock" {
		return ""
	}
	return endpoint
}

// startPostgres boots a throwaway Postgres container, applies every up
// migration, and returns a connected pool. The container and pool are cleaned
// up via t.Cleanup.
func startPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("crawler"),
		tcpostgres.WithUsername("crawler"),
		tcpostgres.WithPassword("crawler"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("terminate container: %v", err)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect pool: %v", err)
	}
	t.Cleanup(pool.Close)

	// Colima maps the container port to the host loopback asynchronously, so the
	// forward can lag behind the "ready" log. Retry Ping until it lands.
	waitForPool(t, ctx, pool)

	applyMigrations(t, ctx, pool)
	return pool
}

// waitForPool pings the pool until it succeeds or a deadline elapses.
func waitForPool(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	deadline := time.Now().Add(30 * time.Second)
	for {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err := pool.Ping(pingCtx)
		cancel()
		if err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("database never became reachable: %v", err)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

// applyMigrations execs every *.up.sql under db/migrations in lexical order.
func applyMigrations(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	// Tests run from the package dir; migrations live two levels up.
	dir := filepath.Join("..", "..", "db", "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	var ups []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			ups = append(ups, e.Name())
		}
	}
	sort.Strings(ups)

	for _, name := range ups {
		sqlBytes, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			t.Fatalf("apply migration %s: %v", name, err)
		}
	}
}

func TestServiceCreateGetIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}

	pool := startPostgres(t)
	svc := crawl.NewService(db.New(pool))
	ctx := context.Background()

	created, err := svc.Create(ctx, crawl.CreateInput{
		SeedURL:  "https://example.com",
		MaxDepth: 3,
		MaxPages: 250,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Status != "queued" {
		t.Errorf("status = %q, want queued", created.Status)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Error("timestamps not populated")
	}

	got, err := svc.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("id = %v, want %v", got.ID, created.ID)
	}
	if got.SeedURL != "https://example.com" || got.MaxDepth != 3 || got.MaxPages != 250 {
		t.Errorf("round-trip mismatch: %+v", got)
	}
}

func TestServiceListPaginationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}

	pool := startPostgres(t)
	svc := crawl.NewService(db.New(pool))
	ctx := context.Background()

	const total = 25
	for i := 0; i < total; i++ {
		if _, err := svc.Create(ctx, crawl.CreateInput{
			SeedURL:  "https://example.com",
			MaxDepth: 1,
			MaxPages: 10,
		}); err != nil {
			t.Fatalf("seed create %d: %v", i, err)
		}
	}

	const pageSize = 10
	seen := map[string]bool{}
	token := ""
	pages := 0

	for {
		page, err := svc.List(ctx, token, pageSize)
		if err != nil {
			t.Fatalf("list page %d: %v", pages, err)
		}
		pages++

		for _, c := range page.Items {
			if seen[c.ID.String()] {
				t.Errorf("duplicate id across pages: %s", c.ID)
			}
			seen[c.ID.String()] = true
		}

		if page.NextCursor == nil {
			if len(page.Items) > pageSize {
				t.Errorf("final page has %d items, want <= %d", len(page.Items), pageSize)
			}
			break
		}

		if len(page.Items) != pageSize {
			t.Errorf("page %d has %d items, want %d", pages, len(page.Items), pageSize)
		}
		token = *page.NextCursor

		if pages > 10 {
			t.Fatal("pagination did not terminate")
		}
	}

	if len(seen) != total {
		t.Errorf("saw %d unique crawls, want %d", len(seen), total)
	}
	// 25 rows at 10/page = 3 pages.
	if pages != 3 {
		t.Errorf("paged in %d pages, want 3", pages)
	}
}

func TestServiceGetNotFoundIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}

	pool := startPostgres(t)
	svc := crawl.NewService(db.New(pool))

	// A valid but absent UUID must surface ErrNotFound.
	_, err := svc.Get(context.Background(), uuid.New())
	if !errors.Is(err, crawl.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
