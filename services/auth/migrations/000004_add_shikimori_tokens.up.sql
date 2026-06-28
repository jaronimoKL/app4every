ALTER TABLE users
ADD COLUMN IF NOT EXISTS shikimori_access_token TEXT,
ADD COLUMN IF NOT EXISTS shikimori_refresh_token TEXT,
ADD COLUMN IF NOT EXISTS shikimori_user_id BIGINT;
