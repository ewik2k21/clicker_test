package service

import (
	"context"
	"errors"
	"github.com/ewik2k21/clicker_test/internal/model"
	"github.com/ewik2k21/clicker_test/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Service interface {
	RegisterClick(ctx *gin.Context, bannerID uuid.UUID) error
	GetStats(ctx *gin.Context, bannerID uuid.UUID, tsFrom, tsTo time.Time) ([]model.ClickStat, error)
	ProcessClicks(ctx context.Context)
	GetRandomBanners(ctx *gin.Context, limit int) ([]model.Banner, error)
}

type service struct {
	repo     repository.Repository
	logger   *zap.Logger
	clicks   chan uuid.UUID
	clickMap map[uuid.UUID]int
	mu       sync.Mutex
}

func NewService(repo repository.Repository, logger *zap.Logger) Service {
	return &service{
		repo:     repo,
		logger:   logger,
		clicks:   make(chan uuid.UUID, 1000),
		clickMap: make(map[uuid.UUID]int),
	}
}

func (s *service) RegisterClick(ctx *gin.Context, bannerID uuid.UUID) error {
	exists, err := s.repo.BannerExists(ctx, bannerID)
	if err != nil {
		s.logger.Error("Failed banner exists method", zap.Error(err))
		return err
	}
	if !exists {
		s.logger.Error("Banner not found")
		return errors.New("banner not found")
	}

	select {
	case s.clicks <- bannerID:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *service) ProcessClicks(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.flushClicks(ctx)
			return
		case bannerID := <-s.clicks:
			s.mu.Lock()
			s.clickMap[bannerID]++
			s.mu.Unlock()
		case <-ticker.C:
			s.flushClicks(ctx)
		}
	}
}

func (s *service) flushClicks(ctx context.Context) {
	s.mu.Lock()
	if len(s.clickMap) == 0 {
		s.mu.Unlock()
		return
	}

	clicks := s.clickMap
	s.clickMap = make(map[uuid.UUID]int)
	s.mu.Unlock()

	if err := s.repo.SaveClicks(ctx, clicks); err != nil {
		s.logger.Error("Failed to save clicks", zap.Error(err))
	}
}

func (s *service) GetStats(ctx *gin.Context, bannerID uuid.UUID, tsFrom, tsTo time.Time) ([]model.ClickStat, error) {
	return s.repo.GetStats(ctx, bannerID, tsFrom, tsTo)
}

func (s *service) GetRandomBanners(ctx *gin.Context, limit int) ([]model.Banner, error) {
	return s.repo.GetRandomBanners(ctx, limit)
}
