package router

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mallvielfrass/templater/internal/models"
)

type testTaskStorage struct {
	tasks map[string]models.Task
}

func (t *testTaskStorage) AddDocHash(id string, docHash string) error {
	task, ok := t.tasks[id]
	if !ok {
		return fmt.Errorf("task not found")
	}
	task.DocHash = docHash
	task.UpdatedAt = time.Now()
	t.tasks[id] = task
	return nil
}

func (t *testTaskStorage) AddExelHash(id string, exelHash string) error {
	task, ok := t.tasks[id]
	if !ok {
		return fmt.Errorf("task not found")
	}
	task.ExelHash = exelHash
	task.UpdatedAt = time.Now()
	t.tasks[id] = task
	return nil
}

func (t *testTaskStorage) CreateTask(userID string) (models.Task, error) {
	task := models.Task{
		ID:        uuid.New().String(),
		UserID:    userID,
		Status:    models.StatusCreated,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if t.tasks == nil {
		t.tasks = map[string]models.Task{}
	}
	t.tasks[task.ID] = task
	return task, nil
}

func (t *testTaskStorage) GetTask(id string) (models.Task, error) {
	task, ok := t.tasks[id]
	if !ok {
		return models.Task{}, fmt.Errorf("task not found")
	}
	return task, nil
}

func (t *testTaskStorage) UpdateTaskStatus(id string, status models.TaskStatus) error {
	task, ok := t.tasks[id]
	if !ok {
		return fmt.Errorf("task not found")
	}
	task.Status = status
	task.UpdatedAt = time.Now()
	t.tasks[id] = task
	return nil
}

type testFile struct {
	data  []byte
	path  string
	size  int
	hash  string
	owner string
}

type testStorage struct {
	files     map[string]testFile
	docs      map[string][]byte
	generated map[string][]byte
	owners    map[string]string
	oo        map[string][3]string
	taskOO    map[string]string
}

func (t *testStorage) SaveDoc(docBytes []byte) (string, error) {
	h := sha256.New()
	h.Write(docBytes)
	hash := hex.EncodeToString(h.Sum(nil))
	if t.generated == nil {
		t.generated = map[string][]byte{}
	}
	t.generated[hash] = docBytes
	return hash, nil
}

func (t *testStorage) GetDocFileData(hash string) ([]byte, error) {
	data, ok := t.docs[hash]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return data, nil
}

func (t *testStorage) GetDocFileInfo(hash string) (models.FileInfo, error) {
	data, err := t.GetDocFileData(hash)
	if err != nil {
		return models.FileInfo{}, err
	}
	return models.FileInfo{FileName: "document.docx", Size: len(data)}, nil
}

func (t *testStorage) SaveDocFile(path string, data []byte) (string, error) {
	h := sha256.New()
	h.Write(data)
	hash := hex.EncodeToString(h.Sum(nil))
	if t.docs == nil {
		t.docs = map[string][]byte{}
	}
	t.docs[hash] = data
	return hash, nil
}

func (t *testStorage) GetExelFileInfo(hash string) (models.FileInfo, error) {
	f, ok := t.files[hash]
	if !ok {
		return models.FileInfo{}, fmt.Errorf("not found")
	}
	return models.FileInfo{FileName: f.path, Size: f.size}, nil
}

func (t *testStorage) GetExelFileData(hash string) ([]byte, error) {
	f, ok := t.files[hash]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return f.data, nil
}

func (t *testStorage) SaveExelFile(path string, data []byte) (string, error) {
	h := sha256.New()
	h.Write(data)
	hash := hex.EncodeToString(h.Sum(nil))
	t.files[hash] = testFile{
		data: data,
		path: path,
		size: len(data),
		hash: hash,
	}
	return hash, nil
}

func (t *testStorage) GetAnyDocData(hash string) ([]byte, error) {
	if data, ok := t.docs[hash]; ok {
		return data, nil
	}
	if data, ok := t.generated[hash]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("not found")
}

func (t *testStorage) BindOwner(hash, userID string) error {
	if t.owners == nil {
		t.owners = map[string]string{}
	}
	t.owners[hash] = userID
	return nil
}

func (t *testStorage) OwnerOf(hash string) (string, error) {
	owner, ok := t.owners[hash]
	if !ok {
		return "", fmt.Errorf("not found")
	}
	return owner, nil
}

func (t *testStorage) SaveOOMapping(key, userID, hash, taskID string) error {
	if t.oo == nil {
		t.oo = map[string][3]string{}
	}
	t.oo[key] = [3]string{userID, hash, taskID}
	return nil
}

func (t *testStorage) GetOOMapping(key string) (string, string, string, error) {
	v, ok := t.oo[key]
	if !ok {
		return "", "", "", fmt.Errorf("not found")
	}
	return v[0], v[1], v[2], nil
}

func (t *testStorage) DeleteOOMapping(key string) error {
	delete(t.oo, key)
	return nil
}

func (t *testStorage) SaveTaskOOKey(taskID, key string) error {
	if t.taskOO == nil {
		t.taskOO = map[string]string{}
	}
	t.taskOO[taskID] = key
	return nil
}

func (t *testStorage) GetTaskOOKey(taskID string) (string, error) {
	if t.taskOO == nil {
		return "", fmt.Errorf("not found")
	}
	v, ok := t.taskOO[taskID]
	if !ok {
		return "", fmt.Errorf("not found")
	}
	return v, nil
}

type testUserStorage struct {
	users map[string]bool
}

func (t *testUserStorage) CreateTempUser() string {
	id := uuid.New().String()
	t.users[id] = true
	return id
}

func (t *testUserStorage) IsUserExist(token string) bool {
	_, ok := t.users[token]
	return ok
}
