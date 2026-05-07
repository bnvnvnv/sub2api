package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var _ OpenAIWebGroupAccessService = (*openAIWebGroupAccessStub)(nil)
var _ OpenAIWebAccountSelector = (*openAIWebAccountSelectorStub)(nil)
var _ OpenAIWebThreadRepository = (*openAIWebThreadRepoStub)(nil)
var _ UserSubscriptionRepository = (*openAIWebUserSubRepoStub)(nil)
var _ AccountRepository = (*openAIWebAccountRepoStub)(nil)
var _ APIKeyRepository = (*openAIWebAPIKeyRepoStub)(nil)
var _ OpenAIWebMessageGateway = (*openAIWebMessageGatewayStub)(nil)

type openAIWebGroupAccessStub struct {
	groups []Group
	err    error
}

func (s openAIWebGroupAccessStub) GetAvailableGroups(ctx context.Context, userID int64) ([]Group, error) {
	if s.err != nil {
		return nil, s.err
	}
	return append([]Group(nil), s.groups...), nil
}

type openAIWebAccountSelectorStub struct {
	account *Account
	err     error
}

func (s openAIWebAccountSelectorStub) SelectAccountForModelWithExclusions(ctx context.Context, groupID *int64, sessionHash, requestedModel string, excludedIDs map[int64]struct{}) (*Account, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.account, nil
}

type openAIWebThreadRepoStub struct {
	OpenAIWebThreadRepository
	created      []*OpenAIWebThread
	updated      []*OpenAIWebThread
	threadsByKey map[string]*OpenAIWebThread
}

func (r *openAIWebThreadRepoStub) Create(ctx context.Context, thread *OpenAIWebThread) error {
	cp := *thread
	cp.ID = int64(len(r.created) + 1)
	thread.ID = cp.ID
	if r.threadsByKey == nil {
		r.threadsByKey = map[string]*OpenAIWebThread{}
	}
	r.threadsByKey[cp.LocalThreadID] = &cp
	r.created = append(r.created, &cp)
	return nil
}

func (r *openAIWebThreadRepoStub) GetByLocalThreadIDAndUserID(ctx context.Context, localThreadID string, userID int64) (*OpenAIWebThread, error) {
	if r.threadsByKey == nil {
		return nil, ErrOpenAIWebThreadNotFound
	}
	thread, ok := r.threadsByKey[localThreadID]
	if !ok || thread.UserID != userID {
		return nil, ErrOpenAIWebThreadNotFound
	}
	cp := *thread
	return &cp, nil
}

func (r *openAIWebThreadRepoStub) ListActiveByUserID(ctx context.Context, userID int64) ([]OpenAIWebThread, error) {
	out := make([]OpenAIWebThread, 0)
	for _, thread := range r.threadsByKey {
		if thread.UserID == userID && thread.DeletedAt == nil {
			out = append(out, *thread)
		}
	}
	return out, nil
}

func (r *openAIWebThreadRepoStub) Update(ctx context.Context, thread *OpenAIWebThread) error {
	if r.threadsByKey == nil {
		r.threadsByKey = map[string]*OpenAIWebThread{}
	}
	cp := *thread
	r.threadsByKey[cp.LocalThreadID] = &cp
	r.updated = append(r.updated, &cp)
	return nil
}

func (r *openAIWebThreadRepoStub) ArchiveByLocalThreadIDAndUserID(ctx context.Context, localThreadID string, userID int64, archivedAt time.Time) error {
	thread, err := r.GetByLocalThreadIDAndUserID(ctx, localThreadID, userID)
	if err != nil {
		return err
	}
	thread.Status = OpenAIWebThreadStatusArchived
	thread.DeletedAt = &archivedAt
	thread.UpdatedAt = archivedAt
	return r.Update(ctx, thread)
}

type openAIWebUserSubRepoStub struct {
	UserSubscriptionRepository
	activeSubs []UserSubscription
}

func (r openAIWebUserSubRepoStub) ListActiveByUserID(ctx context.Context, userID int64) ([]UserSubscription, error) {
	return append([]UserSubscription(nil), r.activeSubs...), nil
}

