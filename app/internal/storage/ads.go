package storage

import (
	"avito/internal"
	"avito/internal/db"
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"strings"
)

type AdsStorageAPI interface {
	AddNewAdvert(ctx context.Context, tx pgx.Tx, link string, name string, price int) (uuid.UUID, error)
	IsRecordExist(ctx context.Context, link string, receiver string) (bool, error)
	UpdatePrice(ctx context.Context, advertID uuid.UUID, newPrice int) error
	UpdateIsRemoved(ctx context.Context, advertID uuid.UUID) error
	GetAds(ctx context.Context, limit int) ([]*internal.Ad, error)
	UpdateLastChecked(ctx context.Context, channelUpdated chan uuid.UUID) error
}

type adsStorage struct {
	db *db.ConnDB
}

func NewAdsStorageAPI(connDB *db.ConnDB) AdsStorageAPI {
	return &adsStorage{
		db: connDB,
	}
}

func (a *adsStorage) AddNewAdvert(ctx context.Context, tx pgx.Tx, link string, name string, price int) (uuid.UUID, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return uuid.Nil, err
	}

	err = tx.QueryRow(ctx, "insert into advert (id, link, name, price) values ($1, $2, $3, $4) on conflict (link) where is_removed=false do update set link=$5 returning id", id, link, name, price, link).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (a *adsStorage) IsRecordExist(ctx context.Context, link string, receiver string) (bool, error) {
	var result int
	err := a.db.DB.QueryRow(ctx, "select count(*) from confirmation c inner join advert a on c.advert_id=a.id where a.link=$1 and c.email=$2 and a.is_removed=false", link, receiver).Scan(&result)
	if err != nil {
		return false, err
	}

	return result == 1, nil

}

func (a *adsStorage) UpdatePrice(ctx context.Context, advertID uuid.UUID, newPrice int) error {
	_, err := a.db.DB.Exec(ctx, "update advert set price=$1 where id=$2", newPrice, advertID)
	if err != nil {
		return err
	}

	return nil
}

func (a *adsStorage) UpdateIsRemoved(ctx context.Context, advertID uuid.UUID) error {
	_, err := a.db.DB.Exec(ctx, "update advert set is_removed=true where id=$1", advertID)
	if err != nil {
		return err
	}

	return nil
}

func (a *adsStorage) GetAds(ctx context.Context, limit int) ([]*internal.Ad, error) {
	rows, err := a.db.DB.Query(ctx, "select id, link, name, price from advert where is_removed=false order by last_check limit $1", limit)
	if err != nil {
		return nil, err
	}

	ads := make([]*internal.Ad, 0, limit)
	for rows.Next() {
		ad := &internal.Ad{}
		err = rows.Scan(&ad.ID, &ad.Link, &ad.Name, &ad.Price)
		ads = append(ads, ad)
	}

	return ads, nil
}

func (a *adsStorage) UpdateLastChecked(ctx context.Context, channelUpdated chan uuid.UUID) error {
	valueString := make([]string, 0)
	valueArgs := make([]interface{}, 0)
	idx := 1
	for id := range channelUpdated {
		valueString = append(valueString, fmt.Sprintf("$%d", idx))
		valueArgs = append(valueArgs, id)
		idx++
	}

	rows, err := a.db.DB.Query(ctx, fmt.Sprintf("update advert set last_check=now() where id in(%s)", strings.Join(valueString, ",")), valueArgs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return nil
}
