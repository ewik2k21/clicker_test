package handler

import (
	"github.com/ewik2k21/clicker_test/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// Service интерфейс для бизнес-логики
type Service interface {
	RegisterClick(ctx *gin.Context, bannerID uuid.UUID) error
	GetStats(ctx *gin.Context, bannerID uuid.UUID, tsFrom, tsTo time.Time) ([]model.ClickStat, error)
	GetRandomBanners(ctx *gin.Context, limit int) ([]model.Banner, error)
}

// Handler обрабатывает HTTP-запросы
type Handler struct {
	service Service
	logger  *zap.Logger
}

func NewHandler(svc Service, logger *zap.Logger) *Handler {
	return &Handler{service: svc, logger: logger}
}

func (h *Handler) handleError(c *gin.Context, status int, err error) {
	h.logger.Error(err.Error(), zap.Int("status", status))
	c.JSON(status, gin.H{"error": err.Error()})
}

// PostCountClick godoc
// @Summary Register a click for a banner
// @Description Increments the click counter for the specified banner ID
// @Tags clicks
// @Accept json
// @Produce json
// @Param bannerID path string true "Banner ID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /counter/{bannerID} [post]
func (h *Handler) PostCountClick(c *gin.Context) {
	bannerIDStr := c.Param("bannerID")
	bannerID, err := uuid.Parse(bannerIDStr)
	if err != nil {
		h.logger.Error("Invalid banner ID", zap.Error(err))
		h.handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := h.service.RegisterClick(c, bannerID); err != nil {
		h.logger.Error("Failed to register click", zap.Error(err))
		h.handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// GetStatsForBanner godoc
// @Summary Get click statistics for a banner
// @Description Retrieves click statistics for a banner within a time range
// @Tags stats
// @Accept json
// @Produce json
// @Param bannerID path string true "Banner ID (UUID)"
// @Param tsFrom query string true "Start timestamp (RFC3339)"
// @Param tsTo query string true "End timestamp (RFC3339)"
// @Success 200 {array} model.ClickStat
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /stats/{bannerID} [get]
func (h *Handler) GetStatsForBanner(c *gin.Context) {
	bannerIDStr := c.Param("bannerID")
	bannerID, err := uuid.Parse(bannerIDStr)
	if err != nil {
		h.logger.Error("Invalid banner ID", zap.Error(err))
		h.handleError(c, http.StatusBadRequest, err)
		return
	}

	tsFrom, err := time.Parse(time.RFC3339, c.Query("tsFrom"))
	if err != nil {
		h.logger.Error("Invalid tsFrom", zap.Error(err))
		h.handleError(c, http.StatusBadRequest, err)
		return
	}

	tsTo, err := time.Parse(time.RFC3339, c.Query("tsTo"))
	if err != nil {
		h.logger.Error("Invalid tsTo", zap.Error(err))
		h.handleError(c, http.StatusBadRequest, err)
		return
	}

	stats, err := h.service.GetStats(c, bannerID, tsFrom, tsTo)
	if err != nil {
		h.logger.Error("Failed to get stats", zap.Error(err))
		h.handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetRandomBanners godoc
// @Summary Get 20 random banners
// @Description Retrieves 20 random banners with their IDs and names
// @Tags banners
// @Accept json
// @Produce json
// @Success 200 {array} model.Banner
// @Failure 500 {object} map[string]string "error"
// @Router /banners/random [get]
func (h *Handler) GetRandomBanners(c *gin.Context) {
	banners, err := h.service.GetRandomBanners(c, 20)
	if err != nil {
		h.logger.Error("Failed to get random banners", zap.Error(err))
		h.handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, banners)
}