func (r openAIWebUserSubRepoStub) GetActiveByUserIDAndGroupID(ctx context.Context, userID, groupID int64) (*UserSubscription, error) {
	for i := range r.activeSubs {
		if r.activeSubs[i].UserID == userID && r.activeSubs[i].GroupID == groupID {
			sub := r.activeSubs[i]
			return &sub, nil
		}
	}
	return nil, ErrSubscriptionNotFound
}

type openAIWebAccountRepoStub struct {
	AccountRepository
	accountsByGroup map[int64][]Account
}

func (r openAIWebAccountRepoStub) ListByGroup(ctx context.Context, groupID int64) ([]Account, error) {
	accounts := r.accountsByGroup[groupID]
	return append([]Account(nil), accounts...), nil
}

type openAIWebAPIKeyRepoStub struct {
	APIKeyRepository
	byKey map[string]*APIKey
}

func (r *openAIWebAPIKeyRepoStub) GetByKey(ctx context.Context, key string) (*APIKey, error) {
	if r.byKey == nil {
		return nil, ErrAPIKeyNotFound
	}
	apiKey, ok := r.byKey[key]
	if !ok {
		return nil, ErrAPIKeyNotFound
	}
	cp := *apiKey
	return &cp, nil
}

func (r *openAIWebAPIKeyRepoStub) Create(ctx context.Context, key *APIKey) error {
	if r.byKey == nil {
		r.byKey = map[string]*APIKey{}
	}
	if _, exists := r.byKey[key.Key]; exists {
		return ErrAPIKeyExists
	}
	cp := *key
	cp.ID = int64(len(r.byKey) + 1)
	if cp.User == nil {
		cp.User = &User{ID: cp.UserID}
	}
	if cp.GroupID != nil && cp.Group == nil {
		cp.Group = &Group{
			ID:               *cp.GroupID,
			Platform:         PlatformOpenAI,
			Status:           StatusActive,
			SubscriptionType: SubscriptionTypeStandard,
		}
	}
	key.ID = cp.ID
	r.byKey[key.Key] = &cp
	return nil
}

type openAIWebMessageGatewayStub struct {
	OpenAIWebMessageGateway
	forward          *OpenAIWebForwardMessageResult
	forwardErr       error
	recordUsageErr   error
	forwardCalls     int
	recordUsageCalls int
	lastForward      *OpenAIWebForwardMessageInput
	lastRecordUsage  *OpenAIRecordUsageInput
}

func (s *openAIWebMessageGatewayStub) ForwardOpenAIWebMessage(ctx context.Context, input *OpenAIWebForwardMessageInput) (*OpenAIWebForwardMessageResult, error) {
	s.forwardCalls++
	s.lastForward = input
	if s.forwardErr != nil {
		return nil, s.forwardErr
	}
	return s.forward, nil
}

func (s *openAIWebMessageGatewayStub) RecordUsage(ctx context.Context, input *OpenAIRecordUsageInput) error {
	s.recordUsageCalls++
	s.lastRecordUsage = input
	return s.recordUsageErr
}

func TestOpenAIWebThreadService_ListUserEntitlementsIncludesStandardGroup(t *testing.T) {
	group := Group{
		ID:               11,
		Name:             "openai-public",
		Platform:         PlatformOpenAI,
		Status:           StatusActive,
		SubscriptionType: SubscriptionTypeStandard,
	}
	account := Account{
		ID:       101,
		Name:     "openai-oauth",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":       "token",
			"chatgpt_account_id": "acc_public",
		},
	}

	svc := NewOpenAIWebThreadService(
		&openAIWebThreadRepoStub{},
		openAIWebGroupAccessStub{groups: []Group{
			group,
			{ID: 88, Name: "claude", Platform: PlatformAnthropic, Status: StatusActive, SubscriptionType: SubscriptionTypeStandard},
		}},
		NewSubscriptionService(nil, openAIWebUserSubRepoStub{}, nil, nil, nil),
		&openAIWebAPIKeyRepoStub{},
		openAIWebAccountRepoStub{accountsByGroup: map[int64][]Account{
			group.ID: {account},
		}},
		nil,
		nil,
	)

	entitlements, err := svc.ListUserEntitlements(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, entitlements, 1)
	require.Equal(t, group.ID, entitlements[0].GroupID)
	require.Equal(t, SubscriptionTypeStandard, entitlements[0].AccessMode)
	require.Nil(t, entitlements[0].SubscriptionID)
	require.Nil(t, entitlements[0].SubscriptionEnd)
	require.True(t, entitlements[0].HasWebChat)
	require.False(t, entitlements[0].HasProAccounts)
}

