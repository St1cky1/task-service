package entity

import (
	"time"
)

type ActionType string

const (
	ActionCreate ActionType = "Create"
	ActionRead   ActionType = "Read"
	ActionUpdate ActionType = "Update"
	ActionDelete ActionType = "Delete"
)

type TaskAudit struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	Action     ActionType `json:"action"`
	EntityType string     `json:"entity_type"`
	EntityID   int        `json:"entity_id"`
	OldValues  *string    `json:"old_values"`
	NewValues  *string    `json:"new_values"`
	Changes    *string    `json:"changes"`
	ChangesAt  time.Time  `json:"changed_at"`
}

type AuditMessage struct {
	UserID    int            `json:"user_id"`
	Action    ActionType     `json:"action"`
	EntityID  int            `json:"entity_id"`
	OldValues map[string]any `json:"old_values"`
	NewValues map[string]any `json:"new_values"`
	Changes   map[string]any `json:"changes"`
	Timestamp time.Time      `json:"timestamp"`
}
