CREATE TABLE "task_audit" (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(100) NOT NULL DEFAULT 'task',
    entity_id INTEGER NOT NULL,
    old_values JSONB,
    new_values JSONB,
    changes JSONB,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_audit_user 
        FOREIGN KEY (user_id) 
        REFERENCES "user"(id)
        ON DELETE SET NULL
);

CREATE INDEX idx_task_audit_entity_type_entity_id ON task_audit(entity_type, entity_id);
CREATE INDEX idx_task_audit_entity_id ON task_audit(entity_id);
CREATE INDEX idx_task_audit_user_id ON task_audit(user_id);
CREATE INDEX idx_task_audit_action ON task_audit(action);
CREATE INDEX idx_task_audit_changed_at ON task_audit(changed_at);