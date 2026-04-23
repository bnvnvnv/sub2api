package service

import (
	"context"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	OpenAIWebThreadProviderOpenAI = "openai_web"

	OpenAIWebThreadStatusActive   = "active"
	OpenAIWebThreadStatusArchived = "archived"
	OpenAIWebThreadStatusBroken   = "broken"

	OpenAIWebCapabilityModeWebChat  = "web_chat"
	OpenAIWebCapabilityModeProModel = "pro_model"

	OpenAIWebHistoryModeUpstreamOnly = "upstream_only"
	OpenAIWebHistoryModeHybrid       = "hybrid"
	OpenAIWebHistoryModeServerMirror = "server_mirror"

	OpenAIWebCachePolicyLocalOnly      = "local_only"
	OpenAIWebCachePolicyLocalEncrypted = "local_encrypted"

	OpenAIWebSyncStatusPending = "pending"
	OpenAIWebSyncStatusReady   = "ready"
	OpenAIWebSyncStatusError   = "error"
)

var (
	ErrOpenAIWebThreadNotFound      = infraerrors.NotFound("OPENAI_WEB_THREAD_NOT_FOUND", "openai web thread not found")
	ErrOpenAIWebThreadAlreadyExists = infraerrors.Conflict("OPENAI_WEB_THREAD_ALREADY_EXISTS", "openai web thread already exists")
	ErrOpenAIWebThreadNilInput      = infraerrors.BadRequest("OPENAI_WEB_THREAD_NIL_INPUT", "openai web thread input cannot be nil")
	ErrOpenAIWebGroupAccessDenied   = infraerrors.Forbidden("OPENAI_WEB_GROUP_ACCESS_DENIED", "no active subscription found for the requested openai group")
	ErrOpenAIWebGroupInvalid        = infraerrors.BadRequest("OPENAI_WEB_GROUP_INVALID", "group is not eligible for openai web chat")
	ErrOpenAIWebNoAvailableAccounts = infraerrors.ServiceUnavailable("OPENAI_WEB_NO_AVAILABLE_ACCOUNTS", "no available openai web accounts")
)

type OpenAIWebThread struct {
	ID        int64
	UserID    int64
	GroupID   int64
	AccountID int64

	LocalThreadID          string
	PageSessionID          string
	UpstreamConversationID *string
	UpstreamSessionID      *string

	Provider       string
	Title          string
	RequestedModel string
	CapabilityMode string
	HistoryMode    string
	CachePolicy    string
	SyncStatus     string
	Status         string

	LastSyncedAt *time.Time
	LastError    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time

	User    *User
	Group   *Group
	Account *Account
}

func (t *OpenAIWebThread) IsActive() bool {
	return t != nil && t.Status == OpenAIWebThreadStatusActive && t.DeletedAt == nil
}

type OpenAIWebThreadEntitlement struct {
	GroupID         int64
	GroupName       string
	GroupDesc       string
	SubscriptionID  int64
	HasWebChat      bool
	HasProAccounts  bool
	DefaultModel    string
	CapabilityMode  string
	SubscriptionEnd time.Time
}

type OpenAIWebThreadRepository interface {
	Create(ctx context.Context, thread *OpenAIWebThread) error
	GetByLocalThreadIDAndUserID(ctx context.Context, localThreadID string, userID int64) (*OpenAIWebThread, error)
	ListActiveByUserID(ctx context.Context, userID int64) ([]OpenAIWebThread, error)
	Update(ctx context.Context, thread *OpenAIWebThread) error
	ArchiveByLocalThreadIDAndUserID(ctx context.Context, localThreadID string, userID int64, archivedAt time.Time) error
}