func TestOpenAIWebThreadService_ListUserEntitlementsIncludesSubscriptionMetadata(t *testing.T) {
	expiresAt := time.Now().Add(48 * time.Hour)
	group := Group{
		ID:               22,
		Name:             "openai-pro",
		Platform:         PlatformOpenAI,
		Status:           StatusActive,
		SubscriptionType: SubscriptionTypeSubscription,
	}
	sub := UserSubscription{
		ID:        9001,
		UserID:    7,
		GroupID:   group.ID,
		Status:    SubscriptionStatusActive,
		ExpiresAt: expiresAt,
		Group:     &group,
	}
	account := Account{
		ID:       202,
		Name:     "openai-pro-account",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":       "token",
			"chatgpt_account_id": "acc_pro",
			"plan_type":          "pro",
		},
	}

	svc := NewOpenAIWebThreadService(
		&openAIWebThreadRepoStub{},
		openAIWebGroupAccessStub{groups: []Group{group}},
		NewSubscriptionService(nil, openAIWebUserSubRepoStub{activeSubs: []UserSubscription{sub}}, nil, nil, nil),
		&openAIWebAPIKeyRepoStub{},
		openAIWebAccountRepoStub{accountsByGroup: map[int64][]Account{
			group.ID: {account},
		}},
		nil,
		nil,
	)

	entitlements, err := svc.ListUserEntitlements(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, entitlements, 1)
	require.Equal(t, SubscriptionTypeSubscription, entitlements[0].AccessMode)
	require.NotNil(t, entitlements[0].SubscriptionID)
	require.Equal(t, sub.ID, *entitlements[0].SubscriptionID)
	require.NotNil(t, entitlements[0].SubscriptionEnd)
	require.WithinDuration(t, expiresAt, *entitlements[0].SubscriptionEnd, time.Second)
	require.True(t, entitlements[0].HasProAccounts)
}

func TestOpenAIWebThreadService_CreateUserThreadAllowsStandardGroupWithoutSubscription(t *testing.T) {
	group := Group{
		ID:               33,
		Name:             "openai-standard",
		Platform:         PlatformOpenAI,
		Status:           StatusActive,
		SubscriptionType: SubscriptionTypeStandard,
	}
	account := &Account{
		ID:       303,
		Name:     "openai-standard-account",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":       "token",
			"chatgpt_account_id": "acc_standard",
		},
	}
	repo := &openAIWebThreadRepoStub{}

	svc := NewOpenAIWebThreadService(
		repo,
		openAIWebGroupAccessStub{groups: []Group{group}},
		NewSubscriptionService(nil, openAIWebUserSubRepoStub{}, nil, nil, nil),
		&openAIWebAPIKeyRepoStub{},
		openAIWebAccountRepoStub{accountsByGroup: map[int64][]Account{
			group.ID: {*account},
		}},
		openAIWebAccountSelectorStub{account: account},
		nil,
	)

	thread, err := svc.CreateUserThread(context.Background(), &CreateOpenAIWebThreadInput{
		UserID:      1,
		GroupID:     group.ID,
		CachePolicy: OpenAIWebCachePolicyLocalEncrypted,
	})
	require.NoError(t, err)
	require.NotNil(t, thread)
	require.Equal(t, group.ID, thread.GroupID)
	require.Equal(t, account.ID, thread.AccountID)
	require.Equal(t, OpenAIWebCachePolicyLocalEncrypted, thread.CachePolicy)
	require.Equal(t, "gpt-5.4-mini", thread.RequestedModel)
	require.Equal(t, group.ID, repo.created[0].GroupID)
	require.Equal(t, account.ID, repo.created[0].AccountID)
	require.NotEmpty(t, thread.LocalThreadID)
	require.NotEmpty(t, thread.PageSessionID)
}

