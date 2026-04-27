package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/gin-gonic/gin"
)

func registerProxySubscriptionRoutes(proxies *gin.RouterGroup, h *handler.Handlers) {
	proxies.POST("/subscription/parse", h.Admin.ProxySubscription.ParseSubscription)
}
