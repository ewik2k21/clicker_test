package repository

import (
	"context"
	"database/sql"
	"github.com/ewik2k21/clicker_test/internal/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Repository interface {
	SaveClicks(ctx context.Context, clicks map[uuid.UUID]int) error
	GetStats(ctx context.Context, bannerID uuid.UUID, tsFrom, tsTo time.Time) ([]model.ClickStat, error)
	BannerExists(ctx context.Context, bannerID uuid.UUID) (bool, error)
	GetRandomBanners(ctx context.Context, limit int) ([]model.Banner, error)
}

type repository struct {
	db     *sql.DB
	logger *zap.Logger
	mu     sync.RWMutex
}

func NewRepository(db *sql.DB, logger *zap.Logger) Repository {
	return &repository{db: db, logger: logger}
}

func (r *repository) SaveClicks(ctx context.Context, clicks map[uuid.UUID]int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction", zap.Error(err))
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO clicks (timestamp, banner_id, count)
		VALUES ($1, $2, $3)
		ON CONFLICT (timestamp, banner_id)
		DO UPDATE SET count = clicks.count + EXCLUDED.count
	`)
	if err != nil {
		tx.Rollback()
		r.logger.Error("Failed to prepare statement", zap.Error(err))
		return err
	}
	defer stmt.Close()

	now := time.Now().Truncate(time.Minute)
	r.mu.RLock()
	for bannerID, count := range clicks {
		if _, err = stmt.ExecContext(ctx, now, bannerID, count); err != nil {
			tx.Rollback()
			r.logger.Error("Failed to execute statement", zap.Error(err))
			return err
		}
	}
	r.mu.RUnlock()

	if err = tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", zap.Error(err))
		return err
	}

	return nil
}

func (r *repository) GetStats(ctx context.Context, bannerID uuid.UUID, tsFrom, tsTo time.Time) ([]model.ClickStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT timestamp, banner_id, count
		FROM clicks
		WHERE banner_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp
	`, bannerID, tsFrom, tsTo)
	if err != nil {
		r.logger.Error("Failed to query stats", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var stats []model.ClickStat
	for rows.Next() {
		var stat model.ClickStat
		if err = rows.Scan(&stat.Timestamp, &stat.BannerID, &stat.Count); err != nil {
			r.logger.Error("Failed to scan row", zap.Error(err))
			return nil, err
		}
		stats = append(stats, stat)
	}
	if err = rows.Err(); err != nil {
		r.logger.Error("Rows error", zap.Error(err))
		return nil, err
	}
	return stats, nil

}

func (r *repository) BannerExists(ctx context.Context, bannerID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM banners WHERE id = $1)", bannerID).Scan(&exists)
	if err != nil {
		r.logger.Error("Failed to check banner existence", zap.Error(err))
		return false, err
	}

	return exists, nil
}

func (r *repository) GetRandomBanners(ctx context.Context, limit int) ([]model.Banner, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name
		FROM banners
		ORDER BY RANDOM()
		LIMIT $1
	`, limit)
	if err != nil {
		r.logger.Error("Failed to query random banners", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var banners []model.Banner
	for rows.Next() {
		var banner model.Banner
		if err := rows.Scan(&banner.ID, &banner.Name); err != nil {
			r.logger.Error("Failed to scan banner row", zap.Error(err))
			return nil, err
		}
		banners = append(banners, banner)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("Rows error", zap.Error(err))
		return nil, err
	}

	return banners, nil
}
