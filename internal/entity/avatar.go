package entity

import "time"

type Avatar struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	FilePath    string    `json:"file_path"`
	FileSize    int       `json:"file_size"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UploadAvatarRequest struct {
	UserID      int
	Data        []byte
	ContentType string
}

type DownloadAvatarRequest struct {
	UserID int
}
