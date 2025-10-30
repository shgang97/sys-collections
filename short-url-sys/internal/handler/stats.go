package handler

import (
	"net/http"
	"short-url-sys/internal/model"
	statsSrc "short-url-sys/internal/service/stats"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsService statsSrc.Service
}

func NewStatsHandler(statsService statsSrc.Service) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
	}
}

// GetLinkStats 获取链接统计
// @Router /api/v1/links/stats/{code} [get]
func (h *StatsHandler) GetLinkStats(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid short code",
			Message: "Short code is required",
		})
		return
	}

	var req model.StatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid query",
			Message: "invalid query parameters",
		})
		return
	}

	// 解析日期参数
	var startDate, endDate *time.Time
	if req.StartDate != nil {
		startDate = req.StartDate
	}
	if req.EndDate != nil {
		endDate = req.EndDate
	}
	resp, err := h.statsService.GetLinkStats(c.Request.Context(), shortCode, startDate, endDate)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetDailyStats 获取链接统计信息
// @Router /api/v1/links/stats/daily/{code} [get]
func (h *StatsHandler) GetDailyStats(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid short code",
			Message: "Short code is required",
		})
		return
	}
	days := 30
	if daysParam := c.Query("days"); daysParam != "" {
		if d, err := strconv.Atoi(daysParam); err == nil {
			days = d
		}
	}
	stats, err := h.statsService.GetDailyStats(c.Request.Context(), shortCode, days)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, stats)
}
