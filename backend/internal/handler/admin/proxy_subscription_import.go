package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type ImportSubscriptionRequest struct {
	Proxies []ParsedSubscriptionProxy `json:"proxies" binding:"required,min=1"`
}

type ImportSubscriptionResponse struct {
	Created       int                             `json:"created"`
	Skipped       int                             `json:"skipped"`
	DirectCreated int                             `json:"direct_created"`
	Warnings      []string                        `json:"warnings"`
	Failed        []ImportSubscriptionFailureItem `json:"failed"`
}

type ImportSubscriptionFailureItem struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func (h *ProxyHandler) ImportSubscription(c *gin.Context) {
	if h.adminService == nil {
		response.BadRequest(c, "proxy subscription import service is not configured")
		return
	}

	var req ImportSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	result := h.importSubscriptionProxies(c.Request.Context(), req.Proxies)
	response.Success(c, result)
}

func (h *ProxyHandler) importSubscriptionProxies(ctx context.Context, proxies []ParsedSubscriptionProxy) ImportSubscriptionResponse {
	result := ImportSubscriptionResponse{
		Failed:   make([]ImportSubscriptionFailureItem, 0),
		Warnings: make([]string, 0),
	}

	for _, item := range proxies {
		created, skipped, err := h.importDirectSubscriptionProxy(ctx, item)
		if err != nil {
			result.Failed = append(result.Failed, ImportSubscriptionFailureItem{Name: item.Name, Message: err.Error()})
			continue
		}
		if skipped {
			result.Skipped++
			continue
		}
		if created {
			result.Created++
			result.DirectCreated++
		}
	}

	return result
}

func (h *ProxyHandler) importDirectSubscriptionProxy(ctx context.Context, item ParsedSubscriptionProxy) (created bool, skipped bool, err error) {
	protocol := normalizeDetectedProtocol(item.Protocol)
	if !isSupportedImportProtocol(protocol) {
		return false, false, fmt.Errorf("unsupported direct proxy protocol %q", item.Protocol)
	}
	host := strings.TrimSpace(item.Host)
	if host == "" || item.Port <= 0 || item.Port > 65535 {
		return false, false, errors.New("invalid direct proxy host or port")
	}
	username := strings.TrimSpace(item.Username)
	password := strings.TrimSpace(item.Password)

	exists, err := h.adminService.CheckProxyExists(ctx, host, item.Port, username, password)
	if err != nil {
		return false, false, err
	}
	if exists {
		return false, true, nil
	}

	_, err = h.adminService.CreateProxy(ctx, &service.CreateProxyInput{
		Name:     defaultProxyName(strings.TrimSpace(item.Name)),
		Protocol: protocol,
		Host:     host,
		Port:     item.Port,
		Username: username,
		Password: password,
	})
	if err != nil {
		return false, false, err
	}
	return true, false, nil
}
