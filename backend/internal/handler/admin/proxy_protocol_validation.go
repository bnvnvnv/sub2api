package admin

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/proxyprotocol"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

func normalizeAdminProxyProtocol(protocol string) string {
	return proxyprotocol.NormalizeScheme(protocol)
}

func isAdminProxyProtocolSupported(protocol string) bool {
	return proxyprotocol.IsSupported(protocol)
}

func validateAdminProxyProtocol(c *gin.Context, protocol string, allowEmpty bool) (string, bool) {
	normalized := normalizeAdminProxyProtocol(protocol)
	if normalized == "" && allowEmpty {
		return "", true
	}
	if isAdminProxyProtocolSupported(normalized) {
		return normalized, true
	}
	response.BadRequest(c, "Invalid proxy protocol: "+strings.TrimSpace(protocol))
	return "", false
}
