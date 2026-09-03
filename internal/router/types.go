package router

import (
	"github.com/mallvielfrass/templater/internal/models"
)

type fileStorage interface {
	SaveExelFile(path string, data []byte) (hash string, err error)
	GetExelFileInfo(hash string) (fileInfo models.FileInfo, err error)
	GetExelFileData(hash string) (data []byte, err error)
	SaveDocFile(path string, data []byte) (hash string, err error)
	GetDocFileInfo(hash string) (fileInfo models.FileInfo, err error)
	GetDocFileData(hash string) (data []byte, err error)
	SaveDoc(docBytes []byte) (hash string, err error)
	GetAnyDocData(hash string) (data []byte, err error)
	BindOwner(hash, userID string) error
	OwnerOf(hash string) (string, error)
	SaveOOMapping(key, userID, hash, taskID string) error
	GetOOMapping(key string) (userID, hash, taskID string, err error)
	DeleteOOMapping(key string) error
	SaveTaskOOKey(taskID, key string) error
	GetTaskOOKey(taskID string) (string, error)
}

type userStorage interface {
	CreateTempUser() (token string)
	IsUserExist(token string) bool
}

type taskStorage interface {
	CreateTask(userID string) (task models.Task, err error)
	UpdateTaskStatus(id string, status models.TaskStatus) (err error)
	AddExelHash(id string, exelHash string) (err error)
	AddDocHash(id string, docHash string) (err error)
	GetTask(id string) (task models.Task, err error)
}
