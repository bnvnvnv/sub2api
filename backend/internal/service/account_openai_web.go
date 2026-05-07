package service

import "strings"

func (a *Account) GetOpenAIPlanType() string {
	if a == nil || !a.IsOpenAIOAuth() {
		return ""
	}
	return strings.TrimSpace(strings.ToLower(a.GetCredential("plan_type")))
}

func (a *Account) SupportsOpenAIWebChat() bool {
	if a == nil || !a.IsOpenAIOAuth() {
		return false
	}
	if strings.TrimSpace(a.GetOpenAIAccessToken()) == "" {
		return false
	}
	return strings.TrimSpace(a.GetChatGPTAccountID()) != ""
}

func (a *Account) SupportsOpenAIWebModel(requestedModel string) bool {
	if !a.SupportsOpenAIWebChat() {
		return false
	}
	if !isOpenAIWebProModel(requestedModel) {
		return true
	}
	planType := a.GetOpenAIPlanType()
	return strings.Contains(planType, "pro")
}
