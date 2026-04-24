package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type OpenAIWebThread struct {
	ID                     int64      `json:"id"`
	UserID                 int64      `json:"user_id"`
	GroupID                int64      `json:"group_id"`
	AccountID              int64      `json:"account_id"`
	LocalThreadID          string     `json:"local_thread_id"`
	PageSessionID          string     `json:"page_session_id"`
	UpstreamConversationID *string    `json:"upstream_conversation_id,omitempty"`
	UpstreamSessionID      *string    `json:"upstream_session_id,omitempty"`
	Provider               string     `json:"provider"`
	Title                  string     `json:"title"`
	RequestedModel         string     `json:"requested_model"`
	CapabilityMode         string     `json:"capability_mode"`
	HistoryMode            string     `json:"history_mode"`
	CachePolicy            string     `json:"cache_policy"`
	SyncStatus             string     `json:"sync_status"`
	Status                 string     `json:"status"`
	LastSyncedAt           *time.Time `json:"last_synced_at,omitempty"`
	LastError              string     `json:"last_error,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`

	Group   *OpenAIWebGroupSummary   `json:"group,omitempty"`
	Account *OpenAIWebAccountSummary `json:"account,omitempty"`
}

type OpenAIWebEntitlement struct {
	GroupID         int64      `json:"group_id"`
	GroupName       string     `json:"group_name"`
	GroupDesc       string     `json:"group_desc"`
	AccessMode      string     `json:"access_mode"`
	SubscriptionID  *int64     `json:"subscription_id,omitempty"`
	HasWebChat      bool       `json:"has_web_chat"`
	HasProAccounts  bool       `json:"has_pro_accounts"`
	DefaultModel    string     `json:"default_model"`
	CapabilityMode  string     `json:"capability_mode"`
	SubscriptionEnd *time.Time `json:"subscription_end,omitempty"`
}

type OpenAIWebThreadMessageUsage struct {
	InputTokens         int `json:"input_tokens"`
	OutputTokens        int `json:"output_tokens"`
	CacheReadTokens     int `json:"cache_read_tokens"`
	CacheCreationTokens int `json:"cache_creation_tokens"`
	ImageOutputTokens   int `json:"image_output_tokens"`
	TotalTokens         int `json:"total_tokens"`
}

type OpenAIWebThreadMessageImage struct {
	DataURL       string `json:"data_url,omitempty"`
	MimeType      string `json:"mime_type,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
	Width         int    `json:"width,omitempty"`
	Height        int    `json:"height,omitempty"`
}

type OpenAIWebThreadMessageResponse struct {
	Thread          *OpenAIWebThread              `json:"thread"`
	AssistantText   string                        `json:"assistant_text"`
	AssistantImages []OpenAIWebThreadMessageImage `json:"assistant_images,omitempty"`
	RequestID       string                        `json:"request_id,omitempty"`
	ResponseID      *string                       `json:"response_id,omitempty"`
	Model           string                        `json:"model"`
	Usage           OpenAIWebThreadMessageUsage   `json:"usage"`
}

type OpenAIWebGroupSummary struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Platform    string `json:"platform"`
}

type OpenAIWebAccountSummary struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Platform string `json:"platform"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	PlanType string `json:"plan_type,omitempty"`
}

func OpenAIWebThreadFromService(thread *service.OpenAIWebThread) *OpenAIWebThread {
	if thread == nil {
		return nil
	}

	out := &OpenAIWebThread{
		ID:                     thread.ID,
		UserID:                 thread.UserID,
		GroupID:                thread.GroupID,
		AccountID:              thread.AccountID,
		LocalThreadID:          thread.LocalThreadID,
		PageSessionID:          thread.PageSessionID,
		UpstreamConversationID: thread.UpstreamConversationID,
		UpstreamSessionID:      thread.UpstreamSessionID,
		Provider:               thread.Provider,
		Title:                  thread.Title,
		RequestedModel:         thread.RequestedModel,
		CapabilityMode:         thread.CapabilityMode,
		HistoryMode:            thread.HistoryMode,
		CachePolicy:            thread.CachePolicy,
		SyncStatus:             thread.SyncStatus,
		Status:                 thread.Status,
		LastSyncedAt:           thread.LastSyncedAt,
		LastError:              thread.LastError,
		CreatedAt:              thread.CreatedAt,
		UpdatedAt:              thread.UpdatedAt,
	}

	if thread.Group != nil {
		out.Group = OpenAIWebGroupSummaryFromService(thread.Group)
	}
	if thread.Account != nil {
		out.Account = OpenAIWebAccountSummaryFromService(thread.Account)
	}

	return out
}

