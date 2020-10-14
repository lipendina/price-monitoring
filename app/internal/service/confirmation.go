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

type ConfirmationServiceAPI interface {
	AcceptConfirmation(ctx context.Context, subscriptionID uuid.UUID) error
	GetReceiverEmailWithAdvertNameById(ctx context.Context, ID uuid.UUID) (*internal.ReceiverEmailWithAdvertName, error)
	Unsubscription(ctx context.Context, receiver string, link string) error
	GetReceivers(ctx context.Context, advertID uuid.UUID) ([]string, error)
}

type confirmationService struct {
	storage storage.StorageAPI
	ctx     context.Context
	log     *log.Logger
}

func NewConfirmationServiceAPI(api storage.StorageAPI) ConfirmationServiceAPI {
	return &confirmationService{
		storage: api,
		ctx:     context.Background(),
		log:     log.New(os.Stdout, "CONFIRMATION-SERVICE: ", log.LstdFlags),
	}
}

func (c *confirmationService) GetReceiverEmailWithAdvertNameById(ctx context.Context, ID uuid.UUID) (*internal.ReceiverEmailWithAdvertName, error) {
	result, err := c.storage.GetConfirmationStorage().GetReceiverEmailWithAdvertNameById(ctx, ID)
	if err != nil {
		c.log.Printf("Error while get receiver email and advert name from DB, reason: %+v\n", err)
		return nil, errors.InternalError
	}

	return result, nil
}

func (c *confirmationService) AcceptConfirmation(ctx context.Context, subscriptionID uuid.UUID) error {
	confirmationStorage := c.storage.GetConfirmationStorage()

	confirmations, err := confirmationStorage.GetConfirmations(ctx, []uuid.UUID{subscriptionID})
	if err != nil {
		c.log.Printf("Error while get subscription ID from DB, reason: %+v\n", err)
		return errors.InternalError
	}

	if len(confirmations) == 0 {
		c.log.Printf("Not found confirmation by id %s\n", subscriptionID)
		return errors.NotFound
	}

	if confirmations[0].IsConfirm == true {
		c.log.Printf("Confirmation with id %s already confirmed", subscriptionID)
		return errors.NotFound
	}

	if err = confirmationStorage.UpdateConfirmation(ctx, subscriptionID); err != nil {
		c.log.Printf("Error while update confirmation status in DB, reason: %+v\n", err)
		return errors.InternalError
	}
	return nil
}

func (c *confirmationService) Unsubscription(ctx context.Context, receiver string, link string) error {
	confirmationStorage := c.storage.GetConfirmationStorage()
	confirmationID, count, err := confirmationStorage.GetConfirmation(ctx, receiver, link)
	if err != nil {
		c.log.Printf("Error while get confirmations from DB, reason: %+v\n", err)
		return errors.InternalError
	}

	if count < 1 {
		return errors.NotFound
	}

	err = confirmationStorage.UpdateConfirmationStatus(ctx, confirmationID)
	if err != nil {
		c.log.Printf("Error while update confirmation in DB, reason: %+v\n", err)
		return errors.InternalError
	}

	return nil
}

func (c *confirmationService) GetReceivers(ctx context.Context, advertID uuid.UUID) ([]string, error) {
	emails, err := c.storage.GetConfirmationStorage().GetReceivers(ctx, advertID)
	if err != nil {
		c.log.Printf("Error while get receivers from DB, reason: %+v\n")
		return nil, errors.InternalError
	}
	return emails, err
}
