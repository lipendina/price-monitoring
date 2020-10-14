package storage

import (
	"avito/internal"
	"avito/internal/db"
	"avito/internal/errors"
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"strings"
)

type ConfirmationStorageAPI interface {
	GetConfirmations(ctx context.Context, IDs []uuid.UUID) ([]*internal.Confirmation, error)
	UpdateConfirmation(ctx context.Context, ID uuid.UUID) error
	AddConfirmation(ctx context.Context, tx pgx.Tx, receiver string, adID uuid.UUID) (uuid.UUID, error)
	GetReceiverEmailWithAdvertNameById(ctx context.Context, ID uuid.UUID) (*internal.ReceiverEmailWithAdvertName, error)
	GetConfirmation(ctx context.Context, receiver string, link string) (uuid.UUID, int, error)
	UpdateConfirmationStatus(ctx context.Context, id uuid.UUID) error
	GetReceivers(ctx context.Context, advertID uuid.UUID) ([]string, error)
}

type confirmationStorage struct {
	db *db.ConnDB
}

func NewConfirmationStorageAPI(connDB *db.ConnDB) ConfirmationStorageAPI {
	return &confirmationStorage{
		db: connDB,
	}
}

func (c *confirmationStorage) GetReceiverEmailWithAdvertNameById(ctx context.Context, ID uuid.UUID) (*internal.ReceiverEmailWithAdvertName, error) {
	row, err := c.db.DB.Query(ctx, "select c.email, a.name from confirmation c inner join advert a on c.advert_id=a.id where c.id=$1", ID)
	if err != nil {
		return nil, err
	}

	result := &internal.ReceiverEmailWithAdvertName{}
	if row.Next() {
		err = row.Scan(&result.Email, &result.Name)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.NotFound
	}
	return result, nil
}

func (c *confirmationStorage) GetConfirmations(ctx context.Context, IDs []uuid.UUID) ([]*internal.Confirmation, error) {
	valueString := make([]string, 0, len(IDs))
	valueArgs := make([]interface{}, 0, len(IDs))
	for idx, id := range IDs {
		valueString = append(valueString, fmt.Sprintf("$%d", idx+1))
		valueArgs = append(valueArgs, id)
	}

	rows, err := c.db.DB.Query(ctx, fmt.Sprintf("select id, email, advert_id, created_at, is_confirm from confirmation where id in (%s)", strings.Join(valueString, ",")), valueArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*internal.Confirmation, 0, len(IDs))
	for rows.Next() {
		confirmation := &internal.Confirmation{}
		err = rows.Scan(&confirmation.ID, &confirmation.Receiver, &confirmation.AdvertID, &confirmation.CreatedAt, &confirmation.IsConfirm)
		if err != nil {
			return nil, err
		}
		result = append(result, confirmation)
	}

	return result, nil
}

func (c *confirmationStorage) UpdateConfirmation(ctx context.Context, ID uuid.UUID) error {
	rows, err := c.db.DB.Query(ctx, "update confirmation set is_confirm=true where id=$1", ID)
	if err != nil {
		return err
	}
	rows.Close()

	return nil
}

func (c *confirmationStorage) AddConfirmation(ctx context.Context, tx pgx.Tx, receiver string, adID uuid.UUID) (uuid.UUID, error) {
	ID := uuid.Must(uuid.NewV4())
	rows, err := tx.Query(ctx, "insert into confirmation (id, email, advert_id) values ($1, $2, $3)", ID, receiver, adID)
	if err != nil {
		return uuid.Nil, err
	}
	defer rows.Close()

	return ID, nil
}

func (c *confirmationStorage) GetConfirmation(ctx context.Context, receiver string, link string) (uuid.UUID, int, error) {
	id := uuid.Nil
	rows, err := c.db.DB.Query(ctx, "select c.id from confirmation c inner join advert a on c.advert_id=a.id where c.email=$1 and a.link=$2 and a.is_removed=false and c.is_confirm=true", receiver, link)
	if err != nil {
		return uuid.Nil, 0, err
	}

	result := 0
	for rows.Next() {
		rows.Scan(&id)
		result++
	}

	return id, result, nil
}

func (c *confirmationStorage) UpdateConfirmationStatus(ctx context.Context, id uuid.UUID) error {
	rows, err := c.db.DB.Query(ctx, "update confirmation set is_confirm=false where id=$1", id)
	if err != nil {
		return err
	}
	rows.Close()

	return nil
}

func (c *confirmationStorage) GetReceivers(ctx context.Context, advertID uuid.UUID) ([]string, error) {
	rows, err := c.db.DB.Query(ctx, "select email from confirmation where advert_id=$1", advertID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	emails := make([]string, 0)
	for rows.Next() {
		email := ""
		rows.Scan(&email)
		emails = append(emails, email)
	}

	return emails, nil
}