func TestOpenAIWebThreadService_CreateUserThreadRejectsUnavailableGroup(t *testing.T) {
	svc := NewOpenAIWebThreadService(
		&openAIWebThreadRepoStub{},
		openAIWebGroupAccessStub{groups: nil},
		NewSubscriptionService(nil, openAIWebUserSubRepoStub{}, nil, nil, nil),
		&openAIWebAPIKeyRepoStub{},
		openAIWebAccountRepoStub{},
		openAIWebAccountSelectorStub{},
		nil,
	)

	_, err := svc.CreateUserThread(context.Background(), &CreateOpenAIWebThreadInput{
		UserID:  1,
		GroupID: 404,
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrOpenAIWebGroupAccessDenied))
}

func TestOpenAIWebThreadService_SendUserThreadMessageUpdatesThreadAndRecordsUsage(t *testing.T) {
	group := Group{
		ID:               44,
		Name:             "openai-web",
		Platform:         PlatformOpenAI,
		Status:           StatusActive,
		SubscriptionType: SubscriptionTypeStandard,
	}
	account := &Account{
		ID:       404,
		Name:     "openai-web-account",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":       "token",
			"chatgpt_account_id": "acc_web",
		},
	}
	threadRepo := &openAIWebThreadRepoStub{
		threadsByKey: map[string]*OpenAIWebThread{
			"thread-1": {
				ID:             1,
				UserID:         7,
				GroupID:        group.ID,
				AccountID:      account.ID,
				LocalThreadID:  "thread-1",
				PageSessionID:  "page-session-1",
				Provider:       OpenAIWebThreadProviderOpenAI,
				Title:          "New Chat",
				RequestedModel: "gpt-5.4-mini",
				CapabilityMode: OpenAIWebCapabilityModeWebChat,
				HistoryMode:    OpenAIWebHistoryModeUpstreamOnly,
				CachePolicy:    OpenAIWebCachePolicyLocalOnly,
				SyncStatus:     OpenAIWebSyncStatusPending,
				Status:         OpenAIWebThreadStatusActive,
				CreatedAt:      time.Now().Add(-time.Hour),
				UpdatedAt:      time.Now().Add(-time.Hour),
				Group:          &group,
				Account:        account,
			},
		},
	}
	apiKeyRepo := &openAIWebAPIKeyRepoStub{}
	gateway := &openAIWebMessageGatewayStub{
		forward: &OpenAIWebForwardMessageResult{
			Result: &OpenAIForwardResult{
				RequestID: "req_oweb_1",
				Usage: OpenAIUsage{
					InputTokens:  18,
					OutputTokens: 9,
				},
				Model:    "gpt-5.4-mini",
				Duration: time.Second,
			},
			AssistantText:          "已收到。",
			RequestPayloadHash:     "payload_hash_1",
			UpstreamConversationID: openAIWebStrPtr("conv_up_1"),
			UpstreamSessionID:      openAIWebStrPtr("sess_up_1"),
			ResponseID:             openAIWebStrPtr("resp_1"),
		},
	}

	svc := NewOpenAIWebThreadService(
		threadRepo,
		openAIWebGroupAccessStub{groups: []Group{group}},
		NewSubscriptionService(nil, openAIWebUserSubRepoStub{}, nil, nil, nil),
		apiKeyRepo,
		openAIWebAccountRepoStub{accountsByGroup: map[int64][]Account{
			group.ID: {*account},
		}},
		openAIWebAccountSelectorStub{account: account},
		gateway,
	)

	result, err := svc.SendUserThreadMessage(context.Background(), &SendOpenAIWebThreadMessageInput{
		UserID:        7,
		LocalThreadID: "thread-1",
		Content:       "帮我整理今天的升级计划",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "127.0.0.1",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "已收到。", result.AssistantText)
	require.Equal(t, "req_oweb_1", result.RequestID)
	require.Equal(t, "gpt-5.4-mini", result.Model)
	require.Equal(t, 27, result.Usage.TotalTokens)
	require.NotNil(t, result.Thread)
	require.Equal(t, OpenAIWebSyncStatusReady, result.Thread.SyncStatus)
	require.Empty(t, result.Thread.LastError)
	require.Equal(t, "conv_up_1", derefString(result.Thread.UpstreamConversationID))
	require.Equal(t, "sess_up_1", derefString(result.Thread.UpstreamSessionID))
	require.Equal(t, "帮我整理今天的升级计划", result.Thread.Title)
	require.Len(t, threadRepo.updated, 1)
	require.Equal(t, 1, gateway.forwardCalls)
	require.Equal(t, 1, gateway.recordUsageCalls)
	require.NotNil(t, gateway.lastRecordUsage)
	require.Equal(t, "payload_hash_1", gateway.lastRecordUsage.RequestPayloadHash)
	require.Equal(t, "/backend-api/f/conversation", gateway.lastRecordUsage.UpstreamEndpoint)
	require.Equal(t, "gpt-5.4-mini", gateway.lastRecordUsage.OriginalModel)
	require.NotNil(t, gateway.lastRecordUsage.APIKey)
	require.Equal(t, StatusAPIKeyDisabled, gateway.lastRecordUsage.APIKey.Status)
	require.Equal(t, group.ID, derefInt64(gateway.lastRecordUsage.APIKey.GroupID))
}

