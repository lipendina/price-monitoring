package service

import (
	"avito/internal"
	"avito/internal/errors"
	"avito/internal/storage"
	"context"
	"github.com/gofrs/uuid"
	"log"
	"os"
)

type AdsServiceAPI interface {
	PriceChangeSubscription(ctx context.Context, ad *internal.Ad, receiver string) (uuid.UUID, error)
	UpdateAdPrice(ctx context.Context, advertID uuid.UUID, price int) error
	UpdateAdIsRemoved(ctx context.Context, advertID uuid.UUID) error
	GetAds(ctx context.Context, limit int) ([]*internal.Ad, error)
	UpdateLastChecked(ctx context.Context, channelUpdated chan uuid.UUID) error
}

type adsService struct {
	storage storage.StorageAPI
	ctx     context.Context
	log     *log.Logger
}

func NewAdsServiceAPI(api storage.StorageAPI) AdsServiceAPI {
	return &adsService{
		storage: api,
		ctx:     context.Background(),
		log:     log.New(os.Stdout, "ADS-SERVICE: ", log.LstdFlags),
	}
}

func (s *adsService) PriceChangeSubscription(ctx context.Context, ad *internal.Ad, receiver string) (uuid.UUID, error) {
	isExist, err := s.storage.GetAdsStorage().IsRecordExist(ctx, ad.Link, receiver)
	if err != nil {
		s.log.Printf("Error while count ads in DB, reason: %+v\n", err)
		return uuid.Nil, errors.InternalError
	}

	if isExist {
		return uuid.Nil, errors.AlreadyExist
	}

	tx, err := s.storage.GetTransaction(ctx)
	if err != nil {
		s.log.Printf("Error while get transaction, reason: %+v\n", err)
		return uuid.Nil, errors.InternalError
	}

	adID, err := s.storage.GetAdsStorage().AddNewAdvert(ctx, tx, ad.Link, ad.Name, ad.Price)
	if err != nil {
		tx.Rollback(ctx)
		s.log.Printf("Error while add new advert in DB, reason: %+v\n", err)
		return uuid.Nil, errors.InternalError
	}

	confirmationID, err := s.storage.GetConfirmationStorage().AddConfirmation(ctx, tx, receiver, adID)
	if err != nil {
		tx.Rollback(ctx)
		s.log.Printf("Error while add confirmation in DB, reason: %+v\n", err)
		return uuid.Nil, errors.InternalError
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Printf("Error while commit transaction, reason: %+v\n", err)
		return uuid.Nil, errors.InternalError
	}

	return confirmationID, nil
}

func (s *adsService) UpdateAdPrice(ctx context.Context, advertID uuid.UUID, price int) error {
	err := s.storage.GetAdsStorage().UpdatePrice(ctx, advertID, price)
	if err != nil {
		s.log.Printf("Error while update price in DB, reason: %+v\n", err)
		return err
	}

	return nil
}

func (s *adsService) UpdateAdIsRemoved(ctx context.Context, advertID uuid.UUID) error {
	err := s.storage.GetAdsStorage().UpdateIsRemoved(ctx, advertID)
	if err != nil {
		s.log.Printf("Error while update is_removed in DB, reason: %+v\n", err)
	}

	return nil
}

func (s *adsService) GetAds(ctx context.Context, limit int) ([]*internal.Ad, error) {
	ads, err := s.storage.GetAdsStorage().GetAds(ctx, limit)
	if err != nil {
		s.log.Printf("Error while get ads from DB, reason: %+v\n", err)
		return nil, err
	}

	return ads, nil
}

func (s *adsService) UpdateLastChecked(ctx context.Context, channelUpdated chan uuid.UUID) error {
	err := s.storage.GetAdsStorage().UpdateLastChecked(ctx, channelUpdated)
	if err != nil {
		s.log.Printf("Error while update last_checked in advert, reason: %+v\n", err)
		return err
	}

	return nil
}
