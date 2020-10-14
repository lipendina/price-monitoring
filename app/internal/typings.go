package internal

import (
	"github.com/gofrs/uuid"
	"time"
)

type Ad struct {
	ID        uuid.UUID
	Link      string
	Name      string
	Price     int
	Removed   bool
	LastCheck time.Time
}

type Receiver struct {
	ID    uuid.UUID
	Email string
}

type Confirmation struct {
	ID        uuid.UUID
	Receiver  string
	AdvertID  uuid.UUID
	CreatedAt time.Time
	IsConfirm bool
}

type ReceiverEmailWithAdvertName struct {
	Email string
	Name  string
}

type SubscriptionRequest struct {
	Link     string `json:"link"`
	Receiver string `json:"receiver"`
}

type Subscription struct {
	Email string
	Link  string
	Name  string
	Price int
}
