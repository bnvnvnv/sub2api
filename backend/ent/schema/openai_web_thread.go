package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// OpenAIWebThread stores user-owned page thread metadata for ChatGPT Web sessions.
type OpenAIWebThread struct {
	ent.Schema
}

func (OpenAIWebThread) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "openai_web_threads"},
	}
}

func (OpenAIWebThread) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.Int64("group_id"),
		field.Int64("account_id"),

		field.String("local_thread_id").
			MaxLen(64).
			NotEmpty().
			Unique(),
		field.String("page_session_id").
			MaxLen(64).
			NotEmpty().
			Unique(),
		field.String("upstream_conversation_id").
			MaxLen(128).
			Optional().
			Nillable(),
		field.String("upstream_session_id").
			MaxLen(128).
			Optional().
			Nillable(),

		field.String("provider").
			MaxLen(32).
			Default("openai_web"),
		field.String("title").
			MaxLen(255).
			Default(""),
		field.String("requested_model").
			MaxLen(100).
			Default(""),
		field.String("capability_mode").
			MaxLen(32).
			Default("web_chat"),
		field.String("history_mode").
			MaxLen(32).
			Default("upstream_only"),
		field.String("cache_policy").
			MaxLen(32).
			Default("local_only"),
		field.String("sync_status").
			MaxLen(32).
			Default("pending"),
		field.String("status").
			MaxLen(32).
			Default("active"),

		field.Time("last_synced_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("last_error").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default(""),

		field.Time("created_at").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("deleted_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (OpenAIWebThread) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("openai_web_threads").
			Field("user_id").
			Required().
			Unique(),
		edge.From("group", Group.Type).
			Ref("openai_web_threads").
			Field("group_id").
			Required().
			Unique(),
		edge.From("account", Account.Type).
			Ref("openai_web_threads").
			Field("account_id").
			Required().
			Unique(),
	}
}

func (OpenAIWebThread) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("group_id"),
		index.Fields("account_id"),
		index.Fields("local_thread_id"),
		index.Fields("page_session_id"),
		index.Fields("upstream_conversation_id"),
		index.Fields("deleted_at"),
		index.Fields("user_id", "updated_at"),
	}
}
