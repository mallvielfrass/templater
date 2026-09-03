package router

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/mallvielfrass/templater/internal/models"
)

type testTaskStorage struct {
	tasks map[string]testTask
}

// AddDocHash implements taskStorage.
func (t *testTaskStorage) AddDocHash(id string, docHash string) (err error) {
	panic("unimplemented")
}

// AddExelHash implements taskStorage.
func (t *testTaskStorage) AddExelHash(id string, exelHash string) (err error) {
	panic("unimplemented")
}

// CreateTask implements taskStorage.
func (t *testTaskStorage) CreateTask(userID string) (task Task, err error) {
	panic("unimplemented")
}

// GetTask implements taskStorage.
func (t *testTaskStorage) GetTask(id string) (task Task, err error) {
	panic("unimplemented")
}

// UpdateTaskStatus implements taskStorage.
func (t *testTaskStorage) UpdateTaskStatus(id string, status status) (err error) {
	panic("unimplemented")
}

type testTask struct {
	id        string
	status    status
	createdAt time.Time
	updatedAt time.Time
	exelHash  string
	docHash   string
}
type testFile struct {
	data []byte
	path string
	size int
	hash string
}
type testStorage struct {
	files map[string]testFile
}

// SaveDoc implements fileStorage.
func (t *testStorage) SaveDoc(docBytes []byte) (hash string, err error) {
	panic("unimplemented")
}

// GetDocFileData implements fileStorage.
func (t *testStorage) GetDocFileData(hash string) (data []byte, err error) {
	panic("unimplemented")
}

// GetDocFileInfo implements fileStorage.
func (t *testStorage) GetDocFileInfo(hash string) (fileInfo models.FileInfo, err error) {
	panic("unimplemented")
}

// SaveDocFile implements fileStorage.
func (t *testStorage) SaveDocFile(path string, data []byte) (hash string, err error) {
	panic("unimplemented")
}

// GetFileInfo implements fileStorage.
func (t *testStorage) GetExelFileInfo(hash string) (fileInfo models.FileInfo, err error) {
	panic("unimplemented")
}

type testUserStorage struct {
	users map[string]bool
}

// CreateTempUser implements userStorage.
func (t *testUserStorage) CreateTempUser() (token string) {
	id := uuid.New().String()
	t.users[id] = true
	return id
}

// IsUserExist implements userStorage.
func (t *testUserStorage) IsUserExist(token string) bool {
	_, ok := t.users[token]
	return ok
}

// GetFileData implements storage.
func (t *testStorage) GetExelFileData(hash string) (data []byte, err error) {
	panic("unimplemented")
}

// SaveFile implements storage.
func (t *testStorage) SaveExelFile(path string, data []byte) (string, error) {
	h := sha256.New()
	h.Write(data)
	bs := h.Sum(nil)
	hash := hex.EncodeToString(bs)
	t.files[hash] = testFile{
		data: data,
		path: path,
		size: len(data),
		hash: hash,
	}
	return hash, nil
}
