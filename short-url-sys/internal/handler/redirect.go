package handler

import (
	"log"
	"net/http"
	"short-url-sys/internal/model"
	"short-url-sys/internal/service/redirect"
	"strings"

	"github.com/gin-gonic/gin"
)

type RedirectHandler struct {
	redirectService redirect.Service
}

func NewRedirectHandler(redirectService redirect.Service) *RedirectHandler {
	return &RedirectHandler{
		redirectService: redirectService,
	}
}

// Redirect
// @Router /{code} [get]
func (h *RedirectHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid short code",
			Message: "Short code is required",
		})
		return
	}

	// 构建重定向请求
	ip := getClientIP(c)
	req := redirect.RedirectRequest{
		IPAddress: ip,
		UserAgent: c.Request.UserAgent(),
		Referer:   c.Request.Referer(),
		// TODO 简化实现，实际应该使用IP地理定位服务
		Country: getCountryFromIP(ip),
		Region:  "",
		City:    "",
	}
	longURL, err := h.redirectService.Redirect(c.Request.Context(), shortCode, &req)
	if err != nil {
		c.Error(err)
		return
	}
	// 使用 302 临时重定向，便于统计
	log.Printf("Redirect URL: %s", longURL)
	c.Redirect(http.StatusFound, longURL)
}

// 获取客户端IP
func getClientIP(c *gin.Context) string {
	// 尝试从 X-Forwarded-For 获取
	// X-Forwarded-For：包含客户端和所有代理服务器的 IP 链
	if forwarded := c.Request.Header.Get("X-Forwarded-For"); forwarded != "" {
		if ips := strings.Split(forwarded, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从 X-Real-IP 获取
	// X-Real-IP：通常由第一个代理设置，直接包含客户端真实 IP
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	// 使用远程地址
	return c.ClientIP()
}

// 根据IP获取国家
func getCountryFromIP(ip string) string {
	// TODO 根据IP获取国家代码
	// ISO 3166-1 alpha-2
	return "CN"
}
