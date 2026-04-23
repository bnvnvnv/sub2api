package service

import (
	"context"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/google/uuid"
)

type OpenAIWebThreadService struct {
	repo          OpenAIWebThreadRepository
	subscriptions *SubscriptionService
	accountRepo   AccountRepository
	openaiGateway *OpenAIGatewayService
}

type CreateOpenAIWebThreadInput struct {
	UserID         int64
	GroupID        int64
	RequestedModel string
	Title          string
	CachePolicy    string
}

func NewOpenAIWebThreadService(
	repo OpenAIWebThreadRepository,
	subscriptions *SubscriptionService,
	accountRepo AccountRepository,
	openaiGateway *OpenAIGatewayService,
) *OpenAIWebThreadService {
	return &OpenAIWebThreadService{
		repo:          repo,
		subscriptions: subscriptions,
		accountRepo:   accountRepo,
		openaiGateway: openaiGateway,
	}
}

func (s *OpenAIWebThreadService) ListUserThreads(ctx context.Context, userID int64) ([]OpenAIWebThread, error) {
	return s.repo.ListActiveByUserID(ctx, userID)
}

func (s *OpenAIWebThreadService) GetUserThread(ctx context.Context, userID int64, localThreadID string) (*OpenAIWebThread, error) {
	return s.repo.GetByLocalThreadIDAndUserID(ctx, strings.TrimSpace(localThreadID), userID)
}

func (s *OpenAIWebThreadService) ListUserEntitlements(ctx context.Context, userID int64) ([]OpenAIWebThreadEntitlement, error) {
	if s.subscriptions == nil || s.accountRepo == nil {
		return nil, nil
	}

	subs, err := s.subscriptions.ListActiveUserSubscriptions(ctx, userID)
	if err != nil {
		return nil, err
	}

	entitlements := make([]OpenAIWebThreadEntitlement, 0, len(subs))
	for i := range subs {
		sub := &subs[i]
		group := sub.Group
		if group == nil || group.Platform != PlatformOpenAI {
			continue
		}

		accounts, err := s.accountRepo.ListByGroup(ctx, group.ID)
		if err != nil {
			return nil, err
		}

		hasWebChat := false
		hasProAccounts := false
		for j := range accounts {
			account := &accounts[j]
			if account.SupportsOpenAIWebChat() {
				hasWebChat = true
			}
			if account.SupportsOpenAIWebModel("gpt-5.4-pro") {
				hasProAccounts = true
			}
			if hasWebChat && hasProAccounts {
				break
			}
		}

		if !hasWebChat {
			continue
		}

		entitlements = append(entitlements, OpenAIWebThreadEntitlement{
			GroupID:         group.ID,
			GroupName:       group.Name,
			GroupDesc:       group.Description,
			SubscriptionID:  sub.ID,
			HasWebChat:      hasWebChat,
			HasProAccounts:  hasProAccounts,
			DefaultModel:    defaultOpenAIWebRequestedModel(hasProAccounts),
			CapabilityMode:  OpenAIWebCapabilityModeWebChat,
			SubscriptionEnd: sub.ExpiresAt,
		})
	}

	return entitlements, nil
}

