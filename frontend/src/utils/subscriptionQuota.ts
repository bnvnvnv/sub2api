import type { UserSubscription } from '@/types'

export type SubscriptionUsagePeriod = 'daily' | 'weekly' | 'monthly'

export function getSubscriptionBaseLimit(
  subscription: UserSubscription,
  period: SubscriptionUsagePeriod
): number | null {
  const group = subscription.group
  if (!group) return null
  const base =
    period === 'daily'
      ? group.daily_limit_usd
      : period === 'weekly'
        ? group.weekly_limit_usd
        : group.monthly_limit_usd
  return base && base > 0 ? base : null
}

export function getSubscriptionBonusUSD(
  subscription: UserSubscription,
  period: SubscriptionUsagePeriod
): number {
  const bonus =
    period === 'daily'
      ? subscription.daily_bonus_usd
      : period === 'weekly'
        ? subscription.weekly_bonus_usd
        : subscription.monthly_bonus_usd
  return Math.max(bonus || 0, 0)
}

export function getEffectiveSubscriptionLimit(
  subscription: UserSubscription,
  period: SubscriptionUsagePeriod
): number | null {
  const base = getSubscriptionBaseLimit(subscription, period)
  if (base == null) return null
  return base + getSubscriptionBonusUSD(subscription, period)
}

export function hasSubscriptionBonus(
  subscription: UserSubscription,
  period: SubscriptionUsagePeriod
): boolean {
  return getSubscriptionBonusUSD(subscription, period) > 0
}
