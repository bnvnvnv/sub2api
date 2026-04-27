-- Add subscription quota add-on redeem code support.
-- Redeem codes keep the target period; active add-ons live on the current
-- subscription usage window and are cleared when that window resets.

ALTER TABLE redeem_codes
    ADD COLUMN IF NOT EXISTS quota_period varchar(20) NOT NULL DEFAULT '';

ALTER TABLE user_subscriptions
    ADD COLUMN IF NOT EXISTS daily_bonus_usd decimal(20,10) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS weekly_bonus_usd decimal(20,10) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS monthly_bonus_usd decimal(20,10) NOT NULL DEFAULT 0;
