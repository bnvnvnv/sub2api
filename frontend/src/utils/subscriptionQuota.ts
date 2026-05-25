import type { UserSubscription } from '@/types'

export type SubscriptionUsagePeriod = 'daily' | 'weekly' | 'monthly'

const ONE_DAY_MS = 24 * 60 * 60 * 1000

export interface RemainingDurationParts {
  days: number
  hours: number
  minutes: number
}

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

export function isOneTimeDailyQuota(
  subscription: Pick<UserSubscription, 'starts_at' | 'expires_at'>
): boolean {
  if (!subscription.starts_at || !subscription.expires_at) return false

  const startsAt = new Date(subscription.starts_at).getTime()
  const expiresAt = new Date(subscription.expires_at).getTime()

  if (!Number.isFinite(startsAt) || !Number.isFinite(expiresAt)) return false

  return expiresAt <= startsAt + ONE_DAY_MS
}

export function getRemainingDurationParts(
  targetAt: Date | string,
  now: Date = new Date()
): RemainingDurationParts | null {
  const targetTime = targetAt instanceof Date ? targetAt.getTime() : new Date(targetAt).getTime()
  const nowTime = now.getTime()

  if (!Number.isFinite(targetTime) || !Number.isFinite(nowTime)) return null

  const diffMs = targetTime - nowTime
  if (diffMs <= 0) return null

  const totalMinutes = Math.floor(diffMs / (1000 * 60))
  const days = Math.floor(totalMinutes / (24 * 60))
  const hours = Math.floor((totalMinutes % (24 * 60)) / 60)
  const minutes = totalMinutes % 60

  return { days, hours, minutes }
}
