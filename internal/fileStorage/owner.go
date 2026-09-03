package filestorage

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

func ownerKey(hash string) string {
	return "owner_" + hash
}

func (f *fileStorage) BindOwner(hash, userID string) error {
	if hash == "" || userID == "" {
		return fmt.Errorf("hash and user id are required")
	}
	return f.Set(ownerKey(hash), []byte(userID))
}

func (f *fileStorage) OwnerOf(hash string) (string, error) {
	data, err := f.Get(ownerKey(hash))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type ooMapping struct {
	UserID string `json:"user_id"`
	Hash   string `json:"hash"`
	TaskID string `json:"task_id"`
}

func ooMapKey(key string) string {
	return "oo_map_" + key
}

func taskOOKey(taskID string) string {
	return "task_oo_key_" + taskID
}

func (f *fileStorage) SaveTaskOOKey(taskID, key string) error {
	if taskID == "" || key == "" {
		return nil
	}
	return f.Set(taskOOKey(taskID), []byte(key))
}

func (f *fileStorage) GetTaskOOKey(taskID string) (string, error) {
	data, err := f.Get(taskOOKey(taskID))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (f *fileStorage) SaveOOMapping(key, userID, hash, taskID string) error {
	b, err := json.Marshal(ooMapping{UserID: userID, Hash: hash, TaskID: taskID})
	if err != nil {
		return err
	}
	if err := f.Set(ooMapKey(key), b); err != nil {
		return err
	}
	if taskID != "" && key != "" {
		_ = f.SaveTaskOOKey(taskID, key)
	}
	return nil
}

func (f *fileStorage) GetOOMapping(key string) (userID, hash, taskID string, err error) {
	b, err := f.Get(ooMapKey(key))
	if err != nil {
		return "", "", "", err
	}
	var m ooMapping
	if err := json.Unmarshal(b, &m); err != nil {
		return "", "", "", err
	}
	return m.UserID, m.Hash, m.TaskID, nil
}

func (f *fileStorage) DeleteOOMapping(key string) error {
	return f.bg.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(ooMapKey(key)))
	})
}

func (f *fileStorage) GetAnyDocData(hash string) ([]byte, error) {
	data, err := f.GetDocFileData(hash)
	if err == nil {
		return data, nil
	}
	return f.Get("generated_doc_" + hash)
}
