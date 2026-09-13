// Command genmodule scaffolds a new domain module (vertical slice).
//
// Usage: go run ./tools/genmodule -name order
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

func main() {
	name := flag.String("name", "", "module name: lowercase letters and digits (e.g. order)")
	dir := flag.String("dir", "internal/modules", "base directory for modules")
	force := flag.Bool("force", false, "overwrite existing files")
	flag.Parse()

	if !namePattern.MatchString(*name) {
		fmt.Fprintln(os.Stderr, "invalid -name: use lowercase letters and digits, starting with a letter")
		os.Exit(1)
	}
	camel := capitalize(*name)

	target := filepath.Join(*dir, *name)
	if err := os.MkdirAll(target, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "create directory:", err)
		os.Exit(1)
	}

	files := []struct {
		name    string
		content string
	}{
		{"dto.go", fmt.Sprintf(dtoTemplate, *name, camel)},
		{"service.go", fmt.Sprintf(serviceTemplate, *name, camel)},
		{"repository.go", fmt.Sprintf(repositoryTemplate, *name, camel)},
		{"handler.go", fmt.Sprintf(handlerTemplate, *name, camel)},
		{"router.go", fmt.Sprintf(routerTemplate, *name, camel)},
	}

	for _, file := range files {
		path := filepath.Join(target, file.name)
		if !*force {
			if _, err := os.Stat(path); err == nil {
				fmt.Fprintln(os.Stderr, "file exists (use -force):", path)
				continue
			}
		}
		if err := os.WriteFile(path, []byte(file.content), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "write file:", err)
			os.Exit(1)
		}
		fmt.Println("created", path)
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

const dtoTemplate = `package %[1]s

// Create%[2]sRequest describes the request body.
type Create%[2]sRequest struct {
	// TODO: add fields with json and validate tags.
}

// %[2]sResponse describes the response body.
type %[2]sResponse struct {
	// TODO: add fields with json tags.
}
`

const serviceTemplate = `package %[1]s

import (
	"context"

	"template-golang/pkg/apperror"
)

// Repository is the persistence contract of the %[1]s module.
type Repository interface {
	Ping(ctx context.Context) error
}

// Service holds the business logic of the %[1]s module.
type Service struct {
	repo Repository
}

// NewService builds the %[1]s service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new entity.
func (s *Service) Create(ctx context.Context) (%[2]sResponse, error) {
	if err := s.repo.Ping(ctx); err != nil {
		return %[2]sResponse{}, apperror.Runtime(apperror.CodeDBQueryFailed, err)
	}
	return %[2]sResponse{}, nil
}
`

const repositoryTemplate = `package %[1]s

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements Repository with pgx.
// Use db tags and pgx.RowToStructByName for every query.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository builds the repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// Ping is a placeholder to keep the scaffold compiling.
func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
`

const handlerTemplate = `package %[1]s

import (
	"context"
)

// Handler exposes the %[1]s endpoints.
type Handler struct {
	service *Service
}

// NewHandler builds the %[1]s handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create handles the create endpoint.
func (h *Handler) Create(ctx context.Context, req Create%[2]sRequest) (%[2]sResponse, error) {
	_ = req
	return h.service.Create(ctx)
}
`

const routerTemplate = `package %[1]s

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"template-golang/pkg/httpx"
)

// Deps lists the infrastructure the %[1]s module needs.
type Deps struct {
	Pool *pgxpool.Pool
	Log  *slog.Logger
}

// Register is the composition root of the module: it builds the repository,
// service and handler and mounts the routes. Call it from
// internal/app/routes.go (public or protected router).
func Register(router fiber.Router, registry *httpx.Registry, deps Deps) {
	repo := NewPostgresRepository(deps.Pool)
	service := NewService(repo)
	handler := NewHandler(service)

	httpx.Post[Create%[2]sRequest, %[2]sResponse](
		registry, router, "/%[1]s", handler.Create, httpx.Protected(), httpx.Tag("%[1]s"),
	)
}
`
