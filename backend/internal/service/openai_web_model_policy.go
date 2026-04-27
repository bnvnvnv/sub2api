package service

import "strings"

const (
	openAIWebDefaultStandardModel = "gpt-5.4-mini"
	openAIWebDefaultPaidModel     = "gpt-5.4"
)

var openAIWebProModelMarkers = []string{
	"5.4-pro",
	"5-4-pro",
	"gpt-5.4-pro",
	"5.5-pro",
	"5-5-pro",
	"gpt-5.5-pro",
}

func isOpenAIWebProModel(requestedModel string) bool {
	model := strings.TrimSpace(strings.ToLower(requestedModel))
	if model == "" {
		return false
	}
	for _, marker := range openAIWebProModelMarkers {
		if strings.Contains(model, marker) {
			return true
		}
	}
	return false
}

func defaultOpenAIWebRequestedModel(hasProAccounts bool) string {
	if hasProAccounts {
		return openAIWebDefaultPaidModel
	}
	return openAIWebDefaultStandardModel
}

func defaultOpenAIWebRequestedModelForAccount(account *Account) string {
	if account != nil && strings.Contains(account.GetOpenAIPlanType(), "pro") {
		return openAIWebDefaultPaidModel
	}
	return openAIWebDefaultStandardModel
}
