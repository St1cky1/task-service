-- Добавляем email и password_hash в таблицу user
ALTER TABLE "user" ADD COLUMN email VARCHAR(255) UNIQUE NOT NULL DEFAULT '';
ALTER TABLE "user" ADD COLUMN password_hash VARCHAR(255) DEFAULT '';
ALTER TABLE "user" ADD COLUMN is_active BOOLEAN DEFAULT true;
ALTER TABLE "user" ADD COLUMN last_login TIMESTAMP WITH TIME ZONE;

-- Индекс для быстрого поиска по email
CREATE INDEX idx_user_email ON "user"(email);