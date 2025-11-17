-- Добавляем колонку для хранения пути аватарки
ALTER TABLE "user" ADD COLUMN avatar_url VARCHAR(500);

CREATE INDEX idx_user_avatar_url ON "user"(avatar_url);