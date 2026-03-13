package build

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides persistence for builds and build versions.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new Repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// List returns all builds ordered by ID.
func (r *Repository) List(ctx context.Context) ([]Build, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, name, class, description, created_at
		FROM builds ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("build: list: %w", err)
	}
	defer rows.Close()
	builds := []Build{}
	for rows.Next() {
		var b Build
		if err := rows.Scan(&b.ID, &b.Slug, &b.Name, &b.Class, &b.Description, &b.CreatedAt); err != nil {
			return nil, err
		}
		builds = append(builds, b)
	}
	return builds, rows.Err()
}

// GetByID returns a single build by primary key.
func (r *Repository) GetByID(ctx context.Context, id int) (*Build, error) {
	b := &Build{}
	err := r.db.QueryRow(ctx, `
		SELECT id, slug, name, class, description, created_at
		FROM builds WHERE id = $1`, id,
	).Scan(&b.ID, &b.Slug, &b.Name, &b.Class, &b.Description, &b.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("build: %d not found", id)
		}
		return nil, fmt.Errorf("build: get %d: %w", id, err)
	}
	return b, nil
}

// Create inserts a new build and fills in the generated ID and timestamp.
func (r *Repository) Create(ctx context.Context, b *Build) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO builds (slug, name, class, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`,
		b.Slug, b.Name, b.Class, b.Description,
	).Scan(&b.ID, &b.CreatedAt)
}

// ListVersions returns all versions for the given build ordered by ID.
func (r *Repository) ListVersions(ctx context.Context, buildID int) ([]Version, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, build_id, version, patch_tag, notes, is_current, released_at, created_at
		FROM build_versions WHERE build_id = $1 ORDER BY id`, buildID)
	if err != nil {
		return nil, fmt.Errorf("build: list versions: %w", err)
	}
	defer rows.Close()
	versions := []Version{}
	for rows.Next() {
		var v Version
		if err := rows.Scan(&v.ID, &v.BuildID, &v.Version, &v.PatchTag,
			&v.Notes, &v.IsCurrent, &v.ReleasedAt, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

// CreateVersion inserts a new build version and fills in the generated ID and timestamp.
func (r *Repository) CreateVersion(ctx context.Context, v *Version) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO build_versions (build_id, version, patch_tag, notes, is_current, released_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`,
		v.BuildID, v.Version, v.PatchTag, v.Notes, v.IsCurrent, v.ReleasedAt,
	).Scan(&v.ID, &v.CreatedAt)
}
