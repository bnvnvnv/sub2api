package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountSupportsOpenAIWebModelRecognizesFiveFivePro(t *testing.T) {
	proAccount := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":       "token",
			"chatgpt_account_id": "acct-pro",
			"plan_type":          "pro",
		},
	}
	plusAccount := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":       "token",
			"chatgpt_account_id": "acct-plus",
			"plan_type":          "plus",
		},
	}

	require.True(t, proAccount.SupportsOpenAIWebModel("gpt-5.5-pro"))
	require.True(t, proAccount.SupportsOpenAIWebModel("gpt-5-5-pro"))
	require.False(t, plusAccount.SupportsOpenAIWebModel("gpt-5.5-pro"))
}
