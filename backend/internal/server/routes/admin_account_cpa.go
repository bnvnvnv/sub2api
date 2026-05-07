package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/gin-gonic/gin"
)

func registerAccountCPARoutes(accounts *gin.RouterGroup, h *handler.Handlers) {
	accounts.POST("/import/cpa/preview", h.Admin.CPAImport.PreviewFromCPA)
	accounts.POST("/import/cpa/remote/preview", h.Admin.CPAImport.PreviewRemoteFromCPA)
	accounts.POST("/import/cpa", h.Admin.CPAImport.ImportFromCPA)
	accounts.POST("/import/cpa/remote", h.Admin.CPAImport.ImportRemoteFromCPA)
}
