package storage

import (
	"avito/internal/db"
	"context"
	"github.com/jackc/pgx/v4"
)

type StorageAPI interface {
	GetAdsStorage() AdsStorageAPI
	GetConfirmationStorage() ConfirmationStorageAPI
	GetTransaction(ctx context.Context) (pgx.Tx, error)
}

type storageAPI struct {
	adsStorage          AdsStorageAPI
	confirmationStorage ConfirmationStorageAPI
	connDB              *db.ConnDB
}

func (s *storageAPI) GetTransaction(ctx context.Context) (pgx.Tx, error) {
	tx, err := s.connDB.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *storageAPI) GetAdsStorage() AdsStorageAPI {
	return s.adsStorage
}

func (s *storageAPI) GetConfirmationStorage() ConfirmationStorageAPI {
	return s.confirmationStorage
}

func NewStorageAPI(connDB *db.ConnDB) StorageAPI {
	return &storageAPI{
		adsStorage:          NewAdsStorageAPI(connDB),
		confirmationStorage: NewConfirmationStorageAPI(connDB),
		connDB:              connDB,
	}
}
