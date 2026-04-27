package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

const openAIWebThreadSelectColumns = `
	id,
	user_id,
	group_id,
	account_id,
	local_thread_id,
	page_session_id,
	upstream_conversation_id,
	upstream_session_id,
	provider,
	title,
	requested_model,
	capability_mode,
	history_mode,
	cache_policy,
	sync_status,
	status,
	last_synced_at,
	last_error,
	created_at,
	updated_at,
	deleted_at`

type openAIWebThreadRepository struct {
	client *dbent.Client
	db     *sql.DB
}

func NewOpenAIWebThreadRepository(client *dbent.Client, db *sql.DB) service.OpenAIWebThreadRepository {
	return &openAIWebThreadRepository{client: client, db: db}
}

func (r *openAIWebThreadRepository) Create(ctx context.Context, thread *service.OpenAIWebThread) error {
	if thread == nil {
		return service.ErrOpenAIWebThreadNilInput
	}
	now := time.Now()
	if thread.CreatedAt.IsZero() {
		thread.CreatedAt = now
	}
	if thread.UpdatedAt.IsZero() {
		thread.UpdatedAt = now
	}

	const query = `
		INSERT INTO openai_web_threads (
			user_id,
			group_id,
			account_id,
			local_thread_id,
			page_session_id,
			upstream_conversation_id,
			upstream_session_id,
			provider,
			title,
			requested_model,
			capability_mode,
			history_mode,
			cache_policy,
			sync_status,
			status,
			last_synced_at,
			last_error,
			created_at,
			updated_at,
			deleted_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)
		RETURNING ` + openAIWebThreadSelectColumns

	created, err := r.scanThread(ctx, query,
		thread.UserID,
		thread.GroupID,
		thread.AccountID,
		thread.LocalThreadID,
		thread.PageSessionID,
		nullableStringArg(thread.UpstreamConversationID),
		nullableStringArg(thread.UpstreamSessionID),
		thread.Provider,
		thread.Title,
		thread.RequestedModel,
		thread.CapabilityMode,
		thread.HistoryMode,
		thread.CachePolicy,
		thread.SyncStatus,
		thread.Status,
		nullableTimeArg(thread.LastSyncedAt),
		thread.LastError,
		thread.CreatedAt,
		thread.UpdatedAt,
		nullableTimeArg(thread.DeletedAt),
	)
	if err != nil {
		return translatePersistenceError(err, nil, service.ErrOpenAIWebThreadAlreadyExists)
	}
	*thread = *created
	return nil
}

func (r *openAIWebThreadRepository) GetByLocalThreadIDAndUserID(ctx context.Context, localThreadID string, userID int64) (*service.OpenAIWebThread, error) {
	const query = `SELECT ` + openAIWebThreadSelectColumns + `
		FROM openai_web_threads
		WHERE local_thread_id = $1 AND user_id = $2 AND deleted_at IS NULL
		LIMIT 1`
	thread, err := r.scanThread(ctx, query, localThreadID, userID)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrOpenAIWebThreadNotFound, nil)
	}
	return thread, nil
}

