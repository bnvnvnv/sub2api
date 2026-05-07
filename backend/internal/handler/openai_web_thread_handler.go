package handler

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type CreateOpenAIWebThreadRequest struct {
	GroupID        int64  `json:"group_id" binding:"required"`
	RequestedModel string `json:"requested_model"`
	Title          string `json:"title"`
	CachePolicy    string `json:"cache_policy"`
}

type SendOpenAIWebThreadMessageRequest struct {
	RequestedModel  string                                 `json:"requested_model,omitempty"`
	ReasoningEffort string                                 `json:"reasoning_effort,omitempty"`
	Content         string                                 `json:"content"`
	Attachments     []SendOpenAIWebThreadAttachmentRequest `json:"attachments,omitempty"`
}

type SendOpenAIWebThreadAttachmentRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	DataURL     string `json:"data_url"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

type OpenAIWebThreadHandler struct {
	threadService *service.OpenAIWebThreadService
}

func NewOpenAIWebThreadHandler(threadService *service.OpenAIWebThreadService) *OpenAIWebThreadHandler {
	return &OpenAIWebThreadHandler{threadService: threadService}
}

func (h *OpenAIWebThreadHandler) ListEntitlements(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	entitlements, err := h.threadService.ListUserEntitlements(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.OpenAIWebEntitlementsFromService(entitlements))
}

func (h *OpenAIWebThreadHandler) ListThreads(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	threads, err := h.threadService.ListUserThreads(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.OpenAIWebThreadsFromService(threads))
}

func (h *OpenAIWebThreadHandler) CreateThread(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req CreateOpenAIWebThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	thread, err := h.threadService.CreateUserThread(c.Request.Context(), &service.CreateOpenAIWebThreadInput{
		UserID:         subject.UserID,
		GroupID:        req.GroupID,
		RequestedModel: strings.TrimSpace(req.RequestedModel),
		Title:          req.Title,
		CachePolicy:    req.CachePolicy,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Created(c, dto.OpenAIWebThreadFromService(thread))
}

func (h *OpenAIWebThreadHandler) GetThread(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	thread, err := h.threadService.GetUserThread(c.Request.Context(), subject.UserID, c.Param("id"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.OpenAIWebThreadFromService(thread))
}

func (h *OpenAIWebThreadHandler) ArchiveThread(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	if err := h.threadService.ArchiveUserThread(c.Request.Context(), subject.UserID, c.Param("id")); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"archived": true})
}

func (h *OpenAIWebThreadHandler) SendThreadMessage(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req SendOpenAIWebThreadMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	result, err := h.threadService.SendUserThreadMessage(c.Request.Context(), &service.SendOpenAIWebThreadMessageInput{
		UserID:          subject.UserID,
		LocalThreadID:   c.Param("id"),
		RequestedModel:  strings.TrimSpace(req.RequestedModel),
		ReasoningEffort: strings.TrimSpace(req.ReasoningEffort),
		Content:         strings.TrimSpace(req.Content),
		Attachments:     openAIWebThreadAttachmentsFromRequest(req.Attachments),
		UserAgent:       strings.TrimSpace(c.GetHeader("User-Agent")),
		IPAddress:       c.ClientIP(),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.OpenAIWebThreadMessageResponseFromService(result))
}

func openAIWebThreadAttachmentsFromRequest(items []SendOpenAIWebThreadAttachmentRequest) []service.OpenAIWebThreadMessageAttachment {
	if len(items) == 0 {
		return nil
	}
	out := make([]service.OpenAIWebThreadMessageAttachment, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, service.OpenAIWebThreadMessageAttachment{
			FileName:    strings.TrimSpace(item.FileName),
			ContentType: strings.TrimSpace(item.ContentType),
			DataURL:     strings.TrimSpace(item.DataURL),
			Width:       item.Width,
			Height:      item.Height,
		})
	}
	return out
}
