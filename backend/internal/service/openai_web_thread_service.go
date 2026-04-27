package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/google/uuid"
)

type OpenAIWebThreadService struct {
	repo            OpenAIWebThreadRepository
	groupAccess     OpenAIWebGroupAccessService
	subscriptions   *SubscriptionService
	apiKeyRepo      APIKeyRepository
	accountRepo     AccountRepository
	accountSelector OpenAIWebAccountSelector
	gateway         OpenAIWebMessageGateway
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
	groupAccess OpenAIWebGroupAccessService,
	subscriptions *SubscriptionService,
	apiKeyRepo APIKeyRepository,
	accountRepo AccountRepository,
	accountSelector OpenAIWebAccountSelector,
	gateway OpenAIWebMessageGateway,
) *OpenAIWebThreadService {
	return &OpenAIWebThreadService{
		repo:            repo,
		groupAccess:     groupAccess,
		subscriptions:   subscriptions,
		apiKeyRepo:      apiKeyRepo,
		accountRepo:     accountRepo,
		accountSelector: accountSelector,
		gateway:         gateway,
	}
}

func (s *OpenAIWebThreadService) ListUserThreads(ctx context.Context, userID int64) ([]OpenAIWebThread, error) {
	return s.repo.ListActiveByUserID(ctx, userID)
}

func (s *OpenAIWebThreadService) GetUserThread(ctx context.Context, userID int64, localThreadID string) (*OpenAIWebThread, error) {
	return s.repo.GetByLocalThreadIDAndUserID(ctx, strings.TrimSpace(localThreadID), userID)
}

