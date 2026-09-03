package filestorage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mallvielfrass/templater/internal/models"
)

func taskKey(id string) string {
	return "task_" + id
}

func (f *fileStorage) saveTask(task models.Task) error {
	task.UpdatedAt = time.Now()
	b, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return f.Set(taskKey(task.ID), b)
}

func (f *fileStorage) CreateTask(userID string) (models.Task, error) {
	if userID == "" {
		return models.Task{}, fmt.Errorf("user id is required")
	}
	now := time.Now()
	task := models.Task{
		ID:        uuid.New().String(),
		UserID:    userID,
		Status:    models.StatusCreated,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := f.saveTask(task); err != nil {
		return models.Task{}, err
	}
	return task, nil
}

func (f *fileStorage) GetTask(id string) (models.Task, error) {
	b, err := f.Get(taskKey(id))
	if err != nil {
		return models.Task{}, err
	}
	var task models.Task
	if err := json.Unmarshal(b, &task); err != nil {
		return models.Task{}, err
	}
	return task, nil
}

func (f *fileStorage) UpdateTaskStatus(id string, status models.TaskStatus) error {
	task, err := f.GetTask(id)
	if err != nil {
		return err
	}
	task.Status = status
	return f.saveTask(task)
}

func (f *fileStorage) AddExelHash(id string, exelHash string) error {
	task, err := f.GetTask(id)
	if err != nil {
		return err
	}
	task.ExelHash = exelHash
	return f.saveTask(task)
}

func (f *fileStorage) AddDocHash(id string, docHash string) error {
	task, err := f.GetTask(id)
	if err != nil {
		return err
	}
	task.DocHash = docHash
	return f.saveTask(task)
}
