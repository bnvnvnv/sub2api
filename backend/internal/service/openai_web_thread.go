package service

import (
	"context"
	"strings"
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

	OpenAIWebInternalAPIKeyNamePrefix = "__sub2api_openai_web_internal__"

	OpenAIWebReasoningEffortLow    = "low"
	OpenAIWebReasoningEffortMedium = "medium"
	OpenAIWebReasoningEffortHigh   = "high"
	OpenAIWebReasoningEffortXHigh  = "xhigh"
)

var (
	ErrOpenAIWebThreadNotFound      = infraerrors.NotFound("OPENAI_WEB_THREAD_NOT_FOUND", "openai web thread not found")
	ErrOpenAIWebThreadAlreadyExists = infraerrors.Conflict("OPENAI_WEB_THREAD_ALREADY_EXISTS", "openai web thread already exists")
	ErrOpenAIWebThreadNilInput      = infraerrors.BadRequest("OPENAI_WEB_THREAD_NIL_INPUT", "openai web thread input cannot be nil")
	ErrOpenAIWebGroupAccessDenied   = infraerrors.Forbidden("OPENAI_WEB_GROUP_ACCESS_DENIED", "no active subscription found for the requested openai group")
	ErrOpenAIWebGroupInvalid        = infraerrors.BadRequest("OPENAI_WEB_GROUP_INVALID", "group is not eligible for openai web chat")
	ErrOpenAIWebNoAvailableAccounts = infraerrors.ServiceUnavailable("OPENAI_WEB_NO_AVAILABLE_ACCOUNTS", "no available openai web accounts")
	ErrOpenAIWebMessageEmpty        = infraerrors.BadRequest("OPENAI_WEB_MESSAGE_EMPTY", "message content or image attachment is required")
	ErrOpenAIWebAttachmentInvalid   = infraerrors.BadRequest("OPENAI_WEB_ATTACHMENT_INVALID", "invalid image attachment")
	ErrOpenAIWebAttachmentTooLarge  = infraerrors.BadRequest("OPENAI_WEB_ATTACHMENT_TOO_LARGE", "image attachment is too large")
	ErrOpenAIWebAttachmentOverflow  = infraerrors.BadRequest("OPENAI_WEB_ATTACHMENT_OVERFLOW", "too many image attachments")
	ErrOpenAIWebGatewayUnavailable  = infraerrors.ServiceUnavailable("OPENAI_WEB_GATEWAY_UNAVAILABLE", "openai web gateway is not fully configured")
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
	AccessMode      string
	SubscriptionID  *int64
	HasWebChat      bool
	HasProAccounts  bool
	DefaultModel    string
	CapabilityMode  string
	SubscriptionEnd *time.Time
}

type OpenAIWebGroupAccessService interface {
	GetAvailableGroups(ctx context.Context, userID int64) ([]Group, error)
}

type OpenAIWebAccountSelector interface {
	SelectAccountForModelWithExclusions(ctx context.Context, groupID *int64, sessionHash, requestedModel string, excludedIDs map[int64]struct{}) (*Account, error)
}

type OpenAIWebMessageGateway interface {
	ForwardOpenAIWebMessage(ctx context.Context, input *OpenAIWebForwardMessageInput) (*OpenAIWebForwardMessageResult, error)
	RecordUsage(ctx context.Context, input *OpenAIRecordUsageInput) error
}

type OpenAIWebThreadRepository interface {
	Create(ctx context.Context, thread *OpenAIWebThread) error
	GetByLocalThreadIDAndUserID(ctx context.Context, localThreadID string, userID int64) (*OpenAIWebThread, error)
	ListActiveByUserID(ctx context.Context, userID int64) ([]OpenAIWebThread, error)
	Update(ctx context.Context, thread *OpenAIWebThread) error
	ArchiveByLocalThreadIDAndUserID(ctx context.Context, localThreadID string, userID int64, archivedAt time.Time) error
}

type SendOpenAIWebThreadMessageInput struct {
	UserID          int64
	LocalThreadID   string
	RequestedModel  string
	ReasoningEffort string
	Content         string
	Attachments     []OpenAIWebThreadMessageAttachment
	UserAgent       string
	IPAddress       string
}

type OpenAIWebThreadMessageAttachment struct {
	FileName    string
	ContentType string
	DataURL     string
	Width       int
	Height      int
}

type OpenAIWebThreadMessageUsage struct {
	InputTokens         int
	OutputTokens        int
	CacheReadTokens     int
	CacheCreationTokens int
	ImageOutputTokens   int
	TotalTokens         int
}

type OpenAIWebThreadMessageImage struct {
	DataURL       string
	MimeType      string
	RevisedPrompt string
	Width         int
	Height        int
}

type OpenAIWebThreadMessageResult struct {
	Thread          *OpenAIWebThread
	AssistantText   string
	AssistantImages []OpenAIWebThreadMessageImage
	RequestID       string
	ResponseID      *string
	Model           string
	Usage           OpenAIWebThreadMessageUsage
}

type OpenAIWebForwardMessageInput struct {
	Thread          *OpenAIWebThread
	APIKey          *APIKey
	Account         *Account
	RequestedModel  string
	ReasoningEffort string
	Content         string
	Attachments     []OpenAIWebThreadMessageAttachment
	UserAgent       string
}

type OpenAIWebForwardMessageResult struct {
	Result                 *OpenAIForwardResult
	AssistantText          string
	AssistantImages        []OpenAIWebThreadMessageImage
	ResponseID             *string
	RequestPayloadHash     string
	UpstreamConversationID *string
	UpstreamSessionID      *string
}

func IsOpenAIWebInternalAPIKeyName(name string) bool {
	return strings.HasPrefix(strings.TrimSpace(strings.ToLower(name)), OpenAIWebInternalAPIKeyNamePrefix)
}
