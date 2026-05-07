package service

import "strings"

const (
	SubscriptionQuotaPeriodDaily   = "daily"
	SubscriptionQuotaPeriodWeekly  = "weekly"
	SubscriptionQuotaPeriodMonthly = "monthly"
)

func (s *UserSubscription) EffectiveDailyLimitUSD(group *Group) *float64 {
	if group == nil || !group.HasDailyLimit() {
		return nil
	}
	return effectiveSubscriptionLimit(group.DailyLimitUSD, s.DailyBonusUSD)
}

func (s *UserSubscription) EffectiveWeeklyLimitUSD(group *Group) *float64 {
	if group == nil || !group.HasWeeklyLimit() {
		return nil
	}
	return effectiveSubscriptionLimit(group.WeeklyLimitUSD, s.WeeklyBonusUSD)
}

func (s *UserSubscription) EffectiveMonthlyLimitUSD(group *Group) *float64 {
	if group == nil || !group.HasMonthlyLimit() {
		return nil
	}
	return effectiveSubscriptionLimit(group.MonthlyLimitUSD, s.MonthlyBonusUSD)
}

func effectiveSubscriptionLimit(base *float64, bonus float64) *float64 {
	if base == nil || *base <= 0 {
		return nil
	}
	limit := *base
	if bonus > 0 {
		limit += bonus
	}
	return &limit
}

func NormalizeSubscriptionQuotaPeriod(period string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(period)) {
	case SubscriptionQuotaPeriodDaily:
		return SubscriptionQuotaPeriodDaily, nil
	case SubscriptionQuotaPeriodWeekly:
		return SubscriptionQuotaPeriodWeekly, nil
	case SubscriptionQuotaPeriodMonthly:
		return SubscriptionQuotaPeriodMonthly, nil
	default:
		return "", ErrInvalidSubscriptionQuotaPeriod
	}
}

func SubscriptionQuotaPeriodHasLimit(group *Group, period string) bool {
	if group == nil {
		return false
	}
	switch period {
	case SubscriptionQuotaPeriodDaily:
		return group.HasDailyLimit()
	case SubscriptionQuotaPeriodWeekly:
		return group.HasWeeklyLimit()
	case SubscriptionQuotaPeriodMonthly:
		return group.HasMonthlyLimit()
	default:
		return false
	}
}
