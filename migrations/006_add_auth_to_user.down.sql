-- Откатываем добавление email и password_hash
ALTER TABLE "user" DROP COLUMN last_login;
ALTER TABLE "user" DROP COLUMN is_active;
ALTER TABLE "user" DROP COLUMN password_hash;
ALTER TABLE "user" DROP COLUMN email;

-- Удаляем индекс
DROP INDEX IF EXISTS idx_user_email;