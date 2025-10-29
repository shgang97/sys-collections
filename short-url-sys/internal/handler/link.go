package handler

import (
	"fmt"
	"net/http"
	"short-url-sys/internal/model"
	linkSrc "short-url-sys/internal/service/link"

	"github.com/gin-gonic/gin"
)

type LinkHandler struct {
	linkService linkSrc.Service
	baseURL     string
}

func NewLinkHandler(linkService linkSrc.Service, baseURL string) *LinkHandler {
	return &LinkHandler{
		linkService: linkService,
		baseURL:     baseURL,
	}
}

func (h *LinkHandler) CreateShortURL(c *gin.Context) {
	var req model.CreateShortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	link, err := h.linkService.CreateShortURL(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	resp := model.CreateShortResponse{
		ShortURL:  h.buildShortURL(link.ShortCode),
		LongURL:   link.LongURL,
		ShortCode: link.ShortCode,
		CreatedAt: link.CreatedAt,
		ExpiresAt: link.ExpiresAt,
		QRCodeURL: h.buildQRCodeURL(link.ShortCode),
	}
	c.JSON(http.StatusCreated, resp)
}

// 构建完整的短链URL
func (h *LinkHandler) buildShortURL(shortCode string) string {
	return fmt.Sprintf("%s/%s", h.baseURL, shortCode)
}

// 构建二维码URL
func (h *LinkHandler) buildQRCodeURL(shortCode string) string {
	return fmt.Sprintf("%s/api/v1/links/%s/qrcode", h.baseURL, shortCode)
}
