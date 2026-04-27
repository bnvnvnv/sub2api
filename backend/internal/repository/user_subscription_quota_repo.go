package repository

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *userSubscriptionRepository) AddQuotaBonus(ctx context.Context, id int64, period string, amountUSD float64) error {
	var field string
	switch period {
	case service.SubscriptionQuotaPeriodDaily:
		field = "daily_bonus_usd"
	case service.SubscriptionQuotaPeriodWeekly:
		field = "weekly_bonus_usd"
	case service.SubscriptionQuotaPeriodMonthly:
		field = "monthly_bonus_usd"
	default:
		return service.ErrInvalidSubscriptionQuotaPeriod
	}

	query := "UPDATE user_subscriptions SET " + field + " = " + field + " + $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL"
	client := clientFromContext(ctx, r.client)
	result, err := client.ExecContext(ctx, query, amountUSD, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrSubscriptionNotFound
	}
	return nil
}
