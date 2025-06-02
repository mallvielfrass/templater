package filestorage

import "github.com/dgraph-io/badger/v4"

type fileStorage struct {
	bg *badger.DB
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
