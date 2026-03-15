package game

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAreaRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAreaRepository(pool *pgxpool.Pool) *PostgresAreaRepository {
	return &PostgresAreaRepository{pool: pool}
}

func (r *PostgresAreaRepository) GetByCode(ctx context.Context, areaCode string) (*Area, error) {
	query := `SELECT area_code, name, act FROM game_areas WHERE area_code = $1`
	var area Area
	err := r.pool.QueryRow(ctx, query, areaCode).Scan(&area.AreaCode, &area.Name, &area.Act)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("query area by code: %w", err)
	}
	return &area, nil
}

func (r *PostgresAreaRepository) GetAll(ctx context.Context) ([]Area, error) {
	query := `SELECT area_code, name, act FROM game_areas`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query all areas: %w", err)
	}
	defer rows.Close()

	var areas []Area
	for rows.Next() {
		var area Area
		if err := rows.Scan(&area.AreaCode, &area.Name, &area.Act); err != nil {
			return nil, fmt.Errorf("scan area: %w", err)
		}
		areas = append(areas, area)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return areas, nil
}

func (r *PostgresAreaRepository) GetByName(ctx context.Context, name string) ([]Area, error) {
        query := `SELECT area_code, name, act FROM game_areas WHERE name = $1`
        rows, err := r.pool.Query(ctx, query, name)
        if err != nil {
                return nil, fmt.Errorf("query area by name: %w", err)
        }
        defer rows.Close()

        var areas []Area
        for rows.Next() {
                var area Area
                if err := rows.Scan(&area.AreaCode, &area.Name, &area.Act); err != nil {
                        return nil, fmt.Errorf("scan area: %w", err)
                }
                areas = append(areas, area)
        }
        if err := rows.Err(); err != nil {
                return nil, fmt.Errorf("rows error: %w", err)
        }
        return areas, nil
}