func (s *OpenAIWebThreadService) CreateUserThread(ctx context.Context, input *CreateOpenAIWebThreadInput) (*OpenAIWebThread, error) {
	if input == nil {
		return nil, ErrOpenAIWebThreadNilInput
	}
	if input.UserID <= 0 || input.GroupID <= 0 {
		return nil, infraerrors.BadRequest("OPENAI_WEB_THREAD_INVALID_INPUT", "user_id and group_id are required")
	}
	if s.subscriptions == nil || s.repo == nil || s.accountRepo == nil || s.openaiGateway == nil {
		return nil, infraerrors.ServiceUnavailable("OPENAI_WEB_THREAD_UNAVAILABLE", "openai web thread service is not fully configured")
	}

	sub, err := s.subscriptions.GetActiveSubscription(ctx, input.UserID, input.GroupID)
	if err != nil {
		return nil, ErrOpenAIWebGroupAccessDenied.WithCause(err)
	}
	if sub.Group == nil || sub.Group.Platform != PlatformOpenAI {
		return nil, ErrOpenAIWebGroupInvalid
	}

	requestedModel := strings.TrimSpace(input.RequestedModel)
	localThreadID := uuid.NewString()
	pageSessionID := uuid.NewString()
	account, err := s.selectInitialAccount(ctx, input.GroupID, pageSessionID, requestedModel)
	if err != nil {
		return nil, err
	}
	if requestedModel == "" {
		requestedModel = defaultOpenAIWebRequestedModelForAccount(account)
	}

	now := time.Now()
	thread := &OpenAIWebThread{
		UserID:         input.UserID,
		GroupID:        input.GroupID,
		AccountID:      account.ID,
		LocalThreadID:  localThreadID,
		PageSessionID:  pageSessionID,
		Provider:       OpenAIWebThreadProviderOpenAI,
		Title:          resolveOpenAIWebThreadTitle(input.Title),
		RequestedModel: requestedModel,
		CapabilityMode: resolveOpenAIWebCapabilityMode(requestedModel),
		HistoryMode:    OpenAIWebHistoryModeUpstreamOnly,
		CachePolicy:    resolveOpenAIWebCachePolicy(input.CachePolicy),
		SyncStatus:     OpenAIWebSyncStatusPending,
		Status:         OpenAIWebThreadStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
		Group:          sub.Group,
		Account:        account,
	}

	if err := s.repo.Create(ctx, thread); err != nil {
		return nil, err
	}
	return thread, nil
}

func (s *OpenAIWebThreadService) ArchiveUserThread(ctx context.Context, userID int64, localThreadID string) error {
	return s.repo.ArchiveByLocalThreadIDAndUserID(ctx, strings.TrimSpace(localThreadID), userID, time.Now())
}

func (s *OpenAIWebThreadService) selectInitialAccount(ctx context.Context, groupID int64, pageSessionID, requestedModel string) (*Account, error) {
	sessionHash := DeriveSessionHashFromSeed(pageSessionID)
	excludedIDs := map[int64]struct{}{}

	for {
		account, err := s.openaiGateway.SelectAccountForModelWithExclusions(ctx, &groupID, sessionHash, requestedModel, excludedIDs)
		if err != nil {
			return nil, ErrOpenAIWebNoAvailableAccounts.WithCause(err)
		}
		if account == nil {
			return nil, ErrOpenAIWebNoAvailableAccounts
		}
		if account.SupportsOpenAIWebModel(requestedModel) {
			return account, nil
		}
		excludedIDs[account.ID] = struct{}{}
	}
}

func resolveOpenAIWebThreadTitle(title string) string {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return "New Chat"
	}
	if len(trimmed) > 255 {
		return trimmed[:255]
	}
	return trimmed
}

func resolveOpenAIWebCapabilityMode(requestedModel string) string {
	if isOpenAIWebProModel(requestedModel) {
		return OpenAIWebCapabilityModeProModel
	}
	return OpenAIWebCapabilityModeWebChat
}

func resolveOpenAIWebCachePolicy(cachePolicy string) string {
	switch strings.TrimSpace(cachePolicy) {
	case OpenAIWebCachePolicyLocalEncrypted:
		return OpenAIWebCachePolicyLocalEncrypted
	default:
		return OpenAIWebCachePolicyLocalOnly
	}
}

func defaultOpenAIWebRequestedModel(hasProAccounts bool) string {
	if hasProAccounts {
		return "gpt-5.4"
	}
	return "gpt-5.4-mini"
}

func defaultOpenAIWebRequestedModelForAccount(account *Account) string {
	if account != nil && strings.Contains(account.GetOpenAIPlanType(), "pro") {
		return "gpt-5.4"
	}
	return "gpt-5.4-mini"
}
