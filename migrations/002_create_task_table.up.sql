CREATE TABLE IF NOT EXISTS "task" (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    owner_id INTEGER NOT NULL,
    
    CONSTRAINT fk_task_owner 
        FOREIGN KEY (owner_id) 
        REFERENCES "user"(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_task_created_at ON task(created_at);
CREATE INDEX IF NOT EXISTS idx_task_updated_at ON task(updated_at);
CREATE INDEX IF NOT EXISTS idx_task_status ON task(status);
CREATE INDEX IF NOT EXISTS idx_task_owner_id ON task(owner_id);
CREATE INDEX IF NOT EXISTS idx_task_title ON task(title);
CREATE INDEX IF NOT EXISTS idx_task_status_owner ON task(status, owner_id);