func TestOpenAIWebThreadService_SendUserThreadMessageAllowsImageOnlyInput(t *testing.T) {
	group := Group{
		ID:               55,
		Name:             "openai-web",
		Platform:         PlatformOpenAI,
		Status:           StatusActive,
		SubscriptionType: SubscriptionTypeStandard,
	}
	account := &Account{
		ID:       505,
		Name:     "openai-web-account",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":       "token",
			"chatgpt_account_id": "acc_web",
		},
	}
	threadRepo := &openAIWebThreadRepoStub{
		threadsByKey: map[string]*OpenAIWebThread{
			"thread-img": {
				ID:             2,
				UserID:         8,
				GroupID:        group.ID,
				AccountID:      account.ID,
				LocalThreadID:  "thread-img",
				PageSessionID:  "page-session-img",
				Provider:       OpenAIWebThreadProviderOpenAI,
				Title:          "New Chat",
				RequestedModel: "gpt-5.4-mini",
				CapabilityMode: OpenAIWebCapabilityModeWebChat,
				HistoryMode:    OpenAIWebHistoryModeUpstreamOnly,
				CachePolicy:    OpenAIWebCachePolicyLocalOnly,
				SyncStatus:     OpenAIWebSyncStatusPending,
				Status:         OpenAIWebThreadStatusActive,
				CreatedAt:      time.Now().Add(-time.Hour),
				UpdatedAt:      time.Now().Add(-time.Hour),
				Group:          &group,
				Account:        account,
			},
		},
	}
	apiKeyRepo := &openAIWebAPIKeyRepoStub{}
	gateway := &openAIWebMessageGatewayStub{
		forward: &OpenAIWebForwardMessageResult{
			Result: &OpenAIForwardResult{
				RequestID: "req_oweb_img",
				Model:     "gpt-5.4-mini",
				Duration:  time.Second,
			},
			AssistantText: "已分析图片。",
		},
	}

	svc := NewOpenAIWebThreadService(
		threadRepo,
		openAIWebGroupAccessStub{groups: []Group{group}},
		NewSubscriptionService(nil, openAIWebUserSubRepoStub{}, nil, nil, nil),
		apiKeyRepo,
		openAIWebAccountRepoStub{accountsByGroup: map[int64][]Account{
			group.ID: {*account},
		}},
		openAIWebAccountSelectorStub{account: account},
		gateway,
	)

	result, err := svc.SendUserThreadMessage(context.Background(), &SendOpenAIWebThreadMessageInput{
		UserID:        8,
		LocalThreadID: "thread-img",
		Attachments: []OpenAIWebThreadMessageAttachment{
			{
				FileName:    "cat.png",
				ContentType: "image/png",
				DataURL:     "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO7Z6bQAAAAASUVORK5CYII=",
			},
		},
		UserAgent: "Mozilla/5.0",
		IPAddress: "127.0.0.1",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "已分析图片。", result.AssistantText)
	require.NotNil(t, gateway.lastForward)
	require.Len(t, gateway.lastForward.Attachments, 1)
	require.Equal(t, "cat.png", gateway.lastForward.Attachments[0].FileName)
	require.Equal(t, "cat", result.Thread.Title)
}