func OpenAIWebThreadsFromService(threads []service.OpenAIWebThread) []OpenAIWebThread {
	out := make([]OpenAIWebThread, 0, len(threads))
	for i := range threads {
		mapped := OpenAIWebThreadFromService(&threads[i])
		if mapped != nil {
			out = append(out, *mapped)
		}
	}
	return out
}

func OpenAIWebGroupSummaryFromService(group *service.Group) *OpenAIWebGroupSummary {
	if group == nil {
		return nil
	}
	return &OpenAIWebGroupSummary{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Platform:    group.Platform,
	}
}

func OpenAIWebAccountSummaryFromService(account *service.Account) *OpenAIWebAccountSummary {
	if account == nil {
		return nil
	}
	return &OpenAIWebAccountSummary{
		ID:       account.ID,
		Name:     account.Name,
		Platform: account.Platform,
		Type:     account.Type,
		Status:   account.Status,
		PlanType: account.GetOpenAIPlanType(),
	}
}

func OpenAIWebEntitlementsFromService(entitlements []service.OpenAIWebThreadEntitlement) []OpenAIWebEntitlement {
	out := make([]OpenAIWebEntitlement, 0, len(entitlements))
	for i := range entitlements {
		item := entitlements[i]
		out = append(out, OpenAIWebEntitlement{
			GroupID:         item.GroupID,
			GroupName:       item.GroupName,
			GroupDesc:       item.GroupDesc,
			AccessMode:      item.AccessMode,
			SubscriptionID:  item.SubscriptionID,
			HasWebChat:      item.HasWebChat,
			HasProAccounts:  item.HasProAccounts,
			DefaultModel:    item.DefaultModel,
			CapabilityMode:  item.CapabilityMode,
			SubscriptionEnd: item.SubscriptionEnd,
		})
	}
	return out
}

func OpenAIWebThreadMessageResponseFromService(result *service.OpenAIWebThreadMessageResult) *OpenAIWebThreadMessageResponse {
	if result == nil {
		return nil
	}
	return &OpenAIWebThreadMessageResponse{
		Thread:          OpenAIWebThreadFromService(result.Thread),
		AssistantText:   result.AssistantText,
		AssistantImages: openAIWebThreadMessageImagesFromService(result.AssistantImages),
		RequestID:       result.RequestID,
		ResponseID:      result.ResponseID,
		Model:           result.Model,
		Usage: OpenAIWebThreadMessageUsage{
			InputTokens:         result.Usage.InputTokens,
			OutputTokens:        result.Usage.OutputTokens,
			CacheReadTokens:     result.Usage.CacheReadTokens,
			CacheCreationTokens: result.Usage.CacheCreationTokens,
			ImageOutputTokens:   result.Usage.ImageOutputTokens,
			TotalTokens:         result.Usage.TotalTokens,
		},
	}
}

func openAIWebThreadMessageImagesFromService(images []service.OpenAIWebThreadMessageImage) []OpenAIWebThreadMessageImage {
	if len(images) == 0 {
		return nil
	}
	out := make([]OpenAIWebThreadMessageImage, 0, len(images))
	for i := range images {
		item := images[i]
		out = append(out, OpenAIWebThreadMessageImage{
			DataURL:       item.DataURL,
			MimeType:      item.MimeType,
			RevisedPrompt: item.RevisedPrompt,
			Width:         item.Width,
			Height:        item.Height,
		})
	}
	return out
}
