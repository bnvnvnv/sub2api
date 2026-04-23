package dto

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAccountFromServiceShallow_EnrichesOpenAIOAuthCredentialsFromIDToken(t *testing.T) {
	t.Parallel()

	account := &service.Account{
		ID:       1,
		Name:     "codex-old",
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeOAuth,
		Credentials: map[string]any{
			"id_token": buildTestOpenAIIDToken(t, map[string]any{
				"email": "user@example.com",
				"https://api.openai.com/auth": map[string]any{
					"chatgpt_account_id": "acct_123",
					"chatgpt_user_id":    "user_123",
					"chatgpt_plan_type":  "plus",
					"organizations": []map[string]any{
						{
							"id":         "org_default",
							"is_default": true,
						},
					},
				},
			}),
		},
	}

	dtoAccount := AccountFromServiceShallow(account)
	require.NotNil(t, dtoAccount)
	require.Equal(t, "plus", dtoAccount.Credentials["plan_type"])
	require.Equal(t, "user@example.com", dtoAccount.Credentials["email"])
	require.Equal(t, "acct_123", dtoAccount.Credentials["chatgpt_account_id"])
	require.Equal(t, "user_123", dtoAccount.Credentials["chatgpt_user_id"])
	require.Equal(t, "org_default", dtoAccount.Credentials["organization_id"])

	require.NotContains(t, account.Credentials, "plan_type")
	require.NotContains(t, account.Credentials, "organization_id")
}

func TestAccountFromServiceShallow_DoesNotOverrideExistingOpenAIOAuthCredentials(t *testing.T) {
	t.Parallel()

	account := &service.Account{
		ID:       2,
		Name:     "codex-existing",
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeOAuth,
		Credentials: map[string]any{
			"id_token":        buildTestOpenAIIDToken(t, map[string]any{"https://api.openai.com/auth": map[string]any{"chatgpt_plan_type": "free"}}),
			"plan_type":       "team",
			"email":           "keep@example.com",
			"chatgpt_user_id": "existing_user",
		},
	}

	dtoAccount := AccountFromServiceShallow(account)
	require.NotNil(t, dtoAccount)
	require.Equal(t, "team", dtoAccount.Credentials["plan_type"])
	require.Equal(t, "keep@example.com", dtoAccount.Credentials["email"])
	require.Equal(t, "existing_user", dtoAccount.Credentials["chatgpt_user_id"])
}

func buildTestOpenAIIDToken(t *testing.T, payload map[string]any) string {
	t.Helper()

	headerBytes, err := json.Marshal(map[string]any{
		"alg": "none",
		"typ": "JWT",
	})
	require.NoError(t, err)

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return base64.RawURLEncoding.EncodeToString(headerBytes) +
		"." +
		base64.RawURLEncoding.EncodeToString(payloadBytes) +
		".signature"
}