func TestOpenAIWebThreadService_SendUserThreadMessageAppliesModelAndReasoningOverrides(t *testing.T) {
	group := Group{
		ID:               66,
		Name:             "openai-web",
		Platform:         PlatformOpenAI,
		Status:           StatusActive,
		SubscriptionType: SubscriptionTypeStandard,
	}
	account := &Account{
		ID:       606,
		Name:     "openai-pro-account",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":       "token",
			"chatgpt_account_id": "acc_pro_web",
			"plan_type":          "pro",
		},
	}
	threadRepo := &openAIWebThreadRepoStub{
		threadsByKey: map[string]*OpenAIWebThread{
			"thread-override": {
				ID:             3,
				UserID:         9,
				GroupID:        group.ID,
				AccountID:      account.ID,
				LocalThreadID:  "thread-override",
				PageSessionID:  "page-session-override",
				Provider:       OpenAIWebThreadProviderOpenAI,
				Title:          "Existing Chat",
				RequestedModel: "gpt-5.4-mini",
				CapabilityMode: OpenAIWebCapabilityModeWebChat,
				HistoryMode:    OpenAIWebHistoryModeUpstreamOnly,
				CachePolicy:    OpenAIWebCachePolicyLocalOnly,
				SyncStatus:     OpenAIWebSyncStatusPending,
				Status:         OpenAIWebThreadStatusActive,
				CreatedAt:      time.Now().Add(-time.Hour),
				UpdatedAt:      time.Now().Add(-time.Hour),
				Group:          &group,
				Account:        account,
			},
		},
	}
	apiKeyRepo := &openAIWebAPIKeyRepoStub{}
	reasoning := "xhigh"
	gateway := &openAIWebMessageGatewayStub{
		forward: &OpenAIWebForwardMessageResult{
			Result: &OpenAIForwardResult{
				RequestID:       "req_oweb_override",
				Model:           "gpt-5.4-pro",
				Duration:        time.Second,
				ReasoningEffort: &reasoning,
			},
			AssistantText:      "已切换到更强模式。",
			RequestPayloadHash: "payload_hash_override",
		},
	}

	svc := NewOpenAIWebThreadService(
		threadRepo,
		openAIWebGroupAccessStub{groups: []Group{group}},
		NewSubscriptionService(nil, openAIWebUserSubRepoStub{}, nil, nil, nil),
		apiKeyRepo,
		openAIWebAccountRepoStub{accountsByGroup: map[int64][]Account{
			group.ID: {*account},
		}},
		openAIWebAccountSelectorStub{account: account},
		gateway,
	)

	result, err := svc.SendUserThreadMessage(context.Background(), &SendOpenAIWebThreadMessageInput{
		UserID:          9,
		LocalThreadID:   "thread-override",
		RequestedModel:  "gpt-5.4-pro",
		ReasoningEffort: "x-high",
		Content:         "给我最强推理。",
		UserAgent:       "Mozilla/5.0",
		IPAddress:       "127.0.0.1",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, gateway.lastForward)
	require.Equal(t, "gpt-5.4-pro", gateway.lastForward.RequestedModel)
	require.Equal(t, "xhigh", gateway.lastForward.ReasoningEffort)
	require.Equal(t, "gpt-5.4-pro", result.Thread.RequestedModel)
	require.Equal(t, OpenAIWebCapabilityModeProModel, result.Thread.CapabilityMode)
	require.NotNil(t, gateway.lastRecordUsage)
	require.Equal(t, "gpt-5.4-pro", gateway.lastRecordUsage.OriginalModel)
	require.NotNil(t, gateway.lastRecordUsage.Result)
	require.NotNil(t, gateway.lastRecordUsage.Result.ReasoningEffort)
	require.Equal(t, "xhigh", *gateway.lastRecordUsage.Result.ReasoningEffort)
}

func openAIWebStrPtr(v string) *string {
	return &v
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}
