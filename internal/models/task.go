package models

import "time"

type TaskStatus string

const (
	StatusCreated    TaskStatus = "created"
	WaitAllDocuments TaskStatus = "wait_all_documents"
	ReadyForConvert  TaskStatus = "ready_for_convert"
	StatusError      TaskStatus = "error"
	StatusCompleted  TaskStatus = "completed"
)

type Task struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Status    TaskStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ExelHash  string     `json:"exel_hash"`
	DocHash   string     `json:"doc_hash"`
}
