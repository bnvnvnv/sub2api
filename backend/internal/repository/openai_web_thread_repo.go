package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/openaiwebthread"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type openAIWebThreadRepository struct {
	client *dbent.Client
}

func NewOpenAIWebThreadRepository(client *dbent.Client) service.OpenAIWebThreadRepository {
	return &openAIWebThreadRepository{client: client}
}

func (r *openAIWebThreadRepository) Create(ctx context.Context, thread *service.OpenAIWebThread) error {
	if thread == nil {
		return service.ErrOpenAIWebThreadNilInput
	}

	client := clientFromContext(ctx, r.client)
	builder := client.OpenAIWebThread.Create().
		SetUserID(thread.UserID).
		SetGroupID(thread.GroupID).
		SetAccountID(thread.AccountID).
		SetLocalThreadID(thread.LocalThreadID).
		SetPageSessionID(thread.PageSessionID).
		SetProvider(thread.Provider).
		SetTitle(thread.Title).
		SetRequestedModel(thread.RequestedModel).
		SetCapabilityMode(thread.CapabilityMode).
		SetHistoryMode(thread.HistoryMode).
		SetCachePolicy(thread.CachePolicy).
		SetSyncStatus(thread.SyncStatus).
		SetStatus(thread.Status).
		SetLastError(thread.LastError).
		SetNillableUpstreamConversationID(thread.UpstreamConversationID).
		SetNillableUpstreamSessionID(thread.UpstreamSessionID).
		SetNillableLastSyncedAt(thread.LastSyncedAt).
		SetNillableDeletedAt(thread.DeletedAt)

	if !thread.CreatedAt.IsZero() {
		builder.SetCreatedAt(thread.CreatedAt)
	}
	if !thread.UpdatedAt.IsZero() {
		builder.SetUpdatedAt(thread.UpdatedAt)
	}

	created, err := builder.Save(ctx)
	if err == nil {
		applyOpenAIWebThreadEntityToService(thread, created)
	}
	return translatePersistenceError(err, nil, service.ErrOpenAIWebThreadAlreadyExists)
}

func (r *openAIWebThreadRepository) GetByLocalThreadIDAndUserID(ctx context.Context, localThreadID string, userID int64) (*service.OpenAIWebThread, error) {
	client := clientFromContext(ctx, r.client)
	entity, err := client.OpenAIWebThread.Query().
		Where(
			openaiwebthread.LocalThreadIDEQ(localThreadID),
			openaiwebthread.UserIDEQ(userID),
			openaiwebthread.DeletedAtIsNil(),
		).
		WithGroup().
		WithAccount().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrOpenAIWebThreadNotFound, nil)
	}
	return openAIWebThreadEntityToService(entity), nil
}

func (r *openAIWebThreadRepository) ListActiveByUserID(ctx context.Context, userID int64) ([]service.OpenAIWebThread, error) {
	client := clientFromContext(ctx, r.client)
	entities, err := client.OpenAIWebThread.Query().
		Where(
			openaiwebthread.UserIDEQ(userID),
			openaiwebthread.DeletedAtIsNil(),
		).
		WithGroup().
		WithAccount().
		Order(dbent.Desc(openaiwebthread.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return openAIWebThreadEntitiesToService(entities), nil
}

func (r *openAIWebThreadRepository) Update(ctx context.Context, thread *service.OpenAIWebThread) error {
	if thread == nil {
		return service.ErrOpenAIWebThreadNilInput
	}

	client := clientFromContext(ctx, r.client)
	builder := client.OpenAIWebThread.UpdateOneID(thread.ID).
		SetUserID(thread.UserID).
		SetGroupID(thread.GroupID).
		SetAccountID(thread.AccountID).
		SetLocalThreadID(thread.LocalThreadID).
		SetPageSessionID(thread.PageSessionID).
		SetProvider(thread.Provider).
		SetTitle(thread.Title).
		SetRequestedModel(thread.RequestedModel).
		SetCapabilityMode(thread.CapabilityMode).
		SetHistoryMode(thread.HistoryMode).
		SetCachePolicy(thread.CachePolicy).
		SetSyncStatus(thread.SyncStatus).
		SetStatus(thread.Status).
		SetLastError(thread.LastError).
		SetNillableUpstreamConversationID(thread.UpstreamConversationID).
		SetNillableUpstreamSessionID(thread.UpstreamSessionID).
		SetNillableLastSyncedAt(thread.LastSyncedAt).
		SetNillableDeletedAt(thread.DeletedAt)

	if !thread.UpdatedAt.IsZero() {
		builder.SetUpdatedAt(thread.UpdatedAt)
	}

	updated, err := builder.Save(ctx)
	if err == nil {
		applyOpenAIWebThreadEntityToService(thread, updated)
		return nil
	}
	return translatePersistenceError(err, service.ErrOpenAIWebThreadNotFound, service.ErrOpenAIWebThreadAlreadyExists)
}

func (r *openAIWebThreadRepository) ArchiveByLocalThreadIDAndUserID(ctx context.Context, localThreadID string, userID int64, archivedAt time.Time) error {
	client := clientFromContext(ctx, r.client)
	updated, err := client.OpenAIWebThread.Update().
		Where(
			openaiwebthread.LocalThreadIDEQ(localThreadID),
			openaiwebthread.UserIDEQ(userID),
			openaiwebthread.DeletedAtIsNil(),
		).
		SetStatus(service.OpenAIWebThreadStatusArchived).
		SetDeletedAt(archivedAt).
		SetUpdatedAt(archivedAt).
		Save(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrOpenAIWebThreadNotFound, nil)
	}
	if updated == 0 {
		return service.ErrOpenAIWebThreadNotFound
	}
	return nil
}

func openAIWebThreadEntityToService(entity *dbent.OpenAIWebThread) *service.OpenAIWebThread {
	if entity == nil {
		return nil
	}
	thread := &service.OpenAIWebThread{}
	applyOpenAIWebThreadEntityToService(thread, entity)

	if edges := entity.Edges; edges.Group != nil {
		thread.Group = groupEntityToService(edges.Group)
	}
	if edges := entity.Edges; edges.Account != nil {
		thread.Account = accountEntityToService(edges.Account)
	}
	if edges := entity.Edges; edges.User != nil {
		thread.User = userEntityToService(edges.User)
	}
	return thread
}

func openAIWebThreadEntitiesToService(entities []*dbent.OpenAIWebThread) []service.OpenAIWebThread {
	out := make([]service.OpenAIWebThread, 0, len(entities))
	for _, entity := range entities {
		if thread := openAIWebThreadEntityToService(entity); thread != nil {
			out = append(out, *thread)
		}
	}
	return out
}

func applyOpenAIWebThreadEntityToService(dst *service.OpenAIWebThread, src *dbent.OpenAIWebThread) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.UserID = src.UserID
	dst.GroupID = src.GroupID
	dst.AccountID = src.AccountID
	dst.LocalThreadID = src.LocalThreadID
	dst.PageSessionID = src.PageSessionID
	dst.UpstreamConversationID = src.UpstreamConversationID
	dst.UpstreamSessionID = src.UpstreamSessionID
	dst.Provider = src.Provider
	dst.Title = src.Title
	dst.RequestedModel = src.RequestedModel
	dst.CapabilityMode = src.CapabilityMode
	dst.HistoryMode = src.HistoryMode
	dst.CachePolicy = src.CachePolicy
	dst.SyncStatus = src.SyncStatus
	dst.Status = src.Status
	dst.LastSyncedAt = src.LastSyncedAt
	dst.LastError = src.LastError
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
	dst.DeletedAt = src.DeletedAt
}
