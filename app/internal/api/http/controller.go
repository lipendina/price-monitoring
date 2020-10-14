package http

import (
	"avito/internal"
	"avito/internal/errors"
	"avito/internal/service"
	"avito/internal/utils"
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"golang.org/x/xerrors"
	"log"
	"net/http"
)

type Controller struct {
	logger  *log.Logger
	service service.ServiceAPI
}

func NewController(service service.ServiceAPI, logger *log.Logger) *Controller {
	return &Controller{
		service: service,
		logger:  logger,
	}
}

// здесь можно увеличить надежность доставки сообщений, если использовать брокер
func (c *Controller) PriceChangeSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var subscriptionRequest internal.SubscriptionRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&subscriptionRequest)
	if err != nil {
		c.logger.Printf("Error while parse request, reason: %+v\n", err)
		c.writeError(w, err)
		return
	}

	ad, err := utils.ParseAd(subscriptionRequest.Link)
	if err != nil {
		c.logger.Printf("Error while parse ad, reason: %+v\n", err)
		c.writeError(w, err)
		return
	}

	confirmationID, err := c.service.GetAdsService().PriceChangeSubscription(ctx, ad, subscriptionRequest.Receiver)
	if err != nil {
		c.logger.Printf("Error while subscript, reason: %+v\n", err)
		c.writeError(w, err)
		return
	}

	err = c.service.GetEmailService().SendEmailConfirmation(subscriptionRequest.Receiver, ad.Name, confirmationID)
	if err != nil {
		c.logger.Printf("Error while send confirmation mail, reason: %+v\n", err)
		c.writeError(w, err)
		return
	}

	c.writeResponse(w, []byte("Confirmation link send on your e-mail\n"))
}

// здесь можно увеличить надежность доставки сообщений, если использовать брокер
func (c *Controller) ConfirmationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	ID := r.URL.Query().Get("id")
	if len(ID) == 0 {
		c.logger.Printf("Error while parse value of subscription id\n")
		c.writeError(w, errors.NotFound)
		return
	}

	subscriptionID, err := uuid.FromString(ID)
	if err != nil {
		c.logger.Printf("Error while convert subscriptionID from string to uuid.UUID, reason: %+v\n", err)
		c.writeError(w, err)
		return
	}

	confirmationService := c.service.GetConfirmationService()
	err = confirmationService.AcceptConfirmation(ctx, subscriptionID)
	if err != nil {
		c.logger.Printf("Error while confirm email, reason: %+v\n", err)
		c.writeError(w, err)
		return
	}

	receiver, err := confirmationService.GetReceiverEmailWithAdvertNameById(ctx, subscriptionID)
	if err != nil {
		c.logger.Printf("Error while get receiver email and advert name, reason: %+v\n", err)
		c.writeError(w, err)
		return
	}

	err = c.service.GetEmailService().SendEmailConfirmed(receiver.Email, receiver.Name)
	if err != nil {
		c.logger.Printf("Error while send mail about confirmed subscription, reason: %+v\n", err)
		c.writeError(w, err)
		return
	}
	c.writeResponse(w, []byte("Subscription confirmed!\n"))
}

// здесь можно увеличить надежность доставки сообщений, если использовать брокер
func (c *Controller) UnsubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var subscriptionRequest internal.SubscriptionRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&subscriptionRequest)
	if err != nil {
		c.logger.Printf("Error while parse request, reason: %+v\n", err)
		c.writeError(w, err)
		return
	}

	err = c.service.GetConfirmationService().Unsubscription(ctx, subscriptionRequest.Receiver, subscriptionRequest.Link)
	if err != nil {
		c.logger.Printf("Error while unsubscript, reason: %+v\n", err)
		c.writeError(w, err)
		return
	}

	err = c.service.GetEmailService().SendEmailUnsubscription(subscriptionRequest.Receiver, subscriptionRequest.Link)
	if err != nil {
		c.logger.Printf("Error while send mail about unsubscription, reason: %+v\n", err)
		c.writeError(w, err)
		return
	}

	c.writeResponse(w, []byte("Confirmation canceled!\n"))
}

func (c *Controller) writeResponse(writer http.ResponseWriter, res []byte) {
	if _, err := writer.Write(res); err != nil {
		status := http.StatusInternalServerError
		c.logger.Printf("Failed to write response, reason: %s\n", err)
		http.Error(writer, http.StatusText(status), status)
	}
}

func (c *Controller) writeError(writer http.ResponseWriter, err error) {
	var status int
	switch {
	case xerrors.Is(err, errors.AlreadyExist):
		status = http.StatusConflict
	case xerrors.Is(err, errors.NotFound):
		status = http.StatusNotFound
	default:
		status = http.StatusInternalServerError
	}
	http.Error(writer, http.StatusText(status), status)
}
