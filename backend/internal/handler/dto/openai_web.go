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
	GroupID         int64     `json:"group_id"`
	GroupName       string    `json:"group_name"`
	GroupDesc       string    `json:"group_desc"`
	SubscriptionID  int64     `json:"subscription_id"`
	HasWebChat      bool      `json:"has_web_chat"`
	HasProAccounts  bool      `json:"has_pro_accounts"`
	DefaultModel    string    `json:"default_model"`
	CapabilityMode  string    `json:"capability_mode"`
	SubscriptionEnd time.Time `json:"subscription_end"`
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