func (s *OpenAIWebThreadService) ListUserEntitlements(ctx context.Context, userID int64) ([]OpenAIWebThreadEntitlement, error) {
	if s.groupAccess == nil || s.accountRepo == nil {
		return nil, nil
	}

	availableGroups, err := s.groupAccess.GetAvailableGroups(ctx, userID)
	if err != nil {
		return nil, err
	}

	activeSubsByGroup := map[int64]*UserSubscription{}
	if s.subscriptions != nil {
		subs, err := s.subscriptions.ListActiveUserSubscriptions(ctx, userID)
		if err != nil {
			return nil, err
		}
		for i := range subs {
			sub := subs[i]
			activeSubsByGroup[sub.GroupID] = &sub
		}
	}

	entitlements := make([]OpenAIWebThreadEntitlement, 0, len(availableGroups))
	for i := range availableGroups {
		group := &availableGroups[i]
		if group.Platform != PlatformOpenAI {
			continue
		}

		subscriptionID, subscriptionEnd := resolveOpenAIWebSubscriptionAccess(group, activeSubsByGroup[group.ID])
		if group.IsSubscriptionType() && subscriptionID == nil {
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
			AccessMode:      resolveOpenAIWebAccessMode(group),
			SubscriptionID:  subscriptionID,
			HasWebChat:      hasWebChat,
			HasProAccounts:  hasProAccounts,
			DefaultModel:    defaultOpenAIWebRequestedModel(hasProAccounts),
			CapabilityMode:  OpenAIWebCapabilityModeWebChat,
			SubscriptionEnd: subscriptionEnd,
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
	if s.groupAccess == nil || s.repo == nil || s.accountRepo == nil || s.accountSelector == nil {
		return nil, infraerrors.ServiceUnavailable("OPENAI_WEB_THREAD_UNAVAILABLE", "openai web thread service is not fully configured")
	}

	group, err := s.resolveEligibleGroup(ctx, input.UserID, input.GroupID)
	if err != nil {
		if errors.Is(err, ErrOpenAIWebGroupInvalid) {
			return nil, err
		}
		return nil, ErrOpenAIWebGroupAccessDenied.WithCause(err)
	}
	if group == nil || group.Platform != PlatformOpenAI {
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
		Group:          group,
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

func (s *OpenAIWebThreadService) SendUserThreadMessage(ctx context.Context, input *SendOpenAIWebThreadMessageInput) (*OpenAIWebThreadMessageResult, error) {
	if input == nil {
		return nil, ErrOpenAIWebThreadNilInput
	}
	content := strings.TrimSpace(input.Content)
	if input.UserID <= 0 || strings.TrimSpace(input.LocalThreadID) == "" {
		return nil, infraerrors.BadRequest("OPENAI_WEB_MESSAGE_INVALID_INPUT", "user_id and local_thread_id are required")
	}
	if content == "" && len(input.Attachments) == 0 {
		return nil, ErrOpenAIWebMessageEmpty
	}
	if s.repo == nil || s.groupAccess == nil || s.accountSelector == nil || s.apiKeyRepo == nil || s.gateway == nil {
		return nil, ErrOpenAIWebGatewayUnavailable
	}

	thread, err := s.GetUserThread(ctx, input.UserID, input.LocalThreadID)
	if err != nil {
		return nil, err
	}
	if thread == nil || !thread.IsActive() {
		return nil, ErrOpenAIWebThreadNotFound
	}

	group, err := s.resolveEligibleGroup(ctx, input.UserID, thread.GroupID)
	if err != nil {
		s.markThreadSyncError(ctx, thread, nil, err)
		return nil, err
	}
	thread.Group = group

	requestedModel := strings.TrimSpace(input.RequestedModel)
	if requestedModel == "" {
		requestedModel = strings.TrimSpace(thread.RequestedModel)
	}

	account, err := s.selectInitialAccount(ctx, thread.GroupID, thread.PageSessionID, requestedModel)
	if err != nil {
		s.markThreadSyncError(ctx, thread, nil, err)
		return nil, err
	}
	if requestedModel == "" {
		requestedModel = defaultOpenAIWebRequestedModelForAccount(account)
	}
	reasoningEffort := resolveOpenAIWebReasoningEffort(input.ReasoningEffort)

	billingAPIKey, err := s.ensureInternalBillingAPIKey(ctx, input.UserID, thread.GroupID)
	if err != nil {
		s.markThreadSyncError(ctx, thread, account, err)
		return nil, err
	}

	forwarded, err := s.gateway.ForwardOpenAIWebMessage(ctx, &OpenAIWebForwardMessageInput{
		Thread:          thread,
		APIKey:          billingAPIKey,
		Account:         account,
		RequestedModel:  requestedModel,
		ReasoningEffort: reasoningEffort,
		Content:         content,
		Attachments:     append([]OpenAIWebThreadMessageAttachment(nil), input.Attachments...),
		UserAgent:       input.UserAgent,
	})
	if err != nil {
		s.markThreadSyncError(ctx, thread, account, err)
		return nil, err
	}

	subscription, subErr := s.resolveBillingSubscription(ctx, input.UserID, thread.GroupID)
	if subErr != nil {
		s.markThreadSyncError(ctx, thread, account, subErr)
		return nil, subErr
	}

	now := time.Now()
	thread.AccountID = account.ID
	thread.Account = account
	thread.Status = OpenAIWebThreadStatusActive
	thread.SyncStatus = OpenAIWebSyncStatusReady
	thread.LastError = ""
	thread.LastSyncedAt = &now
	thread.UpdatedAt = now
	thread.RequestedModel = requestedModel
	thread.CapabilityMode = resolveOpenAIWebCapabilityMode(requestedModel)
	if forwarded.UpstreamConversationID != nil {
		thread.UpstreamConversationID = forwarded.UpstreamConversationID
	}
	if forwarded.UpstreamSessionID != nil {
		thread.UpstreamSessionID = forwarded.UpstreamSessionID
	}
	if shouldAutofillOpenAIWebTitle(thread) {
		thread.Title = deriveOpenAIWebTitleFromInput(content, input.Attachments)
	}
	if err := s.repo.Update(ctx, thread); err != nil {
		return nil, err
	}

	user := billingAPIKey.User
	if user == nil {
		user = &User{ID: input.UserID}
	}

	if forwarded.Result != nil {
		upstreamEndpoint := strings.TrimSpace(forwarded.UpstreamEndpoint)
		if upstreamEndpoint == "" {
			upstreamEndpoint = "/backend-api/f/conversation"
		}
		if err := s.gateway.RecordUsage(ctx, &OpenAIRecordUsageInput{
			Result:             forwarded.Result,
			APIKey:             billingAPIKey,
			User:               user,
			Account:            account,
			Subscription:       subscription,
			InboundEndpoint:    "/openai-web/threads/:id/messages",
			UpstreamEndpoint:   upstreamEndpoint,
			UserAgent:          input.UserAgent,
			IPAddress:          input.IPAddress,
			RequestPayloadHash: forwarded.RequestPayloadHash,
			ChannelUsageFields: ChannelUsageFields{
				OriginalModel: requestedModel,
			},
		}); err != nil {
			thread.LastError = truncateOpenAIWebError(err)
			thread.UpdatedAt = time.Now()
			_ = s.repo.Update(ctx, thread)
		}
	}

	return &OpenAIWebThreadMessageResult{
		Thread:          thread,
		AssistantText:   forwarded.AssistantText,
		AssistantImages: append([]OpenAIWebThreadMessageImage(nil), forwarded.AssistantImages...),
		RequestID:       openAIWebForwardRequestID(forwarded.Result),
		ResponseID:      forwarded.ResponseID,
		Model:           resolveOpenAIWebResponseModel(thread, forwarded.Result),
		Usage:           openAIWebMessageUsageFromForwardResult(forwarded.Result),
	}, nil
}

func (s *OpenAIWebThreadService) selectInitialAccount(ctx context.Context, groupID int64, pageSessionID, requestedModel string) (*Account, error) {
	sessionHash := DeriveSessionHashFromSeed(pageSessionID)
	excludedIDs := map[int64]struct{}{}

	for {
		account, err := s.accountSelector.SelectAccountForModelWithExclusions(ctx, &groupID, sessionHash, requestedModel, excludedIDs)
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

func (s *OpenAIWebThreadService) resolveEligibleGroup(ctx context.Context, userID, groupID int64) (*Group, error) {
	availableGroups, err := s.groupAccess.GetAvailableGroups(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range availableGroups {
		group := availableGroups[i]
		if group.ID != groupID {
			continue
		}
		if group.Platform != PlatformOpenAI {
			return nil, ErrOpenAIWebGroupInvalid
		}
		if group.IsSubscriptionType() {
			if s.subscriptions == nil {
				return nil, ErrOpenAIWebGroupAccessDenied
			}
			sub, err := s.subscriptions.GetActiveSubscription(ctx, userID, groupID)
			if err != nil {
				return nil, err
			}
			if sub.Group != nil {
				return sub.Group, nil
			}
		}
		return &group, nil
	}

	return nil, ErrOpenAIWebGroupAccessDenied
}

func resolveOpenAIWebAccessMode(group *Group) string {
	if group == nil || !group.IsSubscriptionType() {
		return SubscriptionTypeStandard
	}
	return SubscriptionTypeSubscription
}

func resolveOpenAIWebSubscriptionAccess(group *Group, sub *UserSubscription) (*int64, *time.Time) {
	if group == nil || !group.IsSubscriptionType() || sub == nil {
		return nil, nil
	}
	subscriptionID := sub.ID
	subscriptionEnd := sub.ExpiresAt
	return &subscriptionID, &subscriptionEnd
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

func resolveOpenAIWebReasoningEffort(raw string) string {
	return normalizeOpenAIReasoningEffort(raw)
}

func (s *OpenAIWebThreadService) resolveBillingSubscription(ctx context.Context, userID, groupID int64) (*UserSubscription, error) {
	if s.subscriptions == nil {
		return nil, nil
	}
	sub, err := s.subscriptions.GetActiveSubscription(ctx, userID, groupID)
	if err != nil {
		if errors.Is(err, ErrSubscriptionNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return sub, nil
}

func (s *OpenAIWebThreadService) ensureInternalBillingAPIKey(ctx context.Context, userID, groupID int64) (*APIKey, error) {
	if s.apiKeyRepo == nil {
		return nil, ErrOpenAIWebGatewayUnavailable
	}

	keyValue := buildOpenAIWebInternalAPIKey(userID, groupID)
	existing, err := s.apiKeyRepo.GetByKey(ctx, keyValue)
	if err == nil && existing != nil {
		return existing, nil
	}
	if err != nil && err != ErrAPIKeyNotFound {
		return nil, err
	}

	groupIDCopy := groupID
	internalKey := &APIKey{
		UserID:  userID,
		Key:     keyValue,
		Name:    buildOpenAIWebInternalAPIKeyName(groupID),
		GroupID: &groupIDCopy,
		Status:  StatusAPIKeyDisabled,
	}
	if err := s.apiKeyRepo.Create(ctx, internalKey); err != nil && !errors.Is(err, ErrAPIKeyExists) {
		return nil, err
	}

	return s.apiKeyRepo.GetByKey(ctx, keyValue)
}

func (s *OpenAIWebThreadService) markThreadSyncError(ctx context.Context, thread *OpenAIWebThread, account *Account, err error) {
	if thread == nil || s.repo == nil || err == nil {
		return
	}
	now := time.Now()
	if account != nil {
		thread.AccountID = account.ID
		thread.Account = account
	}
	thread.SyncStatus = OpenAIWebSyncStatusError
	thread.UpdatedAt = now
	thread.LastError = truncateOpenAIWebError(err)
	if errors.Is(err, ErrOpenAIWebNoAvailableAccounts) || errors.Is(err, ErrOpenAIWebGroupAccessDenied) {
		thread.Status = OpenAIWebThreadStatusBroken
	}
	_ = s.repo.Update(ctx, thread)
}

func buildOpenAIWebInternalAPIKey(userID, groupID int64) string {
	return fmt.Sprintf("sub2api_openai_web_u%d_g%d_internal", userID, groupID)
}

func buildOpenAIWebInternalAPIKeyName(groupID int64) string {
	return fmt.Sprintf("%s_group_%d", OpenAIWebInternalAPIKeyNamePrefix, groupID)
}

func shouldAutofillOpenAIWebTitle(thread *OpenAIWebThread) bool {
	if thread == nil {
		return false
	}
	title := strings.TrimSpace(thread.Title)
	return title == "" || strings.EqualFold(title, "New Chat")
}

func deriveOpenAIWebTitleFromMessage(content string) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(content, "\n", " "))
	if trimmed == "" {
		return "New Chat"
	}
	runes := []rune(trimmed)
	if len(runes) > 72 {
		trimmed = string(runes[:72])
	}
	return resolveOpenAIWebThreadTitle(trimmed)
}

func deriveOpenAIWebTitleFromInput(content string, attachments []OpenAIWebThreadMessageAttachment) string {
	if strings.TrimSpace(content) != "" {
		return deriveOpenAIWebTitleFromMessage(content)
	}
	for _, attachment := range attachments {
		name := strings.TrimSpace(attachment.FileName)
		if name == "" {
			continue
		}
		if dot := strings.LastIndex(name, "."); dot > 0 {
			name = name[:dot]
		}
		name = strings.TrimSpace(name)
		if name != "" {
			return resolveOpenAIWebThreadTitle(name)
		}
	}
	return resolveOpenAIWebThreadTitle("Image Chat")
}

func truncateOpenAIWebError(err error) string {
	if err == nil {
		return ""
	}
	text := strings.TrimSpace(err.Error())
	if len(text) > 500 {
		return text[:500]
	}
	return text
}

func resolveOpenAIWebResponseModel(thread *OpenAIWebThread, result *OpenAIForwardResult) string {
	if result != nil {
		if model := strings.TrimSpace(result.Model); model != "" {
			return model
		}
	}
	if thread != nil {
		return strings.TrimSpace(thread.RequestedModel)
	}
	return ""
}

func openAIWebMessageUsageFromForwardResult(result *OpenAIForwardResult) OpenAIWebThreadMessageUsage {
	if result == nil {
		return OpenAIWebThreadMessageUsage{}
	}
	usage := result.Usage
	return OpenAIWebThreadMessageUsage{
		InputTokens:         usage.InputTokens,
		OutputTokens:        usage.OutputTokens,
		CacheReadTokens:     usage.CacheReadInputTokens,
		CacheCreationTokens: usage.CacheCreationInputTokens,
		ImageOutputTokens:   usage.ImageOutputTokens,
		TotalTokens: usage.InputTokens +
			usage.OutputTokens +
			usage.CacheReadInputTokens +
			usage.CacheCreationInputTokens +
			usage.ImageOutputTokens,
	}
}

func openAIWebForwardRequestID(result *OpenAIForwardResult) string {
	if result == nil {
		return ""
	}
	return strings.TrimSpace(result.RequestID)
}
