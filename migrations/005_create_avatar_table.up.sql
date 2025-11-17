CREATE TABLE IF NOT EXISTS "avatar" (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE,
    file_path VARCHAR(500) NOT NULL,
    file_size INTEGER NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_avatar_user 
        FOREIGN KEY (user_id) 
        REFERENCES "user"(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_avatar_user_id ON "avatar"(user_id);
CREATE INDEX idx_avatar_created_at ON "avatar"(created_at);