func (r *openAIWebThreadRepository) ListActiveByUserID(ctx context.Context, userID int64) ([]service.OpenAIWebThread, error) {
	const query = `SELECT ` + openAIWebThreadSelectColumns + `
		FROM openai_web_threads
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY updated_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []service.OpenAIWebThread{}
	for rows.Next() {
		thread, err := scanOpenAIWebThreadRow(rows)
		if err != nil {
			return nil, err
		}
		if err := r.attachOpenAIWebThreadEdges(ctx, thread); err != nil {
			return nil, err
		}
		out = append(out, *thread)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *openAIWebThreadRepository) Update(ctx context.Context, thread *service.OpenAIWebThread) error {
	if thread == nil {
		return service.ErrOpenAIWebThreadNilInput
	}
	if thread.UpdatedAt.IsZero() {
		thread.UpdatedAt = time.Now()
	}

	const query = `
		UPDATE openai_web_threads SET
			user_id = $2,
			group_id = $3,
			account_id = $4,
			local_thread_id = $5,
			page_session_id = $6,
			upstream_conversation_id = $7,
			upstream_session_id = $8,
			provider = $9,
			title = $10,
			requested_model = $11,
			capability_mode = $12,
			history_mode = $13,
			cache_policy = $14,
			sync_status = $15,
			status = $16,
			last_synced_at = $17,
			last_error = $18,
			updated_at = $19,
			deleted_at = $20
		WHERE id = $1
		RETURNING ` + openAIWebThreadSelectColumns

	updated, err := r.scanThread(ctx, query,
		thread.ID,
		thread.UserID,
		thread.GroupID,
		thread.AccountID,
		thread.LocalThreadID,
		thread.PageSessionID,
		nullableStringArg(thread.UpstreamConversationID),
		nullableStringArg(thread.UpstreamSessionID),
		thread.Provider,
		thread.Title,
		thread.RequestedModel,
		thread.CapabilityMode,
		thread.HistoryMode,
		thread.CachePolicy,
		thread.SyncStatus,
		thread.Status,
		nullableTimeArg(thread.LastSyncedAt),
		thread.LastError,
		thread.UpdatedAt,
		nullableTimeArg(thread.DeletedAt),
	)
	if err != nil {
		return translatePersistenceError(err, service.ErrOpenAIWebThreadNotFound, service.ErrOpenAIWebThreadAlreadyExists)
	}
	*thread = *updated
	return nil
}

func (r *openAIWebThreadRepository) ArchiveByLocalThreadIDAndUserID(ctx context.Context, localThreadID string, userID int64, archivedAt time.Time) error {
	const query = `
		UPDATE openai_web_threads
		SET status = $3, deleted_at = $4, updated_at = $4
		WHERE local_thread_id = $1 AND user_id = $2 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, localThreadID, userID, service.OpenAIWebThreadStatusArchived, archivedAt)
	if err != nil {
		return translatePersistenceError(err, service.ErrOpenAIWebThreadNotFound, nil)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrOpenAIWebThreadNotFound
	}
	return nil
}

func (r *openAIWebThreadRepository) scanThread(ctx context.Context, query string, args ...any) (*service.OpenAIWebThread, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, sql.ErrNoRows
	}
	thread, err := scanOpenAIWebThreadRow(rows)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, errors.New("openai web thread query returned multiple rows")
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.attachOpenAIWebThreadEdges(ctx, thread); err != nil {
		return nil, err
	}
	return thread, nil
}

type openAIWebThreadScanner interface {
	Scan(dest ...any) error
}

func scanOpenAIWebThreadRow(row openAIWebThreadScanner) (*service.OpenAIWebThread, error) {
	var thread service.OpenAIWebThread
	var upstreamConversationID sql.NullString
	var upstreamSessionID sql.NullString
	var lastSyncedAt sql.NullTime
	var deletedAt sql.NullTime

	if err := row.Scan(
		&thread.ID,
		&thread.UserID,
		&thread.GroupID,
		&thread.AccountID,
		&thread.LocalThreadID,
		&thread.PageSessionID,
		&upstreamConversationID,
		&upstreamSessionID,
		&thread.Provider,
		&thread.Title,
		&thread.RequestedModel,
		&thread.CapabilityMode,
		&thread.HistoryMode,
		&thread.CachePolicy,
		&thread.SyncStatus,
		&thread.Status,
		&lastSyncedAt,
		&thread.LastError,
		&thread.CreatedAt,
		&thread.UpdatedAt,
		&deletedAt,
	); err != nil {
		return nil, err
	}

	thread.UpstreamConversationID = nullStringPtr(upstreamConversationID)
	thread.UpstreamSessionID = nullStringPtr(upstreamSessionID)
	thread.LastSyncedAt = nullTimePtr(lastSyncedAt)
	thread.DeletedAt = nullTimePtr(deletedAt)
	return &thread, nil
}

func (r *openAIWebThreadRepository) attachOpenAIWebThreadEdges(ctx context.Context, thread *service.OpenAIWebThread) error {
	if thread == nil || r.client == nil {
		return nil
	}
	client := clientFromContext(ctx, r.client)

	group, err := client.Group.Get(ctx, thread.GroupID)
	if err != nil {
		return translatePersistenceError(err, service.ErrGroupNotFound, nil)
	}
	thread.Group = groupEntityToService(group)

	account, err := client.Account.Get(ctx, thread.AccountID)
	if err != nil {
		return translatePersistenceError(err, service.ErrAccountNotFound, nil)
	}
	thread.Account = accountEntityToService(account)
	return nil
}

func nullableStringArg(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableTimeArg(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
