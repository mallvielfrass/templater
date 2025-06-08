package filestorage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	exelreader "github.com/mallvielfrass/templater/internal/exelReader"
	"github.com/mallvielfrass/templater/internal/models"
)

type fileStorage struct {
	bg *badger.DB
}

func (b *fileStorage) Set(key string, value []byte) error {
	return b.bg.Update(func(txn *badger.Txn) error {
		//	fmt.Printf("Set key: %s | value: %+v\n", key, parseBytesToBlock(value))
		return txn.Set([]byte(key), value)
	})
}
func (b *fileStorage) Get(key string) ([]byte, error) {
	var value []byte
	err := b.bg.View(func(txn *badger.Txn) error {
		//	fmt.Printf("Get key: %s\n", key)
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		//fmt.Printf("Value: %+v\n", item)
		value, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		//	fmt.Printf("Value: %+v\n", value)
		return nil
	})
	return value, err
}

func NewStorage(BadgerDBPath string) (*fileStorage, error) {
	bgOptions := badger.DefaultOptions(BadgerDBPath)
	//set log level
	bgOptions.Logger = nil
	badgerDB, err := badger.Open(bgOptions)
	if err != nil {
		return &fileStorage{}, err
	}
	return &fileStorage{
		bg: badgerDB,
	}, nil
}
func (f *fileStorage) SaveExelFile(path string, data []byte) (string, error) {
	file, err := exelreader.ReadBuffer(path, data)
	if err != nil {
		return "", err
	}
	info, err := file.FileInfo()
	if err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write(data)
	bs := h.Sum(nil)
	hash := hex.EncodeToString(bs)

	// Сохраняем info в BadgerDB с ключом info_hash
	infoKey := "info_" + hash
	infoBytes, err := json.Marshal(info)
	if err != nil {
		return "", err
	}
	err = f.Set(infoKey, infoBytes)
	if err != nil {
		return "", err
	}

	// Сохраняем data в BadgerDB с ключом data_hash
	dataKey := "data_" + hash
	err = f.Set(dataKey, data)
	if err != nil {
		return "", err
	}

	return hash, nil
}
func (f *fileStorage) GetExelFileInfo(hash string) (fileInfo models.FileInfo, err error) {
	// Формируем ключ для получения информации о файле
	infoKey := "info_" + hash

	// Получаем данные из BadgerDB
	infoBytes, err := f.Get(infoKey)
	if err != nil {
		return models.FileInfo{}, err
	}

	// Десериализуем JSON в структуру exelreader.FileInfo
	var exelInfo models.FileInfo
	err = json.Unmarshal(infoBytes, &exelInfo)
	if err != nil {
		return models.FileInfo{}, err
	}

	// Получаем размер файла из сохраненных данных
	dataKey := "data_" + hash
	dataBytes, err := f.Get(dataKey)
	if err != nil {
		return models.FileInfo{}, err
	}

	fileInfo = models.FileInfo{
		Sheets:   exelInfo.Sheets,
		FileName: exelInfo.FileName,
		Size:     len(dataBytes),
	}

	return fileInfo, nil
}
func (f *fileStorage) GetExelFileData(hash string) (data []byte, err error) {
	dataKey := "data_" + hash
	data, err = f.Get(dataKey)
	if err != nil {
		return nil, err
	}
	return data, nil
}
func (f *fileStorage) CreateTempUser() (token string) {
	token = uuid.New().String()
	userKey := "user_" + token
	err := f.Set(userKey, []byte(token))
	if err != nil {
		return ""
	}
	return token
}
func (f *fileStorage) IsTempUserExist(token string) bool {
	userKey := "user_" + token
	user, err := f.Get(userKey)
	if err != nil {
		return false
	}
	if string(user) != token {
		return false
	}
	return true
}

// SaveDocFile сохраняет doc файл и возвращает его хеш
func (f *fileStorage) SaveDocFile(path string, data []byte) (string, error) {
	h := sha256.New()
	h.Write(data)
	bs := h.Sum(nil)
	hash := hex.EncodeToString(bs)

	// Сохраняем data в BadgerDB с ключом doc_data_hash
	dataKey := "doc_data_" + hash
	err := f.Set(dataKey, data)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// GetDocFileData получает данные doc файла по хешу
func (f *fileStorage) GetDocFileData(hash string) ([]byte, error) {
	dataKey := "doc_data_" + hash
	data, err := f.Get(dataKey)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// GetDocFileInfo получает информацию о doc файле по хешу
func (f *fileStorage) GetDocFileInfo(hash string) (models.FileInfo, error) {
	// Для doc файлов просто возвращаем базовую информацию
	dataKey := "doc_data_" + hash
	dataBytes, err := f.Get(dataKey)
	if err != nil {
		return models.FileInfo{}, err
	}

	fileInfo := models.FileInfo{
		Sheets:   []models.Sheet{}, // doc файлы не имеют листов
		FileName: "document.docx",  // базовое имя
		Size:     len(dataBytes),
	}

	return fileInfo, nil
}

// SaveDoc сохраняет документ и возвращает его хеш
func (f *fileStorage) SaveDoc(docBytes []byte) (string, error) {
	h := sha256.New()
	h.Write(docBytes)
	bs := h.Sum(nil)
	hash := hex.EncodeToString(bs)

	// Сохраняем data в BadgerDB с ключом generated_doc_hash
	dataKey := "generated_doc_" + hash
	err := f.Set(dataKey, docBytes)
	if err != nil {
		return "", err
	}

	return hash, nil
}
