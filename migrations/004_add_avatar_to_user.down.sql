DROP INDEX IF EXISTS idx_user_avatar_url;
ALTER TABLE "user" DROP COLUMN IF EXISTS avatar_url;