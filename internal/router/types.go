package router

import (
	"time"

	"github.com/mallvielfrass/templater/internal/models"
)

type fileStorage interface {
	SaveExelFile(path string, data []byte) (hash string, err error)
	GetExelFileInfo(hash string) (fileInfo models.FileInfo, err error)
	GetExelFileData(hash string) (data []byte, err error)
	SaveDocFile(path string, data []byte) (hash string, err error)
	GetDocFileInfo(hash string) (fileInfo models.FileInfo, err error)
	GetDocFileData(hash string) (data []byte, err error)
}
type userStorage interface {
	CreateTempUser() (token string)
	IsUserExist(token string) bool
}
type status string

const (
	StatusCreated    status = "created"
	WaitAllDocuments status = "wait_all_documents"
	ReadyForConvert  status = "ready_for_convert"
	StatusError      status = "error"
	StatusCompleted  status = "completed"
)

type Task struct {
	ID        string
	UserID    string
	Status    status
	CreatedAt time.Time
	UpdatedAt time.Time
	ExelHash  string
	DocHash   string
}
type taskStorage interface {
	CreateTask(userID string) (task Task, err error)
	UpdateTaskStatus(id string, status status) (err error)
	AddExelHash(id string, exelHash string) (err error)
	AddDocHash(id string, docHash string) (err error)
	GetTask(id string) (task Task, err error)
}
